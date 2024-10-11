package utils

import (
	"context"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethereumTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// WeiToEther converts wei (smallest Ethereum unit) to ether
func WeiToEther(weiAmount *big.Int) float64 {
	etherValue := new(big.Float).SetInt(weiAmount)
	etherValue = etherValue.Quo(etherValue, big.NewFloat(math.Pow10(18)))
	value, _ := etherValue.Float64()
	return value
}

func WaitForTransactionReceipt(ctx context.Context, client *ethclient.Client, txHash common.Hash) (*ethereumTypes.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, err
		}
		// Transaction not yet mined, wait and retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second * 5):
			// Continue waiting
		}
	}
}
