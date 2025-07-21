package ethclient

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockNumberConversion(t *testing.T) {
	// Test toBlockNumArg function with various inputs
	testCases := []struct {
		name     string
		input    *big.Int
		expected string
	}{
		{
			name:     "nil block number",
			input:    nil,
			expected: "latest",
		},
		{
			name:     "zero block number",
			input:    big.NewInt(0),
			expected: "0x0",
		},
		{
			name:     "positive block number",
			input:    big.NewInt(436),
			expected: "0x1b4",
		},
		{
			name:     "large block number",
			input:    big.NewInt(1000000),
			expected: "0xf4240",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toBlockNumArg(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
