package utils

import (
	"math"
	"math/big"
)

// WeiToEther converts wei (smallest Ethereum unit) to ether
func WeiToEther(weiAmount *big.Int) float64 {
	etherValue := new(big.Float).SetInt(weiAmount)
	etherValue = etherValue.Quo(etherValue, big.NewFloat(math.Pow10(18)))
	value, _ := etherValue.Float64()
	return value
}
