package main

import (
	"log"
)

func main() {
	// Load AWS region from environment variable

	// awsRegion := os.Getenv("AWS_REGION")
	// if awsRegion == "" {
	// 	log.Fatal("AWS_REGION environment variable not set")
	// }

	connectMongoDB()
	connectEth()
	Routes()
	log.Println("Server starting on :8085...")
}
