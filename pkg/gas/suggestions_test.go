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
func setupDefaultMockClient(ctrl *gomock.Controller) *mock_gas.MockGasClient {
	mockClient := mock_gas.NewMockGasClient(ctrl)

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

	mockClient.EXPECT().SuggestGasPrice(gomock.Any()).
		Return(big.NewInt(20000000000), nil).AnyTimes() // 20 gwei

	mockClient.EXPECT().BlockNumber(gomock.Any()).
		Return(uint64(123), nil).AnyTimes()
	mockClient.EXPECT().EthGetBlockByNumberWithFullTxs(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, number *big.Int) (*ethclient.BlockWithFullTxs, error) {
			return &ethclient.BlockWithFullTxs{
				Number:        big.NewInt(123),
				BaseFeePerGas: big.NewInt(20000000000), // present => EIP1559 for resolver
				Transactions: []ethclient.Transaction{
					{GasPrice: big.NewInt(10)},
					{GasPrice: big.NewInt(20)},
					{GasPrice: big.NewInt(30)},
				},
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

func TestGetTxSuggestions_LegacyFeeModel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockClient := setupDefaultMockClient(ctrl)

	params := gas.ChainParameters{
		ChainClass:       gas.ChainClassL1,
		NetworkBlockTime: 12,
		FeeModel:         gas.FeeModelLegacy,
	}

	config := gas.DefaultConfig(params.ChainClass)

	callMsg := &ethereum.CallMsg{
		To:    &common.Address{},
		Data:  []byte{},
		Value: big.NewInt(0),
	}

	suggestions, err := gas.GetTxSuggestions(ctx, mockClient, params, config, callMsg)
	require.NoError(t, err)
	require.NotNil(t, suggestions)
	require.NotNil(t, suggestions.FeeSuggestions)

	fs := suggestions.FeeSuggestions
	require.Equal(t, gas.FeeModelLegacy, fs.FeeModel)

	require.NotNil(t, fs.Low.GasPrice)
	require.NotNil(t, fs.Medium.GasPrice)
	require.NotNil(t, fs.High.GasPrice)

	assert.Nil(t, fs.Low.MaxFeePerGas)
	assert.Nil(t, fs.Low.MaxPriorityFeePerGas)
	assert.Nil(t, fs.Medium.MaxFeePerGas)
	assert.Nil(t, fs.Medium.MaxPriorityFeePerGas)
	assert.Nil(t, fs.High.MaxFeePerGas)
	assert.Nil(t, fs.High.MaxPriorityFeePerGas)
}

func TestEstimateInclusion_LegacyFeeModel_UnknownUpperBound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	inclusionClient := mock_gas.NewMockBlockInclusionEstimator(ctrl)
	inclusionClient.EXPECT().BlockNumber(gomock.Any()).Return(uint64(0), assert.AnError).AnyTimes()
	inclusionClient.EXPECT().EthGetBlockByNumberWithFullTxs(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).AnyTimes()
	inclusionClient.EXPECT().FeeHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, assert.AnError).AnyTimes()

	params := gas.ChainParameters{
		ChainClass:       gas.ChainClassL1,
		NetworkBlockTime: 12,
		FeeModel:         gas.FeeModelLegacy,
	}
	config := gas.DefaultConfig(params.ChainClass)

	fee := gas.Fee{
		GasPrice: big.NewInt(20_000_000_000), // 20 gwei
	}

	inc, err := gas.EstimateInclusion(ctx, inclusionClient, params, config, fee)
	require.NoError(t, err)
	require.NotNil(t, inc)

	assert.Equal(t, 1, inc.MinBlocksUntilInclusion)
	assert.Equal(t, 12.0, inc.MinTimeUntilInclusion)
	assert.Equal(t, -1, inc.MaxBlocksUntilInclusion)
	assert.Equal(t, -1.0, inc.MaxTimeUntilInclusion)
}

func TestResolveFeeModel_DetectsLegacyViaMissingBaseFee(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feeResolver := mock_gas.NewMockFeeModelResolver(ctrl)
	feeResolver.EXPECT().SuggestGasTipCap(gomock.Any()).Return(big.NewInt(1), nil).AnyTimes()
	feeResolver.EXPECT().FeeHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&ethereum.FeeHistory{
			BaseFee:      []*big.Int{big.NewInt(1), big.NewInt(1)},
			Reward:       [][]*big.Int{{big.NewInt(1)}},
			GasUsedRatio: []float64{1},
		}, nil).AnyTimes()
	feeResolver.EXPECT().BlockNumber(gomock.Any()).Return(uint64(123), nil).AnyTimes()
	feeResolver.EXPECT().EthGetBlockByNumberWithFullTxs(gomock.Any(), gomock.Any()).
		Return(&ethclient.BlockWithFullTxs{Number: big.NewInt(123), BaseFeePerGas: nil}, nil).AnyTimes()

	fm, err := gas.ResolveFeeModel(ctx, feeResolver)
	require.NoError(t, err)
	assert.Equal(t, gas.FeeModelLegacy, fm)
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
	mockClient := mock_gas.NewMockGasClient(ctrl)

	// Configure mock to return error for FeeHistory
	mockClient.EXPECT().FeeHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError).AnyTimes()

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
