package handlers

import (
	"crypto-wallet-app/internal/models"
	"crypto-wallet-app/internal/services"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *services.Service
}

func NewHandler(r *gin.Engine, service *services.Service) {
	handler := &Handler{
		service: service,
	}

	r.POST("/api/wallet", handler.CreateWallet)
	r.GET("/api/wallets", handler.ListWallets)
	r.GET("/api/wallet/:address", handler.GetWallet)
	r.POST("/api/sign-transaction", handler.SignAndSendTransaction)
}

// creates a new wallet and stores it in the database
func (h *Handler) CreateWallet(c *gin.Context) {

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

	wallet, err = h.service.CreateWallet(walletName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// retrieves a wallet by its address
func (h *Handler) GetWallet(c *gin.Context) {
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

	wallet, err := h.service.GetWallet(walletAddress)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// lists all wallets
func (h *Handler) ListWallets(c *gin.Context) {
	wallets, err := h.service.ListWallets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list wallets"})
		return
	}

	c.JSON(http.StatusOK, wallets)
}

// signs and sends a transaction
func (h *Handler) SignAndSendTransaction(c *gin.Context) {
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

	result, err := h.service.SignAndSendTransaction(fromAddress, toAddress, kmsKeyID, value)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign and send transaction"})
		return
	}

	c.JSON(http.StatusOK, result)
}
