package main

import (
	"math/big"
)

// Helper function to convert wei to ether
func weiToEther(wei *big.Int) string {
	if wei == nil {
		return "0"
	}

	// Convert to ether (divide by 10^18)
	ether := new(big.Float).Quo(new(big.Float).SetInt(wei), new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	return ether.Text('f', 18)
}
