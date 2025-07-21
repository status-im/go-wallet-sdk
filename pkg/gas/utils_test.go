package gas

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPercentile_PercentileBoundaries(t *testing.T) {
	data := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5)}

	// Test boundary conditions
	assert.Equal(t, big.NewInt(1), getPercentile(data, 0))   // 0th percentile
	assert.Equal(t, big.NewInt(1), getPercentile(data, 1))   // 1st percentile
	assert.Equal(t, big.NewInt(2), getPercentile(data, 25))  // 25th percentile
	assert.Equal(t, big.NewInt(3), getPercentile(data, 50))  // 50th percentile
	assert.Equal(t, big.NewInt(4), getPercentile(data, 75))  // 75th percentile
	assert.Equal(t, big.NewInt(5), getPercentile(data, 99))  // 99th percentile
	assert.Equal(t, big.NewInt(5), getPercentile(data, 100)) // 100th percentile
}
