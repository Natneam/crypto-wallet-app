package main

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

var web3Client *ethclient.Client

func connectEth() error {
	var err error
	web3Client, err = ethclient.Dial("https://eth-sepolia.g.alchemy.com/v2/UduxQFsw3MQ2ITMhVgk2hNXqoGyvoiHV")
	if err != nil {
		return fmt.Errorf("failed to connect to Sepolia: %v", err)
	}

	blockNumber, err := web3Client.BlockNumber(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get block number: %v", err)
	}

	fmt.Printf("Connected to Sepolia! Current block number: %v\n", blockNumber)
	return nil
}
