package main

import (
	"crypto-wallet-app/internal/config"
	"crypto-wallet-app/internal/db"
	"crypto-wallet-app/internal/handlers"
	"crypto-wallet-app/internal/utils"
	"crypto-wallet-app/internal/web3"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	config.Load()

	// Initialize database connection
	dbClient, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize web3 connection
	web3Client, err := web3.Connect()

	// Set up the router
	router := gin.Default()
	router.Use(utils.CORSMiddleware())

	// Setup routes with the handler
	handlers.RegisterRoutes(router, dbClient, web3Client)

	// Start the server
	if err := router.Run(":8085"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
