package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/natneam/crypto-wallet-app/backend/internal/middlewares"
	"github.com/natneam/crypto-wallet-app/backend/internal/models"
	"github.com/natneam/crypto-wallet-app/backend/internal/services"

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

	// Public routes
	r.POST("/api/signup", handler.SignUp)
	r.POST("/api/login", handler.Login)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware(service))
	protected.GET("/wallets", handler.ListWallets)
	protected.GET("/wallet/:address", handler.GetWallet)
	protected.POST("/sign-transaction", handler.SignAndSendTransaction)
	protected.POST("/wallet", handler.CreateWallet)
}

// creates a new wallet and stores it in the database
func (h *Handler) CreateWallet(c *gin.Context) {

	var jsonData map[string]string

	if err := c.ShouldBindJSON(&jsonData); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	walletName, exists := jsonData["name"]
	userId, _ := c.Get("user_id")
	if !exists {
		c.JSON(400, gin.H{"error": "Missing wallet name"})
		return
	}

	var wallet models.Wallet
	var err error

	wallet, err = h.service.CreateWallet(walletName, userId.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// retrieves a wallet by its address
func (h *Handler) GetWallet(c *gin.Context) {
	walletAddress := c.Param("address")
	userID, _ := c.Get("user_id")

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

	wallet, err := h.service.GetWallet(walletAddress, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// lists all wallets
func (h *Handler) ListWallets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	wallets, err := h.service.ListWallets(userID.(string))
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
	value := transaction.Value
	userID, _ := c.Get("user_id")

	result, err := h.service.SignAndSendTransaction(fromAddress, toAddress, value, userID.(string))

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) SignUp(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SignUp(input.Username, input.Email, input.Password); err != nil {
		fmt.Println(`Error: `, err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (h *Handler) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Login(input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
