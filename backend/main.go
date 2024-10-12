package main

import (
	"context"
	"log"

	"github.com/natneam/crypto-wallet-app/internal/config"
	"github.com/natneam/crypto-wallet-app/internal/db"
	"github.com/natneam/crypto-wallet-app/internal/handlers"
	"github.com/natneam/crypto-wallet-app/internal/kms"
	"github.com/natneam/crypto-wallet-app/internal/middlewares"
	"github.com/natneam/crypto-wallet-app/internal/repositories"
	"github.com/natneam/crypto-wallet-app/internal/services"
	"github.com/natneam/crypto-wallet-app/internal/web3"

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

	// Initialize database
	err = db.InitDatabase(dbClient)
	if err != nil {
		return nil, err
	}

	// Initialize repository
	repository := repositories.NewRepository(dbClient)

	// Initialize services
	service := services.NewService(repository, web3Client, kmsClient)

	// Set up the router
	router := gin.Default()
	router.Use(middlewares.CORSMiddleware())

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
