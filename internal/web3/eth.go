package web3

import (
	"context"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
)

func Connect() (*ethclient.Client, error) {
	web3Client, err := ethclient.Dial(os.Getenv("SEPOLIA_URL"))
	if err != nil {
		return nil, err
	}

	// Get the current block number to confirm connection
	blockNumber, err := web3Client.BlockNumber(context.Background())
	if err != nil {
		return nil, err
	}

	fmt.Printf("Connected to Sepolia! Current block number: %v\n", blockNumber)

	return web3Client, nil
}
