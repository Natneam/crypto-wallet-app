package config

import (
	"log"

	"github.com/joho/godotenv"
)

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found. Using default or environment variables.")
	}
}
