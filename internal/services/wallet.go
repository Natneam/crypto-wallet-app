package services

import (
	"context"
	"crypto-wallet-app/internal/models"
	"crypto-wallet-app/internal/utils"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateWallet(dbClient *mongo.Client, web3Client *ethclient.Client, walletName string) (models.Wallet, error) {
	var newWallet models.Wallet

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return newWallet, err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	publicKeyHex := address.Hex()

	// Check the balance on Sepolia
	balance, err := web3Client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return newWallet, err
	}

	newWallet = models.Wallet{
		Name:       walletName,
		PrivateKey: privateKeyHex,
		PublicKey:  publicKeyHex,
		Balance:    balance.String(),
	}

	collection := dbClient.Database("walletdb").Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, newWallet)

	newWallet.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return newWallet, err
}

func ListWallets(dbClient *mongo.Client, web3Client *ethclient.Client) ([]models.Wallet, error) {
	collection := dbClient.Database("walletdb").Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var wallets []models.Wallet
	if err := cursor.All(ctx, &wallets); err != nil {
		return nil, err
	}

	for i, wallet := range wallets {
		// Convert the public key to an Ethereum address
		address := common.HexToAddress(wallet.PublicKey)

		// Fetch balance from Sepolia
		balance, err := web3Client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			return nil, err
		}

		// Convert balance from wei to ether and update the wallet
		wallets[i].Balance = fmt.Sprintf("%f ETH", utils.WeiToEther(balance))
	}

	return wallets, nil
}

func GetWallet(dbClient *mongo.Client, web3Client *ethclient.Client, address string) (*models.Wallet, error) {
	collection := dbClient.Database("walletdb").Collection("wallets")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wallet models.Wallet
	if err := collection.FindOne(ctx, bson.M{"publickey": address}).Decode(&wallet); err != nil {
		return nil, err
	}

	// Convert the wallet address string to a common.Address type
	account := common.HexToAddress(address)

	// Get the balance of the wallet
	balance, err := web3Client.BalanceAt(ctx, account, nil)
	if err != nil {
		return nil, err
	}

	wallet.Balance = fmt.Sprintf("%f ETH", utils.WeiToEther(balance))

	return &wallet, nil
}

func SignAndSendTransaction(dbClient *mongo.Client, web3Client *ethclient.Client, fromAddress common.Address, toAddress common.Address, privateKey *ecdsa.PrivateKey) (models.TransactionResult, error) {
	collection := dbClient.Database("walletdb").Collection("transactions")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var transaction models.TransactionResult

	// Get the nonce for the from address
	nonce, err := web3Client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return transaction, err
	}

	// Get the current gas price
	gasPrice, err := web3Client.SuggestGasPrice(ctx)
	if err != nil {
		return transaction, err
	}

	// Create the transaction
	tx := types.NewTransaction(nonce, toAddress, big.NewInt(0), 21000, gasPrice, nil)

	// Get the chain ID
	chainID, err := web3Client.NetworkID(ctx)
	if err != nil {
		return transaction, err
	}

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return transaction, err
	}

	// Send the transaction
	err = web3Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return transaction, err
	}

	transaction = models.TransactionResult{
		TransactionHash: signedTx.Hash().Hex(),
		From:            fromAddress.Hex(),
		To:              toAddress.Hex(),
		GasPrice:        gasPrice.String(),
		Value:           "0",
	}

	_, err = collection.InsertOne(ctx, transaction)

	return transaction, err
}
