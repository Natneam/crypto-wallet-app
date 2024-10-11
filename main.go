package main

import (
	"context"
	"crypto-wallet-app/internal/config"
	"crypto-wallet-app/internal/db"
	"crypto-wallet-app/internal/handlers"
	"crypto-wallet-app/internal/kms"
	"crypto-wallet-app/internal/repositories"
	"crypto-wallet-app/internal/services"
	"crypto-wallet-app/internal/utils"
	"crypto-wallet-app/internal/web3"
	"log"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Router  *gin.Engine
	Cleanup func()
}

func NewApplication() (*Application, error) {
	// Load configuration
	config.Load()

	// Initialize database connection
	dbClient, err := db.Connect()
	if err != nil {
		return nil, err
	}

	// Initialize web3 connection
	web3Client, err := web3.Connect()
	if err != nil {
		return nil, err
	}

	// Initialize KMS client
	kmsClient, err := kms.NewKMSClient()
	if err != nil {
		return nil, err
	}

	// Initialize repository
	repository := repositories.NewRepository(dbClient)

	// Initialize services
	service := services.NewService(repository, web3Client, kmsClient)

	// Set up the router
	router := gin.Default()
	router.Use(utils.CORSMiddleware())

	// Initialize handlers
	handlers.NewHandler(router, service)

	return &Application{
		Router: router,
		Cleanup: func() {
			dbClient.Disconnect(context.Background())
			web3Client.Close()
		},
	}, nil
}

func main() {
	app, err := NewApplication()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.Cleanup()

	// Start the Gin server
	if err := app.Router.Run(":8085"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
