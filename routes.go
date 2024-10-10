package main

import (
	"github.com/gin-gonic/gin"
)

// Routes sets up the API routes.
func Routes() {
	router := gin.Default()
	router.POST("/api/wallet", CreateWallet)
	router.GET("/api/wallets", ListWallets)
	router.GET("/api/wallet/:address", GetWallet)
	router.POST("/api/sign-transaction", SignAndSendTransaction)

	router.Run("0.0.0.0:8085")
}
