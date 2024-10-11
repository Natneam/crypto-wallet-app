package handlers

import (
	"crypto-wallet-app/internal/models"
	"crypto-wallet-app/internal/services"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(router *gin.Engine, dbClient *mongo.Client, web3Client *ethclient.Client, kmsClient *kms.Client) {
	router.POST("/api/wallet", CreateWallet(dbClient, web3Client, kmsClient))
	router.GET("/api/wallets", ListWallets(dbClient, web3Client))
	router.GET("/api/wallet/:address", GetWallet(dbClient, web3Client, kmsClient))
	router.POST("/api/sign-transaction", SignAndSendTransaction(dbClient, web3Client, kmsClient))
}

// creates a new wallet and stores it in the database
func CreateWallet(dbClient *mongo.Client, web3Client *ethclient.Client, kmsClient *kms.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

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

		var wallet models.Wallet
		var err error

		wallet, err = services.CreateWallet(dbClient, web3Client, kmsClient, walletName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		c.JSON(http.StatusOK, wallet)
	}
}

// retrieves a wallet by its address
func GetWallet(dbClient *mongo.Client, web3Client *ethclient.Client, kmsClient *kms.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		wallet, err := services.GetWallet(dbClient, web3Client, kmsClient, walletAddress)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			return
		}

		c.JSON(http.StatusOK, wallet)
	}
}

// lists all wallets
func ListWallets(dbClient *mongo.Client, web3Client *ethclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		wallets, err := services.ListWallets(dbClient, web3Client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list wallets"})
			return
		}

		c.JSON(http.StatusOK, wallets)
	}
}

// signs and sends a transaction
func SignAndSendTransaction(dbClient *mongo.Client, web3Client *ethclient.Client, kmsClient *kms.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var transaction models.TransactionRequest
		var result models.TransactionResult

		if err := c.ShouldBindJSON(&transaction); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Validate addresses
		if !common.IsHexAddress(transaction.FromAddress) || !common.IsHexAddress(transaction.ToAddress) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address format"})
			return
		}

		fromAddress := common.HexToAddress(transaction.FromAddress)
		toAddress := common.HexToAddress(transaction.ToAddress)
		kmsKeyID := transaction.KMSKeyID
		value := transaction.Value

		result, err := services.SignAndSendTransaction(dbClient, web3Client, kmsClient, fromAddress, toAddress, kmsKeyID, value)

		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign and send transaction"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
