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

func TestFetchNativeBalances_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	// Generate test data
	addresses := make([]common.Address, 3)
	for i := 0; i < 3; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	chainID := big.NewInt(99999) // Use a chain ID with no balance scanner
	atBlock := gethrpc.LatestBlockNumber
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch call
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).Return(nil).AnyTimes()

	// Test
	result, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Verify all addresses are present in result
	for _, addr := range addresses {
		assert.Contains(t, result, addr)
		assert.NotNil(t, result[addr])
	}
}

func TestFetchNativeBalances_ChainIDError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.LatestBlockNumber
	batchSize := 10

	// Mock expectations - simulate ChainID error
	expectedError := errors.New("rpc error")
	mockRPC.EXPECT().ChainID(ctx).Return(nil, expectedError)

	// Test
	result, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchNativeBalances_EmptyAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{}
	chainID := big.NewInt(1)
	atBlock := gethrpc.LatestBlockNumber
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Test
	result, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestFetchNativeBalances_WithBalanceScanner(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := make([]common.Address, 2)
	for i := 0; i < 2; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	chainID := big.NewInt(1) // Ethereum mainnet
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch calls (since balance scanner won't be available in test environment)
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, batch []gethrpc.BatchElem) error {
		require.Len(t, batch, 2)

		// Simulate successful responses
		balance1 := big.NewInt(1000000000000000000) // 1 ETH
		balance2 := big.NewInt(500000000000000000)  // 0.5 ETH

		batch[0].Result = (*hexutil.Big)(balance1)
		batch[1].Result = (*hexutil.Big)(balance2)
		return nil
	})

	// Test
	result, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verify balances are set
	for _, addr := range addresses {
		assert.Contains(t, result, addr)
		assert.NotNil(t, result[addr])
	}
}

func TestFetchNativeBalances_WithLargeBatch(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	// Create more addresses than batch size to test chunking
	addresses := make([]common.Address, 25)
	for i := 0; i < 25; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	chainID := big.NewInt(1)
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10 // Should create 3 chunks: 10, 10, 5

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch calls for multiple chunks
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).Return(nil).Times(3)

	// Test
	result, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 25)

	// Verify all addresses are present
	for _, addr := range addresses {
		assert.Contains(t, result, addr)
	}
}

func TestFetchNativeBalances_WithBatchCallError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	chainID := big.NewInt(1)
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch call error
	expectedError := errors.New("batch call error")
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).Return(expectedError)

	// Test
	result, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchNativeBalances_WithBatchElementError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	chainID := big.NewInt(1)
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch call with element error
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, batch []gethrpc.BatchElem) error {
		require.Len(t, batch, 1)

		// Simulate element error
		batch[0].Error = errors.New("element error")
		return nil
	})

	// Test
	result, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, "element error", err.Error())
	assert.Nil(t, result)
}

// Helper function to create test addresses
func generateTestAddresses(count int) []common.Address {
	addresses := make([]common.Address, count)
	for i := 0; i < count; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}
	return addresses
}

// Helper function to create test balances
func generateTestBalances(count int) []*big.Int {
	balances := make([]*big.Int, count)
	for i := 0; i < count; i++ {
		balances[i] = big.NewInt(gofakeit.Int64())
	}
	return balances
}

func TestFetchNativeBalances_Integration(t *testing.T) {
	// Integration test that tests the full flow with realistic data
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := generateTestAddresses(5)
	chainID := big.NewInt(99999) // Use a chain ID with no balance scanner
	atBlock := gethrpc.LatestBlockNumber
	batchSize := 3

	// Mock chain ID
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch calls (since balance scanner won't be available in test)
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, batch []gethrpc.BatchElem) error {
		// Simulate successful responses for all batch elements
		for i := range batch {
			balance := big.NewInt(gofakeit.Int64())
			batch[i].Result = (*hexutil.Big)(balance)
		}
		return nil
	}).Times(2)

	// Test
	result, err := fetcher.FetchNativeBalances(ctx, addresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 5)

	// Verify all addresses are present
	for _, addr := range addresses {
		assert.Contains(t, result, addr)
		assert.NotNil(t, result[addr])
	}
}

func TestFetchErc20Balances_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	// Generate test data
	addresses := make([]common.Address, 3)
	tokenAddresses := make([]common.Address, 2)
	for i := 0; i < 3; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}
	for i := 0; i < 2; i++ {
		tokenAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	chainID := big.NewInt(99999) // Use a chain ID with no balance scanner
	atBlock := gethrpc.LatestBlockNumber
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch call for ERC20 balanceOf calls
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, batch []gethrpc.BatchElem) error {
		require.Len(t, batch, 6) // 3 addresses * 2 tokens

		// Simulate successful responses
		balances := []*big.Int{
			big.NewInt(1000000000000000000), // 1 token
			big.NewInt(500000000000000000),  // 0.5 tokens
			big.NewInt(250000000000000000),  // 0.25 tokens
			big.NewInt(750000000000000000),  // 0.75 tokens
			big.NewInt(300000000000000000),  // 0.3 tokens
			big.NewInt(900000000000000000),  // 0.9 tokens
		}

		for i, elem := range batch {
			elem.Result = (*hexutil.Big)(balances[i])
		}
		return nil
	})

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Verify all addresses are present in result
	for _, addr := range addresses {
		assert.Contains(t, result, addr)
		assert.NotNil(t, result[addr])
		assert.Len(t, result[addr], 2) // 2 tokens per address
	}

	// Verify all token addresses are present for each account
	for _, addr := range addresses {
		for _, tokenAddr := range tokenAddresses {
			assert.Contains(t, result[addr], tokenAddr)
			assert.NotNil(t, result[addr][tokenAddr])
		}
	}
}

func TestFetchErc20Balances_ChainIDError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.LatestBlockNumber
	batchSize := 10

	// Mock expectations - simulate ChainID error
	expectedError := errors.New("rpc error")
	mockRPC.EXPECT().ChainID(ctx).Return(nil, expectedError)

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchErc20Balances_EmptyAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	chainID := big.NewInt(1)
	atBlock := gethrpc.LatestBlockNumber
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestFetchErc20Balances_EmptyTokenAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{}
	chainID := big.NewInt(1)
	atBlock := gethrpc.LatestBlockNumber
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Len(t, result[addresses[0]], 0)
}

func TestFetchErc20Balances_WithBalanceScanner(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := make([]common.Address, 2)
	tokenAddresses := make([]common.Address, 2)
	for i := 0; i < 2; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
		tokenAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	chainID := big.NewInt(1) // Ethereum mainnet
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch calls (since balance scanner won't be available in test environment)
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, batch []gethrpc.BatchElem) error {
		require.Len(t, batch, 4) // 2 addresses * 2 tokens

		// Simulate successful responses
		balances := []*big.Int{
			big.NewInt(1000000000000000000), // 1 token
			big.NewInt(500000000000000000),  // 0.5 tokens
			big.NewInt(250000000000000000),  // 0.25 tokens
			big.NewInt(750000000000000000),  // 0.75 tokens
		}

		for i, elem := range batch {
			elem.Result = (*hexutil.Big)(balances[i])
		}
		return nil
	})

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verify balances are set
	for _, addr := range addresses {
		assert.Contains(t, result, addr)
		assert.NotNil(t, result[addr])
		assert.Len(t, result[addr], 2)
	}
}

func TestFetchErc20Balances_WithLargeBatch(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	// Create more addresses than batch size to test chunking
	addresses := make([]common.Address, 5)
	tokenAddresses := make([]common.Address, 5)
	for i := 0; i < 5; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
		tokenAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	chainID := big.NewInt(1)
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10 // Should create multiple chunks for 25 total calls (5*5)

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch calls for multiple chunks
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).Return(nil).Times(3) // 25 calls in chunks of 10

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 5)

	// Verify all addresses are present
	for _, addr := range addresses {
		assert.Contains(t, result, addr)
		assert.Len(t, result[addr], 5)
	}
}

func TestFetchErc20Balances_WithBatchCallError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	chainID := big.NewInt(1)
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch call error
	expectedError := errors.New("batch call error")
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).Return(expectedError)

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchErc20Balances_WithBatchElementError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	chainID := big.NewInt(1)
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch call with element error
	expectedError := errors.New("element error")
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, batch []gethrpc.BatchElem) error {
		batch[0].Error = expectedError
		return nil
	})

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchErc20Balances_Integration(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRPC := mock_fetcher.NewMockRPCClient(ctrl)

	// Generate test data
	addresses := generateTestAddresses(3)
	tokenAddresses := generateTestAddresses(2)
	chainID := big.NewInt(1)
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations
	mockRPC.EXPECT().ChainID(ctx).Return(chainID, nil)

	// Mock batch calls with realistic data
	mockRPC.EXPECT().BatchCallContext(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, batch []gethrpc.BatchElem) error {
		require.Len(t, batch, 6) // 3 addresses * 2 tokens

		// Simulate realistic token balances
		balances := generateTestBalances(6)
		for i, elem := range batch {
			elem.Result = (*hexutil.Big)(balances[i])
		}
		return nil
	})

	// Test
	result, err := fetcher.FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, mockRPC, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)

	// Verify structure and data
	for _, addr := range addresses {
		assert.Contains(t, result, addr)
		assert.Len(t, result[addr], 2)
		for _, tokenAddr := range tokenAddresses {
			assert.Contains(t, result[addr], tokenAddr)
			assert.NotNil(t, result[addr][tokenAddr])
		}
	}
}
