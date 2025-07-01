package fetcher_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/balance/fetcher"
	mock_fetcher "github.com/status-im/go-wallet-sdk/pkg/balance/fetcher/mock"
)

func TestFetchNativeBalancesStandard_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBatchCaller := mock_fetcher.NewMockBatchCaller(ctrl)

	// Generate test data
	addresses := make([]common.Address, 3)
	for i := 0; i < 3; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Expected balances
	expectedBalances := []*big.Int{
		big.NewInt(1000000000000000000), // 1 ETH
		big.NewInt(500000000000000000),  // 0.5 ETH
		big.NewInt(250000000000000000),  // 0.25 ETH
	}

	// Mock expectations
	mockBatchCaller.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, batch []gethrpc.BatchElem) error {
			require.Len(t, batch, 3)

			// Verify batch elements are correctly formatted and set results
			for i, elem := range batch {
				assert.Equal(t, "eth_getBalance", elem.Method)
				assert.Len(t, elem.Args, 2)
				assert.Equal(t, addresses[i], elem.Args[0])
				assert.Equal(t, atBlock, elem.Args[1])

				// Set the result directly on the batch element
				if hexResult, ok := elem.Result.(*hexutil.Big); ok {
					*hexResult = *(*hexutil.Big)(expectedBalances[i])
				}
			}
			return nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesStandard(ctx, addresses, atBlock, mockBatchCaller, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Verify balances are correct
	for i, addr := range addresses {
		assert.Contains(t, result, addr)
		assert.Equal(t, expectedBalances[i], result[addr])
	}
}

func TestFetchNativeBalancesStandard_EmptyAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBatchCaller := mock_fetcher.NewMockBatchCaller(ctrl)

	addresses := []common.Address{}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Test
	result, err := fetcher.FetchNativeBalancesStandard(ctx, addresses, atBlock, mockBatchCaller, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestFetchNativeBalancesStandard_BatchCallError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBatchCaller := mock_fetcher.NewMockBatchCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate batch call error
	expectedError := errors.New("batch call error")
	mockBatchCaller.EXPECT().BatchCallContext(ctx, gomock.Any()).Return(expectedError)

	// Test
	result, err := fetcher.FetchNativeBalancesStandard(ctx, addresses, atBlock, mockBatchCaller, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchNativeBalancesStandard_BatchElementError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBatchCaller := mock_fetcher.NewMockBatchCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate batch element error
	mockBatchCaller.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, batch []gethrpc.BatchElem) error {
			require.Len(t, batch, 1)

			// Simulate element error
			batch[0].Error = errors.New("element error")
			return nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesStandard(ctx, addresses, atBlock, mockBatchCaller, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, "element error", err.Error())
	assert.Nil(t, result)
}

func TestFetchNativeBalancesStandard_LargeBatch(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBatchCaller := mock_fetcher.NewMockBatchCaller(ctrl)

	// Create more addresses than batch size to test chunking
	addresses := make([]common.Address, 25)
	for i := 0; i < 25; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10 // Should create 3 chunks: 10, 10, 5

	// Mock expectations for multiple chunks
	mockBatchCaller.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, batch []gethrpc.BatchElem) error {
			require.Len(t, batch, 10) // First chunk

			// Set results for first chunk
			for i := range batch {
				balance := new(big.Int).Mul(big.NewInt(int64(i)), big.NewInt(1000000000000000000))
				if hexResult, ok := batch[i].Result.(*hexutil.Big); ok {
					*hexResult = *(*hexutil.Big)(balance)
				}
			}
			return nil
		})

	mockBatchCaller.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, batch []gethrpc.BatchElem) error {
			require.Len(t, batch, 10) // Second chunk

			// Set results for second chunk
			for i := range batch {
				balance := new(big.Int).Mul(big.NewInt(int64(i+10)), big.NewInt(1000000000000000000))
				if hexResult, ok := batch[i].Result.(*hexutil.Big); ok {
					*hexResult = *(*hexutil.Big)(balance)
				}
			}
			return nil
		})

	mockBatchCaller.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, batch []gethrpc.BatchElem) error {
			require.Len(t, batch, 5) // Third chunk

			// Set results for third chunk
			for i := range batch {
				balance := new(big.Int).Mul(big.NewInt(int64(i+20)), big.NewInt(1000000000000000000))
				if hexResult, ok := batch[i].Result.(*hexutil.Big); ok {
					*hexResult = *(*hexutil.Big)(balance)
				}
			}
			return nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesStandard(ctx, addresses, atBlock, mockBatchCaller, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 25)

	// Verify all addresses are present with correct balances
	for i, addr := range addresses {
		assert.Contains(t, result, addr)
		expectedBalance := new(big.Int).Mul(big.NewInt(int64(i)), big.NewInt(1000000000000000000))
		assert.Equal(t, expectedBalance, result[addr])
	}
}

func TestFetchNativeBalancesStandard_ZeroBalance(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBatchCaller := mock_fetcher.NewMockBatchCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate zero balance
	mockBatchCaller.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, batch []gethrpc.BatchElem) error {
			require.Len(t, batch, 1)

			// Set zero balance result
			if hexResult, ok := batch[0].Result.(*hexutil.Big); ok {
				*hexResult = *(*hexutil.Big)(big.NewInt(0))
			}
			return nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesStandard(ctx, addresses, atBlock, mockBatchCaller, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, big.NewInt(0), result[addresses[0]])
}

func TestFetchNativeBalancesStandard_ContextCancellation(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	mockBatchCaller := mock_fetcher.NewMockBatchCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Cancel context before calling
	cancel()

	// Mock expectations
	mockBatchCaller.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, batch []gethrpc.BatchElem) error {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Set a result
				if hexResult, ok := batch[0].Result.(*hexutil.Big); ok {
					*hexResult = *(*hexutil.Big)(big.NewInt(1000000000000000000))
				}
				return nil
			}
		})

	// Test
	result, err := fetcher.FetchNativeBalancesStandard(ctx, addresses, atBlock, mockBatchCaller, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, result)
}
