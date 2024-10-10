package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateWallet(c *gin.Context) {
	// Check if we're connected to Sepolia
	if web3Client == nil {
		if err := connectEth(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to connect to Sepolia network"})
			return
		}
	}

	var jsonData map[string]string

	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	walletName, exists := jsonData["name"]
	if !exists {
		c.JSON(400, gin.H{"error": "Missing wallet name"})
		return
	}

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate private key"})
		return
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	publicKeyHex := address.Hex()

	// Check the balance on Sepolia
	balance, err := web3Client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch balance from Sepolia"})
		return
	}

	newWallet := Wallet{
		Name:       walletName,
		PrivateKey: privateKeyHex,
		PublicKey:  publicKeyHex,
		Balance:    balance.String(),
	}

	collection := dbClient.Database("walletdb").Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, newWallet)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to insert wallet"})
		return
	}

	newWallet.ID = result.InsertedID.(primitive.ObjectID).Hex()

	c.JSON(201, gin.H{"status": "wallet created", "wallet": newWallet})
}

func ListWallets(c *gin.Context) {
	collection := dbClient.Database("walletdb").Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// MongoDB query to find all wallets
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve wallets"})
		return
	}
	defer cursor.Close(ctx)

	var wallets []Wallet
	if err = cursor.All(ctx, &wallets); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse wallets"})
		return
	}

	// For each wallet, retrieve its balance from the Sepolia network
	for i, wallet := range wallets {
		// Convert the public key to an Ethereum address
		address := common.HexToAddress(wallet.PublicKey)

		// Fetch balance from Sepolia
		balance, err := web3Client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve balance for wallet", "wallet": wallet.PublicKey})
			return
		}

		// Convert balance from wei to ether and update the wallet
		wallets[i].Balance = fmt.Sprintf("%f ETH", WeiToEther(balance))
	}

	// Return the list of wallets with balances as JSON
	c.IndentedJSON(http.StatusOK, wallets)
}

// GetWalletHandler checks if a wallet exists in the specified Ethereum network
func GetWallet(c *gin.Context) {
	// Get the wallet address from the URL parameter
	walletAddress := c.Param("address")

	// Trim any whitespace and ensure the address starts with "0x"
	walletAddress = strings.TrimSpace(walletAddress)
	if !strings.HasPrefix(walletAddress, "0x") {
		walletAddress = "0x" + walletAddress
	}

	// Validate the wallet address
	if !common.IsHexAddress(walletAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet address", "address": walletAddress})
		return
	}

	// Convert the wallet address string to a common.Address type
	address := common.HexToAddress(walletAddress)

	// Check if we're connected to the Ethereum network
	if web3Client == nil {
		if err := connectEth(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Ethereum network"})
			return
		}
	}

	// Create a context with timeout for our Ethereum calls
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the address exists by getting its nonce
	nonce, err := web3Client.PendingNonceAt(ctx, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check address existence", "details": err.Error()})
		return
	}

	// Get the balance of the wallet
	balance, err := web3Client.BalanceAt(ctx, address, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wallet balance", "details": err.Error()})
		return
	}

	// Prepare the response
	response := gin.H{
		"address": address.Hex(),
		"exists":  true, // If we got here, the address exists on the blockchain
		"balance": fmt.Sprintf("%f ETH", WeiToEther(balance)),
		"nonce":   nonce,
	}

	c.JSON(http.StatusOK, response)
}

func SignAndSendTransaction(c *gin.Context) {
	var req SignTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate addresses
	if !common.IsHexAddress(req.FromAddress) || !common.IsHexAddress(req.ToAddress) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address format"})
		return
	}

	fromAddress := common.HexToAddress(req.FromAddress)
	toAddress := common.HexToAddress(req.ToAddress)

	// Parse private key
	privateKey, err := crypto.HexToECDSA(req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid private key"})
		return
	}

	// Ensure we're connected to the Ethereum network
	if web3Client == nil {
		if err := connectEth(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Ethereum network"})
			return
		}
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the nonce for the from address
	nonce, err := web3Client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve nonce"})
		return
	}

	// Get the current gas price
	gasPrice, err := web3Client.SuggestGasPrice(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve gas price"})
		return
	}

	// Create the transaction
	tx := types.NewTransaction(nonce, toAddress, big.NewInt(0), 21000, gasPrice, nil)

	// Get the chain ID
	chainID, err := web3Client.NetworkID(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chain ID"})
		return
	}

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign transaction"})
		return
	}

	// Send the transaction
	err = web3Client.SendTransaction(ctx, signedTx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send transaction", "details": err.Error()})
		return
	}

	// Return the transaction hash and details
	c.JSON(http.StatusOK, gin.H{
		"transactionHash": signedTx.Hash().Hex(),
		"from":            fromAddress.Hex(),
		"to":              toAddress.Hex(),
		"nonce":           nonce,
		"gasPrice":        gasPrice.String(),
		"value":           "0",
	})
}

// WeiToEther converts wei (smallest Ethereum unit) to ether
func WeiToEther(weiAmount *big.Int) float64 {
	etherValue := new(big.Float).SetInt(weiAmount)
	etherValue = etherValue.Quo(etherValue, big.NewFloat(math.Pow10(18)))
	value, _ := etherValue.Float64()
	return value
}
