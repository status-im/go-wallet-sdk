package gas

import (
	"math/big"
)

// BigIntToGweiFloat converts wei to gwei as a float64
func BigIntToGweiFloat(wei *big.Int) float64 {
	if wei == nil {
		return 0
	}

	// Convert wei to gwei (1 gwei = 10^9 wei)
	gwei := new(big.Float).Quo(
		new(big.Float).SetInt(wei),
		new(big.Float).SetInt(big.NewInt(1e9)),
	)

	result, _ := gwei.Float64()
	return result
}
