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

func TestGetPercentile_TenValues(t *testing.T) {
	// Test with 10 values evenly distributed from 10 to 100
	data := []*big.Int{
		big.NewInt(10), big.NewInt(20), big.NewInt(30), big.NewInt(40), big.NewInt(50),
		big.NewInt(60), big.NewInt(70), big.NewInt(80), big.NewInt(90), big.NewInt(100),
	}

	assert.Equal(t, big.NewInt(10), getPercentile(data, 10))
	assert.Equal(t, big.NewInt(30), getPercentile(data, 25))
	assert.Equal(t, big.NewInt(50), getPercentile(data, 50))
	assert.Equal(t, big.NewInt(80), getPercentile(data, 75))
	assert.Equal(t, big.NewInt(90), getPercentile(data, 90))
	assert.Equal(t, big.NewInt(100), getPercentile(data, 95))
	assert.Equal(t, big.NewInt(100), getPercentile(data, 99))
	assert.Equal(t, big.NewInt(100), getPercentile(data, 100))

	// Edge cases
	assert.Equal(t, big.NewInt(10), getPercentile(data, 0)) // 0th percentile = minimum
	assert.Equal(t, big.NewInt(10), getPercentile(data, 5)) // 5th percentile
}
