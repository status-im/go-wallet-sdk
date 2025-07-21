package gas

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEthClient is a simple mock implementation for testing
type MockEthClient struct {
	estimateGasFunc      func(context.Context, ethereum.CallMsg) (uint64, error)
	feeHistoryFunc       func(context.Context, uint64, *big.Int, []float64) (*ethclient.FeeHistory, error)
	lineaEstimateGasFunc func(context.Context, ethereum.CallMsg) (*ethclient.LineaEstimateGasResult, error)
}

func (m *MockEthClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	if m.estimateGasFunc != nil {
		return m.estimateGasFunc(ctx, msg)
	}
	return 21000, nil // Default gas limit for simple transfers
}

func (m *MockEthClient) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethclient.FeeHistory, error) {
	if m.feeHistoryFunc != nil {
		return m.feeHistoryFunc(ctx, blockCount, lastBlock, rewardPercentiles)
	}

	// Return default fee history data
	baseFee := big.NewInt(20000000000)       // 20 gwei
	lowPriority := big.NewInt(1000000000)    // 1 gwei
	mediumPriority := big.NewInt(2000000000) // 2 gwei
	highPriority := big.NewInt(5000000000)   // 5 gwei

	baseFees := make([]*big.Int, int(blockCount)+1)
	rewards := make([][]*big.Int, int(blockCount))

	for i := range baseFees {
		baseFees[i] = new(big.Int).Set(baseFee)
	}

	for i := range rewards {
		rewards[i] = []*big.Int{lowPriority, mediumPriority, highPriority}
	}

	return &ethclient.FeeHistory{
		BaseFeePerGas: baseFees,
		Reward:        rewards,
		GasUsedRatio:  make([]float64, int(blockCount)),
	}, nil
}

func (m *MockEthClient) LineaEstimateGas(ctx context.Context, msg ethereum.CallMsg) (*ethclient.LineaEstimateGasResult, error) {
	if m.lineaEstimateGasFunc != nil {
		return m.lineaEstimateGasFunc(ctx, msg)
	}

	// Return default Linea estimate
	return &ethclient.LineaEstimateGasResult{
		GasLimit:          big.NewInt(21000),
		BaseFeePerGas:     big.NewInt(20000000000), // 20 gwei
		PriorityFeePerGas: big.NewInt(2000000000),  // 2 gwei
	}, nil
}

func TestGetTxSuggestions_ChainClassL1(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockEthClient{}
	config := SuggestionsConfig{
		ChainClass:               ChainClassL1,
		NetworkBlockTime:         12,
		NetworkCongestionBlocks:  5,
		GasPriceEstimationBlocks: 10,
		LowRewardPercentile:      10,
		MediumRewardPercentile:   50,
		HighRewardPercentile:     90,
		BaseFeeMultiplier:        1.025,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := GetTxSuggestions(ctx, mockClient, config, callMsg)
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.NotNil(t, suggestions.FeeSuggestions)
	assert.NotNil(t, suggestions.GasLimit)

	// Verify fee suggestions structure
	fs := suggestions.FeeSuggestions
	assert.NotNil(t, fs.Low)
	assert.NotNil(t, fs.Medium)
	assert.NotNil(t, fs.High)
	assert.NotNil(t, fs.EstimatedBaseFee)
	assert.NotNil(t, fs.PriorityFeeLowerBound)
	assert.NotNil(t, fs.PriorityFeeUpperBound)

	// Verify time estimates are set
	assert.GreaterOrEqual(t, fs.Low.MinTimeUntilInclusion, 0.0)
	assert.GreaterOrEqual(t, fs.Low.MaxTimeUntilInclusion, fs.Low.MinTimeUntilInclusion)
}

func TestGetTxSuggestions_ChainClassLineaStack(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockEthClient{}
	config := SuggestionsConfig{
		ChainClass:               ChainClassLineaStack,
		NetworkBlockTime:         2,
		NetworkCongestionBlocks:  5,
		GasPriceEstimationBlocks: 10,
		LowRewardPercentile:      10,
		MediumRewardPercentile:   50,
		HighRewardPercentile:     90,
		BaseFeeMultiplier:        1.025,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := GetTxSuggestions(ctx, mockClient, config, callMsg)
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.NotNil(t, suggestions.FeeSuggestions)
	assert.NotNil(t, suggestions.GasLimit)

	// Verify Linea-specific structure
	fs := suggestions.FeeSuggestions
	assert.NotNil(t, fs.Low)
	assert.NotNil(t, fs.Medium)
	assert.NotNil(t, fs.High)
}

func TestGetTxSuggestions_ChainClassArbStack(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockEthClient{}
	config := SuggestionsConfig{
		ChainClass:               ChainClassArbStack,
		NetworkBlockTime:         0.25,
		NetworkCongestionBlocks:  5,
		GasPriceEstimationBlocks: 10,
		LowRewardPercentile:      10,
		MediumRewardPercentile:   50,
		HighRewardPercentile:     90,
		BaseFeeMultiplier:        1.025,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := GetTxSuggestions(ctx, mockClient, config, callMsg)
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.NotNil(t, suggestions.FeeSuggestions)
	assert.NotNil(t, suggestions.GasLimit)
}

func TestGetTxSuggestions_ChainClassOPStack(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockEthClient{}
	config := SuggestionsConfig{
		ChainClass:               ChainClassOPStack,
		NetworkBlockTime:         2,
		NetworkCongestionBlocks:  5,
		GasPriceEstimationBlocks: 10,
		LowRewardPercentile:      10,
		MediumRewardPercentile:   50,
		HighRewardPercentile:     90,
		BaseFeeMultiplier:        1.025,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := GetTxSuggestions(ctx, mockClient, config, callMsg)
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.NotNil(t, suggestions.FeeSuggestions)
	assert.NotNil(t, suggestions.GasLimit)
}

func TestGetTxSuggestions_InvalidChainClass(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockEthClient{}
	config := SuggestionsConfig{
		ChainClass:               "InvalidChain",
		NetworkBlockTime:         12,
		NetworkCongestionBlocks:  5,
		GasPriceEstimationBlocks: 10,
		LowRewardPercentile:      10,
		MediumRewardPercentile:   50,
		HighRewardPercentile:     90,
		BaseFeeMultiplier:        1.025,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	// Should fall back to L2 suggestions for unknown chain classes
	suggestions, err := GetTxSuggestions(ctx, mockClient, config, callMsg)
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.NotNil(t, suggestions.FeeSuggestions)
	assert.NotNil(t, suggestions.GasLimit)
}

func TestGetTxSuggestions_FeeHistoryError(t *testing.T) {
	ctx := context.Background()
	mockClient := &MockEthClient{
		feeHistoryFunc: func(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethclient.FeeHistory, error) {
			return nil, assert.AnError
		},
	}

	config := DefaultConfig()
	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	_, err := GetTxSuggestions(ctx, mockClient, config, callMsg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get fee history")
}
