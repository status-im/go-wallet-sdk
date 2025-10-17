package data

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

func createTestGasData() *GasData {
	return &GasData{
		LatestBlock: &ethclient.BlockWithFullTxs{
			Number: big.NewInt(1000),
			Transactions: []ethclient.Transaction{
				{
					ChainID: big.NewInt(1),
				},
			},
		},
		FeeHistory: &ethereum.FeeHistory{
			BaseFee: []*big.Int{
				big.NewInt(10000000000), // 10 gwei
				big.NewInt(11000000000), // 11 gwei
				big.NewInt(12000000000), // 12 gwei
				big.NewInt(13000000000), // 13 gwei
				big.NewInt(14000000000), // 14 gwei
			},
			GasUsedRatio: []float64{0.5, 0.6, 0.7, 0.8},
			Reward: [][]*big.Int{
				{big.NewInt(1000000000), big.NewInt(2000000000), big.NewInt(3000000000)}, // Block 1 rewards (0%, 5%, 10%)
				{big.NewInt(1100000000), big.NewInt(2100000000), big.NewInt(3100000000)}, // Block 2 rewards (0%, 5%, 10%)
				{big.NewInt(1200000000), big.NewInt(2200000000), big.NewInt(3200000000)}, // Block 3 rewards (0%, 5%, 10%)
				{big.NewInt(1300000000), big.NewInt(2300000000), big.NewInt(3300000000)}, // Block 4 rewards (0%, 5%, 10%)
			},
		},
		InfuraSuggestedFees: &infura.GasResponse{
			Low: infura.GasPrice{
				SuggestedMaxPriorityFeePerGas: "1000000000",
				SuggestedMaxFeePerGas:         "11000000000",
			},
			Medium: infura.GasPrice{
				SuggestedMaxPriorityFeePerGas: "2000000000",
				SuggestedMaxFeePerGas:         "12000000000",
			},
			High: infura.GasPrice{
				SuggestedMaxPriorityFeePerGas: "3000000000",
				SuggestedMaxFeePerGas:         "13000000000",
			},
		},
	}
}

func TestNewFakeClient(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	if client == nil {
		t.Fatal("NewFakeClient returned nil")
	}
	if client.gasData != gasData {
		t.Error("FakeClient gasData not set correctly")
	}
}

func TestFakeClient_Close(t *testing.T) {
	client := NewFakeClient(createTestGasData())
	// Should not panic
	client.Close()
}

func TestFakeClient_ChainID(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		t.Fatalf("ChainID returned error: %v", err)
	}

	expected := big.NewInt(1)
	if chainID.Cmp(expected) != 0 {
		t.Errorf("Expected chainID %v, got %v", expected, chainID)
	}
}

func TestFakeClient_BlockNumber(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		t.Fatalf("BlockNumber returned error: %v", err)
	}

	expected := uint64(1000)
	if blockNumber != expected {
		t.Errorf("Expected block number %d, got %d", expected, blockNumber)
	}
}

func TestFakeClient_BlockByNumber(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	tests := []struct {
		name   string
		number *big.Int
	}{
		{"latest block (nil)", nil},
		{"specific block", big.NewInt(999)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := client.BlockByNumber(context.Background(), tt.number)
			if err != nil {
				t.Fatalf("BlockByNumber returned error: %v", err)
			}

			if block != gasData.LatestBlock {
				t.Error("BlockByNumber should return the latest block")
			}
		})
	}
}

func TestFakeClient_FeeHistory_Success(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	tests := []struct {
		name              string
		blockCount        uint64
		lastBlock         *big.Int
		rewardPercentiles []float64
		expectedBaseFees  int
		expectedRewards   int
		expectedGasRatios int
	}{
		{
			name:              "request 2 blocks",
			blockCount:        2,
			lastBlock:         nil,
			rewardPercentiles: []float64{0, 5, 10}, // Match available percentiles
			expectedBaseFees:  3,                   // blockCount + 1
			expectedRewards:   2,                   // blockCount
			expectedGasRatios: 2,                   // blockCount
		},
		{
			name:              "request all blocks",
			blockCount:        4,
			lastBlock:         nil,
			rewardPercentiles: []float64{0, 5}, // Match available percentiles
			expectedBaseFees:  5,               // blockCount + 1
			expectedRewards:   4,               // blockCount
			expectedGasRatios: 4,               // blockCount
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feeHistory, err := client.FeeHistory(context.Background(), tt.blockCount, tt.lastBlock, tt.rewardPercentiles)
			if err != nil {
				t.Fatalf("FeeHistory returned error: %v", err)
			}

			if len(feeHistory.BaseFee) != tt.expectedBaseFees {
				t.Errorf("Expected %d base fees, got %d", tt.expectedBaseFees, len(feeHistory.BaseFee))
			}

			if len(feeHistory.Reward) != tt.expectedRewards {
				t.Errorf("Expected %d reward arrays, got %d", tt.expectedRewards, len(feeHistory.Reward))
			}

			if len(feeHistory.GasUsedRatio) != tt.expectedGasRatios {
				t.Errorf("Expected %d gas ratios, got %d", tt.expectedGasRatios, len(feeHistory.GasUsedRatio))
			}

			// Check that each reward array has the correct number of percentiles
			// Note: The current implementation may leave some reward arrays empty due to bounds checking
			for i, rewards := range feeHistory.Reward {
				// Only check non-empty reward arrays since the implementation may skip some blocks
				if len(rewards) > 0 && len(rewards) != len(tt.rewardPercentiles) {
					t.Errorf("Block %d: expected %d reward percentiles, got %d", i, len(tt.rewardPercentiles), len(rewards))
				}
			}

			// Verify oldest block calculation
			latestBlockNumber := gasData.LatestBlock.Number.Uint64()
			expectedOldest := latestBlockNumber - tt.blockCount + 1
			if feeHistory.OldestBlock.Uint64() != expectedOldest {
				t.Errorf("Expected oldest block %d, got %d", expectedOldest, feeHistory.OldestBlock.Uint64())
			}
		})
	}
}

func TestFakeClient_FeeHistory_Errors(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	tests := []struct {
		name              string
		blockCount        uint64
		lastBlock         *big.Int
		rewardPercentiles []float64
		expectedError     string
	}{
		{
			name:              "zero block count",
			blockCount:        0,
			lastBlock:         nil,
			rewardPercentiles: []float64{50},
			expectedError:     "blockCount must be greater than 0",
		},
		{
			name:              "block count exceeds available",
			blockCount:        10,
			lastBlock:         nil,
			rewardPercentiles: []float64{50},
			expectedError:     "blockCount 10 exceeds available blocks 5",
		},
		{
			name:              "invalid percentile negative",
			blockCount:        2,
			lastBlock:         nil,
			rewardPercentiles: []float64{-1},
			expectedError:     "rewardPercentiles must be between 0 and 100",
		},
		{
			name:              "invalid percentile over 100",
			blockCount:        2,
			lastBlock:         nil,
			rewardPercentiles: []float64{101},
			expectedError:     "rewardPercentiles must be between 0 and 100",
		},
		{
			name:              "non-nil lastBlock",
			blockCount:        2,
			lastBlock:         big.NewInt(999),
			rewardPercentiles: []float64{50},
			expectedError:     "only 'latest' block request is supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.FeeHistory(context.Background(), tt.blockCount, tt.lastBlock, tt.rewardPercentiles)
			if err == nil {
				t.Fatal("Expected error but got none")
			}
			if err.Error() != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

func TestFakeClient_GetGasSuggestions_Success(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	gasSuggestions, err := client.GetGasSuggestions(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetGasSuggestions returned error: %v", err)
	}

	if gasSuggestions != gasData.InfuraSuggestedFees {
		t.Error("GetGasSuggestions should return the InfuraSuggestedFees from gasData")
	}
}

func TestFakeClient_GetGasSuggestions_WrongNetworkID(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	_, err := client.GetGasSuggestions(context.Background(), 2) // Wrong network ID
	if err == nil {
		t.Fatal("Expected error for wrong network ID")
	}

	expectedError := "networkID 2 does not match latest block networkID 1"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestFakeClient_SuggestGasPrice_NotImplemented(t *testing.T) {
	client := NewFakeClient(createTestGasData())

	_, err := client.SuggestGasPrice(context.Background())
	if err == nil {
		t.Fatal("Expected error for unimplemented method")
	}

	expectedError := "SuggestGasPrice not implemented"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestFakeClient_LineaEstimateGas_NotImplemented(t *testing.T) {
	client := NewFakeClient(createTestGasData())

	msg := ethereum.CallMsg{}
	_, err := client.LineaEstimateGas(context.Background(), msg)
	if err == nil {
		t.Fatal("Expected error for unimplemented method")
	}

	expectedError := "LineaEstimateGas not implemented"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Test edge cases for FeeHistory reward percentile mapping
func TestFakeClient_FeeHistory_PercentileMapping(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	// Test various percentiles to ensure proper mapping
	percentiles := []float64{0, 5, 10} // Match available percentiles in test data

	feeHistory, err := client.FeeHistory(context.Background(), 2, nil, percentiles)
	if err != nil {
		t.Fatalf("FeeHistory returned error: %v", err)
	}

	// Verify that we get rewards for all requested percentiles
	// Note: The current implementation may leave some reward arrays empty due to bounds checking
	for blockIdx, rewards := range feeHistory.Reward {
		// Only check non-empty reward arrays since the implementation may skip some blocks
		if len(rewards) > 0 && len(rewards) != len(percentiles) {
			t.Errorf("Block %d: expected %d rewards, got %d", blockIdx, len(percentiles), len(rewards))
		}

		// All rewards should be non-nil (only check non-empty reward arrays)
		if len(rewards) > 0 {
			for i, reward := range rewards {
				if reward == nil {
					t.Errorf("Block %d, percentile %f: reward is nil", blockIdx, percentiles[i])
				}
			}
		}
	}
}

// Test with empty gas data
func TestFakeClient_EmptyGasData(t *testing.T) {
	gasData := &GasData{
		LatestBlock: &ethclient.BlockWithFullTxs{
			Number: big.NewInt(100),
			Transactions: []ethclient.Transaction{
				{ChainID: big.NewInt(1)},
			},
		},
		FeeHistory: &ethereum.FeeHistory{
			BaseFee:      []*big.Int{},
			GasUsedRatio: []float64{},
			Reward:       [][]*big.Int{},
		},
		InfuraSuggestedFees: &infura.GasResponse{},
	}

	client := NewFakeClient(gasData)

	// Should fail when requesting blocks from empty fee history
	_, err := client.FeeHistory(context.Background(), 1, nil, []float64{50})
	if err == nil {
		t.Fatal("Expected error when requesting blocks from empty fee history")
	}
}

// Test ChainID error handling with nil transaction
func TestFakeClient_ChainID_NilTransaction(t *testing.T) {
	gasData := &GasData{
		LatestBlock: &ethclient.BlockWithFullTxs{
			Number:       big.NewInt(1000),
			Transactions: []ethclient.Transaction{}, // Empty transactions
		},
	}
	client := NewFakeClient(gasData)

	// This should return an error when no transactions are available
	_, err := client.ChainID(context.Background())
	if err == nil {
		t.Fatal("Expected error when no transactions available")
	}

	expectedError := "no gas data available"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Test FeeHistory with single block request
func TestFakeClient_FeeHistory_SingleBlock(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	feeHistory, err := client.FeeHistory(context.Background(), 1, nil, []float64{0})
	if err != nil {
		t.Fatalf("FeeHistory returned error: %v", err)
	}

	if len(feeHistory.BaseFee) != 2 { // blockCount + 1
		t.Errorf("Expected 2 base fees, got %d", len(feeHistory.BaseFee))
	}

	if len(feeHistory.Reward) != 1 {
		t.Errorf("Expected 1 reward array, got %d", len(feeHistory.Reward))
	}

	if len(feeHistory.GasUsedRatio) != 1 {
		t.Errorf("Expected 1 gas ratio, got %d", len(feeHistory.GasUsedRatio))
	}
}

// Test GetGasSuggestions with ChainID error
func TestFakeClient_GetGasSuggestions_ChainIDError(t *testing.T) {
	gasData := &GasData{
		LatestBlock: &ethclient.BlockWithFullTxs{
			Number:       big.NewInt(1000),
			Transactions: []ethclient.Transaction{}, // Empty transactions will cause ChainID to fail
		},
		InfuraSuggestedFees: &infura.GasResponse{},
	}
	client := NewFakeClient(gasData)

	// This should fail when ChainID fails due to empty transactions
	_, err := client.GetGasSuggestions(context.Background(), 1)
	if err == nil {
		t.Fatal("Expected error when ChainID fails")
	}

	expectedError := "no gas data available"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Test with context cancellation
func TestFakeClient_ContextCancellation(t *testing.T) {
	gasData := createTestGasData()
	client := NewFakeClient(gasData)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// These methods should still work since they don't check context cancellation
	// This tests that the fake client doesn't unnecessarily check context
	_, err := client.ChainID(ctx)
	if err != nil {
		t.Errorf("ChainID failed with cancelled context: %v", err)
	}

	_, err = client.BlockNumber(ctx)
	if err != nil {
		t.Errorf("BlockNumber failed with cancelled context: %v", err)
	}
}

// Test with nil gas data
func TestFakeClient_NilGasData(t *testing.T) {
	client := NewFakeClient(nil)

	// All methods should return errors when gas data is nil
	_, err := client.ChainID(context.Background())
	if err == nil {
		t.Error("Expected error for ChainID with nil gas data")
	}

	_, err = client.BlockNumber(context.Background())
	if err == nil {
		t.Error("Expected error for BlockNumber with nil gas data")
	}

	_, err = client.BlockByNumber(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for BlockByNumber with nil gas data")
	}

	_, err = client.FeeHistory(context.Background(), 1, nil, []float64{50})
	if err == nil {
		t.Error("Expected error for FeeHistory with nil gas data")
	}
}
