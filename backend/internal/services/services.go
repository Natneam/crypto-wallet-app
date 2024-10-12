package services

import (
	"context"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/natneam/crypto-wallet-app/backend/internal/models"
	"github.com/natneam/crypto-wallet-app/backend/internal/repositories"
	"github.com/natneam/crypto-wallet-app/backend/internal/utils"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmsTypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang-jwt/jwt/v4"
	ethawskmssigner "github.com/welthee/go-ethereum-aws-kms-tx-signer/v2"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo       *repositories.Repository
	web3Client *ethclient.Client
	kmsClient  *kms.Client
	jwtSecret  []byte
}

func NewService(repo *repositories.Repository, web3Client *ethclient.Client, kmsClient *kms.Client) *Service {
	return &Service{
		repo:       repo,
		web3Client: web3Client,
		kmsClient:  kmsClient,
	}
}

func (s *Service) CreateWallet(walletName string, userId string) (models.Wallet, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	var newWallet models.Wallet

	// Create a new KMS key
	createKeyInput := &kms.CreateKeyInput{
		Description: aws.String("Ethereum wallet key"),
		KeyUsage:    kmsTypes.KeyUsageTypeSignVerify,
		KeySpec:     kmsTypes.KeySpecEccSecgP256k1, // Ethereum uses secp256k1 curve
	}

	createKeyOutput, err := s.kmsClient.CreateKey(ctx, createKeyInput)
	if err != nil {
		return newWallet, fmt.Errorf("failed to create KMS key: %v", err)
	}

	// Get the public key
	getPublicKeyInput := &kms.GetPublicKeyInput{
		KeyId: createKeyOutput.KeyMetadata.KeyId,
	}

	getPublicKeyOutput, err := s.kmsClient.GetPublicKey(ctx, getPublicKeyInput)
	if err != nil {
		return newWallet, fmt.Errorf("failed to get public key: %v", err)
	}

	// Parse the ASN.1 DER encoded public key
	var publicKey struct {
		Algorithm struct {
			Algorithm  asn1.ObjectIdentifier
			Parameters asn1.RawValue
		}
		PublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(getPublicKeyOutput.PublicKey, &publicKey)
	if err != nil {
		return newWallet, fmt.Errorf("failed to parse public key: %v", err)
	}

	// Extract the actual public key bytes
	publicKeyBytes := publicKey.PublicKey.Bytes
	if len(publicKeyBytes) == 65 && publicKeyBytes[0] == 4 {
		// Remove the leading 0x04 byte (uncompressed point indicator)
		publicKeyBytes = publicKeyBytes[1:]
	} else if len(publicKeyBytes) != 64 {
		return newWallet, fmt.Errorf("unexpected public key length after parsing: %d", len(publicKeyBytes))
	}

	// Calculate the Ethereum address
	hash := crypto.Keccak256(publicKeyBytes)
	address := common.BytesToAddress(hash[12:])

	publicKeyHex := address.Hex()

	// Check the balance on Sepolia
	balance, err := s.web3Client.BalanceAt(ctx, address, nil)
	if err != nil {
		return newWallet, fmt.Errorf("failed to get balance: %v", err)
	}

	newWallet = models.Wallet{
		Name:      walletName,
		PublicKey: publicKeyHex,
		Balance:   balance.String(),
		KMSKeyID:  *createKeyOutput.KeyMetadata.KeyId,
		UserID:    userId,
	}

	// Save the wallet to the database
	newWallet, err = s.repo.SaveWallet(ctx, &newWallet)
	if err != nil {
		return newWallet, fmt.Errorf("failed to save wallet: %v", err)
	}
	return newWallet, nil
}

func (s *Service) ListWallets(userId string) ([]models.Wallet, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	wallets, err := s.repo.ListWallets(ctx, userId)
	if err != nil {
		return nil, err
	}

	// Fetch the balance for each wallet
	for i, wallet := range wallets {
		// Convert the public key to an Ethereum address
		address := common.HexToAddress(wallet.PublicKey)

		// Fetch balance from Sepolia
		balance, err := s.web3Client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			return nil, err
		}

		// Convert balance from wei to ether and update the wallet
		wallets[i].Balance = fmt.Sprintf("%f ETH", utils.WeiToEther(balance))
	}

	return wallets, nil
}

func (s *Service) GetWallet(address string, userId string) (*models.Wallet, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	wallet, err := s.repo.GetWallet(ctx, address, userId)
	if err != nil {
		return nil, err
	}

	// Convert the wallet address string to a common.Address type
	account := common.HexToAddress(address)

	// Get the balance of the wallet
	balance, err := s.web3Client.BalanceAt(ctx, account, nil)
	if err != nil {
		return nil, err
	}

	wallet.Balance = fmt.Sprintf("%f ETH", utils.WeiToEther(balance))

	return &wallet, nil
}

func (s *Service) SignAndSendTransaction(fromAddress common.Address, toAddress common.Address, value string, userId string) (models.TransactionResult, error) {

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	// Get user's wallet details
	wallet, err := s.repo.GetWallet(ctx, fromAddress.Hex(), userId)
	if err != nil {
		return models.TransactionResult{}, err
	}

	// Get the chain ID
	chainID, err := s.web3Client.NetworkID(ctx)
	if err != nil {
		return models.TransactionResult{}, err
	}

	// Create AWS KMS transactor
	transactOpts, err := ethawskmssigner.NewAwsKmsTransactorWithChainIDCtx(ctx, s.kmsClient, wallet.KMSKeyID, chainID)
	if err != nil {
		return models.TransactionResult{}, err
	}

	// Get the latest nonce for the fromAddress
	nonce, err := s.web3Client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return models.TransactionResult{}, err
	}

	// Get the current gas price
	gasPrice, err := s.web3Client.SuggestGasPrice(ctx)
	if err != nil {
		return models.TransactionResult{}, err
	}

	// Estimate gas limit
	gasLimit, err := s.web3Client.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddress,
		To:   &toAddress,
		Data: nil,
	})
	if err != nil {
		return models.TransactionResult{}, err
	}

	// Create the transaction
	val, _ := new(big.Int).SetString(value, 10)
	tx := ethereumTypes.NewTransaction(nonce, toAddress, val, gasLimit, gasPrice, nil)

	// Sign the transaction using AWS KMS
	signedTx, err := transactOpts.Signer(transactOpts.From, tx)
	if err != nil {
		return models.TransactionResult{}, err
	}

	// Send the transaction
	err = s.web3Client.SendTransaction(ctx, signedTx)
	if err != nil {
		return models.TransactionResult{}, err
	}

	// Wait for the transaction to be mined
	var receipt *ethereumTypes.Receipt

	for {
		receipt, err = s.web3Client.TransactionReceipt(ctx, signedTx.Hash())
		if err == nil {
			break
		}
		if err != ethereum.NotFound {
			return models.TransactionResult{}, err
		}
		// Transaction not yet mined, wait and retry
		select {
		case <-ctx.Done():
			return models.TransactionResult{}, ctx.Err()
		case <-time.After(time.Second * 5):
			// Continue waiting
		}
	}

	// Create the transaction result
	result := models.TransactionResult{
		TransactionHash: signedTx.Hash().Hex(),
		BlockNumber:     receipt.BlockNumber.Uint64(),
		GasUsed:         receipt.GasUsed,
		From:            fromAddress.Hex(),
		To:              toAddress.Hex(),
		GasPrice:        gasPrice.String(),
		Value:           value,
		UserID:          userId,
	}

	// Save the transaction result to the database
	savedTrx, err := s.repo.SaveTransaction(ctx, &result)
	if err != nil {
		return models.TransactionResult{}, err
	}

	return savedTrx, nil
}

func (s *Service) SignUp(username, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	return s.repo.CreateUser(user)
}

func (s *Service) Login(username, password string) (string, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			// If user_id is not a string, try to convert it to a string
			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				return "", fmt.Errorf("invalid user_id in token")
			}
			userID = fmt.Sprintf("%.0f", userIDFloat)
		}
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}
