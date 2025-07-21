package gas

import (
	"math"
	"math/big"
)

// getPercentile calculates the value at a given percentile from sorted data
func getPercentile(sortedData []*big.Int, percentile float64) *big.Int {
	if len(sortedData) == 0 {
		return big.NewInt(0)
	}

	index := (len(sortedData) * int(math.Ceil(percentile))) / 100
	if index >= len(sortedData) {
		index = len(sortedData) - 1
	}

	return new(big.Int).Set(sortedData[index])
}
