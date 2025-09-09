package fetcher_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/balance/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	mock_multicall "github.com/status-im/go-wallet-sdk/pkg/multicall/mock"
)

func TestFetchNativeBalancesWithMulticall_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	// Generate test data
	addresses := make([]common.Address, 3)
	for i := 0; i < 3; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10
	multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Expected balances
	expectedBalances := []*big.Int{
		big.NewInt(1000000000000000000), // 1 ETH
		big.NewInt(500000000000000000),  // 0.5 ETH
		big.NewInt(250000000000000000),  // 0.25 ETH
	}

	// Mock expectations
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			require.Equal(t, ctx, opts.Context)
			require.Equal(t, big.NewInt(int64(atBlock)), opts.BlockNumber)
			require.Equal(t, false, requireSuccess)
			require.Len(t, calls, 3)

			// Verify calls are correctly formatted
			for _, call := range calls {
				assert.Equal(t, multicallAddress, call.Target)
				assert.NotEmpty(t, call.CallData)
			}

			// Create results
			results := make([]multicall3.IMulticall3Result, len(calls))
			for i := range calls {
				results[i] = multicall3.IMulticall3Result{
					Success:    true,
					ReturnData: expectedBalances[i].Bytes(),
				}
			}

			blockNumber := big.NewInt(1000)
			var blockHash [32]byte
			return blockNumber, blockHash, results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, mockMulticallCaller, multicallAddress, batchSize)

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

func TestFetchNativeBalancesWithMulticall_EmptyAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	addresses := []common.Address{}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10
	multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Test
	result, err := fetcher.FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, mockMulticallCaller, multicallAddress, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestFetchNativeBalancesWithMulticall_MulticallError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10
	multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Mock expectations - simulate multicall error
	expectedError := errors.New("multicall error")
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).Return(nil, [32]byte{}, nil, expectedError)

	// Test
	result, err := fetcher.FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, mockMulticallCaller, multicallAddress, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchNativeBalancesWithMulticall_LargeBatch(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	// Create more addresses than batch size to test chunking
	addresses := make([]common.Address, 25)
	for i := 0; i < 25; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10 // Should create 3 chunks: 10, 10, 5
	multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Mock expectations for multiple chunks
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			require.Len(t, calls, 10) // First chunk

			// Create results for first chunk
			results := make([]multicall3.IMulticall3Result, len(calls))
			for i := range calls {
				balance := new(big.Int).Mul(big.NewInt(int64(i)), big.NewInt(1000000000000000000))
				results[i] = multicall3.IMulticall3Result{
					Success:    true,
					ReturnData: balance.Bytes(),
				}
			}

			blockNumber := big.NewInt(1000)
			var blockHash [32]byte
			return blockNumber, blockHash, results, nil
		})

	mockMulticallCaller.EXPECT().ViewTryAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error) {
			require.Len(t, calls, 10) // Second chunk

			// Create results for second chunk
			results := make([]multicall3.IMulticall3Result, len(calls))
			for i := range calls {
				balance := new(big.Int).Mul(big.NewInt(int64(i+10)), big.NewInt(1000000000000000000))
				results[i] = multicall3.IMulticall3Result{
					Success:    true,
					ReturnData: balance.Bytes(),
				}
			}

			return results, nil
		})

	mockMulticallCaller.EXPECT().ViewTryAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error) {
			require.Len(t, calls, 5) // Third chunk

			// Create results for third chunk
			results := make([]multicall3.IMulticall3Result, len(calls))
			for i := range calls {
				balance := new(big.Int).Mul(big.NewInt(int64(i+20)), big.NewInt(1000000000000000000))
				results[i] = multicall3.IMulticall3Result{
					Success:    true,
					ReturnData: balance.Bytes(),
				}
			}

			return results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, mockMulticallCaller, multicallAddress, batchSize)

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

func TestFetchNativeBalancesWithMulticall_FailedResults(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10
	multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Mock expectations - simulate failed results
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			results := []multicall3.IMulticall3Result{
				{
					Success:    true,
					ReturnData: big.NewInt(1000000000000000000).Bytes(), // 1 ETH
				},
				{
					Success:    false, // Failed result
					ReturnData: []byte("call failed"),
				},
			}

			blockNumber := big.NewInt(1000)
			var blockHash [32]byte
			return blockNumber, blockHash, results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, mockMulticallCaller, multicallAddress, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)

	// First address should have balance, second should be missing due to failed result
	assert.Equal(t, big.NewInt(1000000000000000000), result[addresses[0]])
	// Second address should not be in the result map due to failed multicall result
	assert.NotContains(t, result, addresses[1])
}

func TestFetchNativeBalancesWithMulticall_ZeroBalance(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10
	multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Mock expectations - simulate zero balance
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			results := []multicall3.IMulticall3Result{
				{
					Success:    true,
					ReturnData: big.NewInt(0).Bytes(), // Zero balance
				},
			}

			blockNumber := big.NewInt(1000)
			var blockHash [32]byte
			return blockNumber, blockHash, results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, mockMulticallCaller, multicallAddress, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, big.NewInt(0), result[addresses[0]])
}

func TestFetchNativeBalancesWithMulticall_ContextCancellation(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10
	multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Cancel context before calling
	cancel()

	// Mock expectations
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			// Check if context is cancelled
			select {
			case <-opts.Context.Done():
				return nil, [32]byte{}, nil, opts.Context.Err()
			default:
				results := []multicall3.IMulticall3Result{
					{
						Success:    true,
						ReturnData: big.NewInt(1000000000000000000).Bytes(),
					},
				}
				blockNumber := big.NewInt(1000)
				var blockHash [32]byte
				return blockNumber, blockHash, results, nil
			}
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, mockMulticallCaller, multicallAddress, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, result)
}

func TestFetchNativeBalancesWithMulticall_InvalidData(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10
	multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Mock expectations - simulate invalid data
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			results := []multicall3.IMulticall3Result{
				{
					Success:    true,
					ReturnData: []byte{0x01, 0x02, 0x03}, // Invalid data that can't be converted to big.Int
				},
			}

			blockNumber := big.NewInt(1000)
			var blockHash [32]byte
			return blockNumber, blockHash, results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, mockMulticallCaller, multicallAddress, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	// The result should be a valid big.Int even with invalid data
	assert.NotNil(t, result[addresses[0]])
}

// Helper function to create test multicall results
func generateTestMulticallResults(count int) []multicall3.IMulticall3Result {
	results := make([]multicall3.IMulticall3Result, count)
	for i := 0; i < count; i++ {
		balance := new(big.Int).Mul(big.NewInt(int64(i)), big.NewInt(1000000000000000000))
		results[i] = multicall3.IMulticall3Result{
			Success:    true,
			ReturnData: balance.Bytes(),
		}
	}
	return results
}

func TestFetchErc20BalancesWithMulticall_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	// Generate test data
	accountAddresses := make([]common.Address, 3)
	tokenAddresses := make([]common.Address, 2)
	for i := 0; i < 3; i++ {
		accountAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}
	for i := 0; i < 2; i++ {
		tokenAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Expected balances for each account-token combination
	expectedBalances := [][]*big.Int{
		{big.NewInt(1000000000000000000), big.NewInt(500000000000000000)}, // Account 0: Token 0=1, Token 1=0.5
		{big.NewInt(250000000000000000), big.NewInt(750000000000000000)},  // Account 1: Token 0=0.25, Token 1=0.75
		{big.NewInt(300000000000000000), big.NewInt(900000000000000000)},  // Account 2: Token 0=0.3, Token 1=0.9
	}

	// Mock expectations
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).DoAndReturn(
		func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			require.Equal(t, ctx, opts.Context)
			require.Equal(t, big.NewInt(int64(atBlock)), opts.BlockNumber)
			require.Equal(t, false, requireSuccess)
			require.Len(t, calls, 6) // 3 accounts * 2 tokens

			// Verify calls are correctly formatted
			for i, call := range calls {
				tokenIdx := i % 2
				assert.Equal(t, tokenAddresses[tokenIdx], call.Target)
				assert.NotEmpty(t, call.CallData)
			}

			// Create results
			results := make([]multicall3.IMulticall3Result, len(calls))
			for i := range calls {
				accountIdx := i / 2
				tokenIdx := i % 2
				results[i] = multicall3.IMulticall3Result{
					Success:    true,
					ReturnData: expectedBalances[accountIdx][tokenIdx].Bytes(),
				}
			}

			blockNumber := big.NewInt(1000)
			var blockHash [32]byte
			return blockNumber, blockHash, results, nil
		})

	// Test
	result, err := fetcher.FetchErc20BalancesWithMulticall(ctx, accountAddresses, tokenAddresses, atBlock, mockMulticallCaller, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Verify balances are correct
	for i, accountAddr := range accountAddresses {
		assert.Contains(t, result, accountAddr)
		assert.Len(t, result[accountAddr], 2)
		for j, tokenAddr := range tokenAddresses {
			assert.Contains(t, result[accountAddr], tokenAddr)
			assert.Equal(t, expectedBalances[i][j], result[accountAddr][tokenAddr])
		}
	}
}

func TestFetchErc20BalancesWithMulticall_EmptyAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	accountAddresses := []common.Address{}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Test
	result, err := fetcher.FetchErc20BalancesWithMulticall(ctx, accountAddresses, tokenAddresses, atBlock, mockMulticallCaller, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestFetchErc20BalancesWithMulticall_EmptyTokenAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	accountAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Test
	result, err := fetcher.FetchErc20BalancesWithMulticall(ctx, accountAddresses, tokenAddresses, atBlock, mockMulticallCaller, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Len(t, result[accountAddresses[0]], 0)
}

func TestFetchErc20BalancesWithMulticall_MulticallError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockMulticallCaller := mock_multicall.NewMockCaller(ctrl)

	accountAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate multicall error
	expectedError := errors.New("multicall error")
	mockMulticallCaller.EXPECT().ViewTryBlockAndAggregate(gomock.Any(), false, gomock.Any()).Return(nil, [32]byte{}, nil, expectedError)

	// Test
	result, err := fetcher.FetchErc20BalancesWithMulticall(ctx, accountAddresses, tokenAddresses, atBlock, mockMulticallCaller, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}
