package gas_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas"
	mock_gas "github.com/status-im/go-wallet-sdk/pkg/gas/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// setupDefaultMockClient configures the mock client with default responses
func setupDefaultMockClient(ctrl *gomock.Controller) *mock_gas.MockEthClient {
	mockClient := mock_gas.NewMockEthClient(ctrl)

	// Default EstimateGas behavior
	mockClient.EXPECT().EstimateGas(gomock.Any(), gomock.Any()).
		Return(uint64(21000), nil).AnyTimes()

	// Default FeeHistory behavior
	mockClient.EXPECT().FeeHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
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

			return &ethereum.FeeHistory{
				BaseFee:      baseFees,
				Reward:       rewards,
				GasUsedRatio: make([]float64, int(blockCount)),
			}, nil
		}).AnyTimes()

	// Default LineaEstimateGas behavior
	mockClient.EXPECT().LineaEstimateGas(gomock.Any(), gomock.Any()).
		Return(&ethclient.LineaEstimateGasResult{
			GasLimit:          big.NewInt(21000),
			BaseFeePerGas:     big.NewInt(20000000000), // 20 gwei
			PriorityFeePerGas: big.NewInt(2000000000),  // 2 gwei
		}, nil).AnyTimes()

	return mockClient
}

func TestGetTxSuggestions_ChainClassL1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := setupDefaultMockClient(ctrl)

	params := gas.ChainParameters{
		ChainClass:       gas.ChainClassL1,
		NetworkBlockTime: 12,
	}

	config := gas.SuggestionsConfig{
		NetworkCongestionBlocks:           5,
		GasPriceEstimationBlocks:          10,
		LowRewardPercentile:               10,
		MediumRewardPercentile:            50,
		HighRewardPercentile:              90,
		LowBaseFeeMultiplier:              1.025,
		MediumBaseFeeMultiplier:           1.025,
		HighBaseFeeMultiplier:             1.025,
		LowBaseFeeCongestionMultiplier:    0.0,
		MediumBaseFeeCongestionMultiplier: 10.0,
		HighBaseFeeCongestionMultiplier:   10.0,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := gas.GetTxSuggestions(ctx, mockClient, params, config, callMsg)
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
	assert.GreaterOrEqual(t, fs.LowInclusion.MinTimeUntilInclusion, 0.0)
	assert.GreaterOrEqual(t, fs.LowInclusion.MaxTimeUntilInclusion, fs.LowInclusion.MinTimeUntilInclusion)
}

func TestGetTxSuggestions_ChainClassLineaStack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := setupDefaultMockClient(ctrl)

	params := gas.ChainParameters{
		ChainClass:       gas.ChainClassLineaStack,
		NetworkBlockTime: 2,
	}

	config := gas.SuggestionsConfig{
		NetworkCongestionBlocks:           5,
		GasPriceEstimationBlocks:          10,
		LowRewardPercentile:               10,
		MediumRewardPercentile:            50,
		HighRewardPercentile:              90,
		LowBaseFeeMultiplier:              1.025,
		MediumBaseFeeMultiplier:           1.025,
		HighBaseFeeMultiplier:             1.025,
		LowBaseFeeCongestionMultiplier:    0.0,
		MediumBaseFeeCongestionMultiplier: 10.0,
		HighBaseFeeCongestionMultiplier:   10.0,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := gas.GetTxSuggestions(ctx, mockClient, params, config, callMsg)
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := setupDefaultMockClient(ctrl)

	params := gas.ChainParameters{
		ChainClass:       gas.ChainClassArbStack,
		NetworkBlockTime: 0.25,
	}

	config := gas.SuggestionsConfig{
		NetworkCongestionBlocks:           5,
		GasPriceEstimationBlocks:          10,
		LowRewardPercentile:               10,
		MediumRewardPercentile:            50,
		HighRewardPercentile:              90,
		LowBaseFeeMultiplier:              1.025,
		MediumBaseFeeMultiplier:           1.025,
		HighBaseFeeMultiplier:             1.025,
		LowBaseFeeCongestionMultiplier:    0.0,
		MediumBaseFeeCongestionMultiplier: 10.0,
		HighBaseFeeCongestionMultiplier:   10.0,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := gas.GetTxSuggestions(ctx, mockClient, params, config, callMsg)
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.NotNil(t, suggestions.FeeSuggestions)
	assert.NotNil(t, suggestions.GasLimit)
}

func TestGetTxSuggestions_ChainClassOPStack(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := setupDefaultMockClient(ctrl)

	params := gas.ChainParameters{
		ChainClass:       gas.ChainClassOPStack,
		NetworkBlockTime: 2,
	}

	config := gas.SuggestionsConfig{
		NetworkCongestionBlocks:           5,
		GasPriceEstimationBlocks:          10,
		LowRewardPercentile:               10,
		MediumRewardPercentile:            50,
		HighRewardPercentile:              90,
		LowBaseFeeMultiplier:              1.025,
		MediumBaseFeeMultiplier:           1.025,
		HighBaseFeeMultiplier:             1.025,
		LowBaseFeeCongestionMultiplier:    0.0,
		MediumBaseFeeCongestionMultiplier: 10.0,
		HighBaseFeeCongestionMultiplier:   10.0,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := gas.GetTxSuggestions(ctx, mockClient, params, config, callMsg)
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.NotNil(t, suggestions.FeeSuggestions)
	assert.NotNil(t, suggestions.GasLimit)
}

func TestGetTxSuggestions_InvalidChainClass(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := setupDefaultMockClient(ctrl)

	params := gas.ChainParameters{
		ChainClass:       "InvalidChain",
		NetworkBlockTime: 12,
	}

	config := gas.SuggestionsConfig{
		NetworkCongestionBlocks:           5,
		GasPriceEstimationBlocks:          10,
		LowRewardPercentile:               10,
		MediumRewardPercentile:            50,
		HighRewardPercentile:              90,
		LowBaseFeeMultiplier:              1.025,
		MediumBaseFeeMultiplier:           1.025,
		HighBaseFeeMultiplier:             1.025,
		LowBaseFeeCongestionMultiplier:    0.0,
		MediumBaseFeeCongestionMultiplier: 10.0,
		HighBaseFeeCongestionMultiplier:   10.0,
	}

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	// Should fall back to L2 suggestions for unknown chain classes
	suggestions, err := gas.GetTxSuggestions(ctx, mockClient, params, config, callMsg)
	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.NotNil(t, suggestions.FeeSuggestions)
	assert.NotNil(t, suggestions.GasLimit)
}

func TestGetTxSuggestions_FeeHistoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := mock_gas.NewMockEthClient(ctrl)

	// Configure mock to return error for FeeHistory
	mockClient.EXPECT().FeeHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	mockClient.EXPECT().EstimateGas(gomock.Any(), gomock.Any()).
		Return(uint64(21000), nil).AnyTimes()

	params := gas.ChainParameters{
		ChainClass:       gas.ChainClassL1,
		NetworkBlockTime: 12,
	}
	config := gas.DefaultConfig(gas.ChainClassL1)
	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	_, err := gas.GetTxSuggestions(ctx, mockClient, params, config, callMsg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get fee history")
}
