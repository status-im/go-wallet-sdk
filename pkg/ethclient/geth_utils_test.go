package ethclient

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

	// Test negative block numbers
	negativeTestCases := []struct {
		name     string
		input    *big.Int
		expected string
	}{
		{
			name:     "earliest block",
			input:    big.NewInt(-5),
			expected: "earliest",
		},
		{
			name:     "safe block",
			input:    big.NewInt(-4),
			expected: "safe",
		},
		{
			name:     "finalized block",
			input:    big.NewInt(-3),
			expected: "finalized",
		},
		{
			name:     "latest block",
			input:    big.NewInt(-2),
			expected: "latest",
		},
		{
			name:     "pending block",
			input:    big.NewInt(-1),
			expected: "pending",
		},
		{
			name:     "large negative number",
			input:    big.NewInt(-1000000),
			expected: "<invalid -1000000>",
		},
	}

	for _, tc := range negativeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toBlockNumArg(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToCallArg(t *testing.T) {
	address := common.HexToAddress("0xa7d9ddbe1f17865597fbd27ec712455208b6b76d")
	toAddress := common.HexToAddress("0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb")

	testCases := []struct {
		name     string
		msg      ethereum.CallMsg
		validate func(*testing.T, interface{})
	}{
		{
			name: "minimal call",
			msg: ethereum.CallMsg{
				From: address,
				To:   &toAddress,
			},
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, address, m["from"])
				assert.Equal(t, &toAddress, m["to"])
				_, hasInput := m["input"]
				assert.False(t, hasInput)
			},
		},
		{
			name: "call with data",
			msg: ethereum.CallMsg{
				From: address,
				To:   &toAddress,
				Data: []byte{0x01, 0x02, 0x03},
			},
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, hexutil.Bytes([]byte{0x01, 0x02, 0x03}), m["input"])
			},
		},
		{
			name: "call with value",
			msg: ethereum.CallMsg{
				From:  address,
				To:    &toAddress,
				Value: big.NewInt(1000000000000000000),
			},
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, (*hexutil.Big)(big.NewInt(1000000000000000000)), m["value"])
			},
		},
		{
			name: "call with gas",
			msg: ethereum.CallMsg{
				From: address,
				To:   &toAddress,
				Gas:  21000,
			},
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, hexutil.Uint64(21000), m["gas"])
			},
		},
		{
			name: "call with gas price",
			msg: ethereum.CallMsg{
				From:     address,
				To:       &toAddress,
				GasPrice: big.NewInt(20000000000),
			},
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, (*hexutil.Big)(big.NewInt(20000000000)), m["gasPrice"])
			},
		},
		{
			name: "call with maxFeePerGas and maxPriorityFeePerGas",
			msg: ethereum.CallMsg{
				From:      address,
				To:        &toAddress,
				GasFeeCap: big.NewInt(30000000000),
				GasTipCap: big.NewInt(2000000000),
			},
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, (*hexutil.Big)(big.NewInt(30000000000)), m["maxFeePerGas"])
				assert.Equal(t, (*hexutil.Big)(big.NewInt(2000000000)), m["maxPriorityFeePerGas"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			arg := toCallArg(tc.msg)
			tc.validate(t, arg)
		})
	}
}

func TestToFilterArg(t *testing.T) {
	address1 := common.HexToAddress("0xa7d9ddbe1f17865597fbd27ec712455208b6b76d")
	address2 := common.HexToAddress("0xf02c1c8e6114b1dbe8937a39260b5b0a374432bb")
	blockHash := common.HexToHash("0xa917fcc721a5465a484e9be17cda0cc5493933dd3bc70c9adbee192cb419c9d7")
	topic1 := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	testCases := []struct {
		name      string
		query     ethereum.FilterQuery
		shouldErr bool
		validate  func(*testing.T, interface{})
	}{
		{
			name: "filter with fromBlock and toBlock",
			query: ethereum.FilterQuery{
				FromBlock: big.NewInt(436),
				ToBlock:   big.NewInt(500),
				Addresses: []common.Address{address1},
				Topics:    [][]common.Hash{{topic1}},
			},
			shouldErr: false,
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "0x1b4", m["fromBlock"])
				assert.Equal(t, "0x1f4", m["toBlock"])
				assert.Equal(t, []common.Address{address1}, m["address"])
				assert.Equal(t, [][]common.Hash{{topic1}}, m["topics"])
			},
		},
		{
			name: "filter with nil blocks defaults fromBlock to 0x0",
			query: ethereum.FilterQuery{
				FromBlock: nil,
				ToBlock:   nil,
				Addresses: []common.Address{address1},
			},
			shouldErr: false,
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "0x0", m["fromBlock"])
				assert.Equal(t, "latest", m["toBlock"])
			},
		},
		{
			name: "filter with blockHash",
			query: ethereum.FilterQuery{
				BlockHash: &blockHash,
				Addresses: []common.Address{address1, address2},
			},
			shouldErr: false,
			validate: func(t *testing.T, arg interface{}) {
				m, ok := arg.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, blockHash, m["blockHash"])
				assert.Equal(t, []common.Address{address1, address2}, m["address"])
				_, hasFromBlock := m["fromBlock"]
				assert.False(t, hasFromBlock)
			},
		},
		{
			name: "filter with blockHash and fromBlock should error",
			query: ethereum.FilterQuery{
				BlockHash: &blockHash,
				FromBlock: big.NewInt(436),
				Addresses: []common.Address{address1},
			},
			shouldErr: true,
		},
		{
			name: "filter with blockHash and toBlock should error",
			query: ethereum.FilterQuery{
				BlockHash: &blockHash,
				ToBlock:   big.NewInt(500),
				Addresses: []common.Address{address1},
			},
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			arg, err := toFilterArg(tc.query)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, arg)
			} else {
				assert.NoError(t, err)
				if tc.validate != nil {
					tc.validate(t, arg)
				}
			}
		})
	}
}
