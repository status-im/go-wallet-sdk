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
	mock_balancescanner "github.com/status-im/go-wallet-sdk/pkg/balance/fetcher/mock"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/balancescanner"
)

func TestFetchNativeBalancesWithBalanceScanner_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

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
	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			require.Equal(t, ctx, opts.Context)
			require.Equal(t, big.NewInt(int64(atBlock)), opts.BlockNumber)
			require.Equal(t, addresses, addrs)

			results := make([]balancescanner.BalanceScannerResult, len(addrs))
			for i := range addrs {
				results[i] = balancescanner.BalanceScannerResult{
					Success: true,
					Data:    expectedBalances[i].Bytes(),
				}
			}
			return results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, mockBalanceScanner, batchSize)

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

func TestFetchNativeBalancesWithBalanceScanner_EmptyAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	addresses := []common.Address{}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Test
	result, err := fetcher.FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestFetchNativeBalancesWithBalanceScanner_BalanceScannerError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate BalanceScanner error
	expectedError := errors.New("balance scanner error")
	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses).Return(nil, expectedError)

	// Test
	result, err := fetcher.FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchNativeBalancesWithBalanceScanner_LargeBatch(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	// Create more addresses than batch size to test chunking
	addresses := make([]common.Address, 25)
	for i := 0; i < 25; i++ {
		addresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10 // Should create 3 chunks: 10, 10, 5

	// Mock expectations for multiple chunks
	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses[:10]).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			results := make([]balancescanner.BalanceScannerResult, len(addrs))
			for i := range addrs {
				balance := new(big.Int).Mul(big.NewInt(int64(i)), big.NewInt(1000000000000000000))
				results[i] = balancescanner.BalanceScannerResult{
					Success: true,
					Data:    balance.Bytes(), // 0, 1, 2, ... ETH
				}
			}
			return results, nil
		})

	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses[10:20]).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			results := make([]balancescanner.BalanceScannerResult, len(addrs))
			for i := range addrs {
				balance := new(big.Int).Mul(big.NewInt(int64(i+10)), big.NewInt(1000000000000000000))
				results[i] = balancescanner.BalanceScannerResult{
					Success: true,
					Data:    balance.Bytes(), // 10, 11, 12, ... ETH
				}
			}
			return results, nil
		})

	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses[20:]).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			results := make([]balancescanner.BalanceScannerResult, len(addrs))
			for i := range addrs {
				balance := new(big.Int).Mul(big.NewInt(int64(i+20)), big.NewInt(1000000000000000000))
				results[i] = balancescanner.BalanceScannerResult{
					Success: true,
					Data:    balance.Bytes(), // 20, 21, 22, 23, 24 ETH
				}
			}
			return results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, mockBalanceScanner, batchSize)

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

func TestFetchNativeBalancesWithBalanceScanner_FailedResults(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate failed results
	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			results := []balancescanner.BalanceScannerResult{
				{
					Success: true,
					Data:    big.NewInt(1000000000000000000).Bytes(), // 1 ETH
				},
				{
					Success: false, // Failed result
					Data:    []byte{},
				},
			}
			return results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// First address should have balance, second should have zero balance due to failed result
	assert.Equal(t, big.NewInt(1000000000000000000), result[addresses[0]])
	assert.Equal(t, big.NewInt(0), result[addresses[1]])
}

func TestFetchNativeBalancesWithBalanceScanner_ZeroBalance(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate zero balance
	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			results := []balancescanner.BalanceScannerResult{
				{
					Success: true,
					Data:    big.NewInt(0).Bytes(), // Zero balance
				},
			}
			return results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, big.NewInt(0), result[addresses[0]])
}

func TestFetchNativeBalancesWithBalanceScanner_ContextCancellation(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Cancel context before calling
	cancel()

	// Mock expectations
	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			// Check if context is cancelled
			select {
			case <-opts.Context.Done():
				return nil, opts.Context.Err()
			default:
				results := []balancescanner.BalanceScannerResult{
					{
						Success: true,
						Data:    big.NewInt(1000000000000000000).Bytes(),
					},
				}
				return results, nil
			}
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, result)
}

func TestFetchNativeBalancesWithBalanceScanner_InvalidData(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	addresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate invalid data
	mockBalanceScanner.EXPECT().EtherBalances(gomock.Any(), addresses).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			results := []balancescanner.BalanceScannerResult{
				{
					Success: true,
					Data:    []byte{0x01, 0x02, 0x03}, // Invalid data that can't be converted to big.Int
				},
			}
			return results, nil
		})

	// Test
	result, err := fetcher.FetchNativeBalancesWithBalanceScanner(ctx, addresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	// The result should be a valid big.Int even with invalid data
	assert.NotNil(t, result[addresses[0]])
}

// Helper function to create test balance results
func generateTestBalanceResults(count int) []balancescanner.BalanceScannerResult {
	results := make([]balancescanner.BalanceScannerResult, count)
	for i := 0; i < count; i++ {
		balance := new(big.Int).Mul(big.NewInt(int64(i)), big.NewInt(1000000000000000000))
		results[i] = balancescanner.BalanceScannerResult{
			Success: true,
			Data:    balance.Bytes(),
		}
	}
	return results
}

func TestFetchErc20BalancesWithBalanceScanner_Success(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

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

	// Since we have more accounts than tokens, it should loop over tokens and batch accounts
	// Mock TokenBalances calls for each token
	for tokenIdx, tokenAddress := range tokenAddresses {
		mockBalanceScanner.EXPECT().TokenBalances(gomock.Any(), accountAddresses, tokenAddress).DoAndReturn(
			func(opts *bind.CallOpts, addrs []common.Address, token common.Address) ([]balancescanner.BalanceScannerResult, error) {
				require.Equal(t, ctx, opts.Context)
				require.Equal(t, big.NewInt(int64(atBlock)), opts.BlockNumber)
				require.Equal(t, accountAddresses, addrs)
				require.Equal(t, tokenAddress, token)

				results := make([]balancescanner.BalanceScannerResult, len(addrs))
				for i := range addrs {
					results[i] = balancescanner.BalanceScannerResult{
						Success: true,
						Data:    expectedBalances[i][tokenIdx].Bytes(),
					}
				}
				return results, nil
			})
	}

	// Test
	result, err := fetcher.FetchErc20BalancesWithBalanceScanner(ctx, accountAddresses, tokenAddresses, atBlock, mockBalanceScanner, batchSize)

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

func TestFetchErc20BalancesWithBalanceScanner_MoreTokensThanAccounts(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	// Generate test data - more tokens than accounts
	accountAddresses := make([]common.Address, 2)
	tokenAddresses := make([]common.Address, 3)
	for i := 0; i < 2; i++ {
		accountAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}
	for i := 0; i < 3; i++ {
		tokenAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Expected balances for each account-token combination
	expectedBalances := [][]*big.Int{
		{big.NewInt(1000000000000000000), big.NewInt(500000000000000000), big.NewInt(250000000000000000)}, // Account 0
		{big.NewInt(750000000000000000), big.NewInt(300000000000000000), big.NewInt(900000000000000000)},  // Account 1
	}

	// Since we have more tokens than accounts, it should loop over accounts and batch tokens
	// Mock TokensBalance calls for each account
	for accountIdx, accountAddress := range accountAddresses {
		mockBalanceScanner.EXPECT().TokensBalance(gomock.Any(), accountAddress, tokenAddresses).DoAndReturn(
			func(opts *bind.CallOpts, owner common.Address, contracts []common.Address) ([]balancescanner.BalanceScannerResult, error) {
				require.Equal(t, ctx, opts.Context)
				require.Equal(t, big.NewInt(int64(atBlock)), opts.BlockNumber)
				require.Equal(t, accountAddress, owner)
				require.Equal(t, tokenAddresses, contracts)

				results := make([]balancescanner.BalanceScannerResult, len(contracts))
				for i := range contracts {
					results[i] = balancescanner.BalanceScannerResult{
						Success: true,
						Data:    expectedBalances[accountIdx][i].Bytes(),
					}
				}
				return results, nil
			})
	}

	// Test
	result, err := fetcher.FetchErc20BalancesWithBalanceScanner(ctx, accountAddresses, tokenAddresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// Verify balances are correct
	for i, accountAddr := range accountAddresses {
		assert.Contains(t, result, accountAddr)
		assert.Len(t, result[accountAddr], 3)
		for j, tokenAddr := range tokenAddresses {
			assert.Contains(t, result[accountAddr], tokenAddr)
			assert.Equal(t, expectedBalances[i][j], result[accountAddr][tokenAddr])
		}
	}
}

func TestFetchErc20BalancesWithBalanceScanner_EmptyAccountAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	accountAddresses := []common.Address{}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Test
	result, err := fetcher.FetchErc20BalancesWithBalanceScanner(ctx, accountAddresses, tokenAddresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestFetchErc20BalancesWithBalanceScanner_EmptyTokenAddresses(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	accountAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Test
	result, err := fetcher.FetchErc20BalancesWithBalanceScanner(ctx, accountAddresses, tokenAddresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Len(t, result[accountAddresses[0]], 0)
}

func TestFetchErc20BalancesWithBalanceScanner_BalanceScannerError(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	accountAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate BalanceScanner error
	// Since len(accountAddresses) == len(tokenAddresses), it goes to the else branch and calls TokensBalance
	expectedError := errors.New("balance scanner error")
	mockBalanceScanner.EXPECT().TokensBalance(gomock.Any(), accountAddresses[0], tokenAddresses).Return(nil, expectedError)

	// Test
	result, err := fetcher.FetchErc20BalancesWithBalanceScanner(ctx, accountAddresses, tokenAddresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestFetchErc20BalancesWithBalanceScanner_WithLargeBatch(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	// Create more accounts than batch size to test chunking
	accountAddresses := make([]common.Address, 25)
	tokenAddresses := make([]common.Address, 2)
	for i := 0; i < 25; i++ {
		accountAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}
	for i := 0; i < 2; i++ {
		tokenAddresses[i] = common.HexToAddress(gofakeit.HexUint(160))
	}

	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10 // Should create 3 chunks: 10, 10, 5

	// Mock TokenBalances calls for each token with chunking
	for _, tokenAddress := range tokenAddresses {
		// First chunk: accounts[0:10]
		mockBalanceScanner.EXPECT().TokenBalances(gomock.Any(), accountAddresses[0:10], tokenAddress).DoAndReturn(
			func(opts *bind.CallOpts, addrs []common.Address, token common.Address) ([]balancescanner.BalanceScannerResult, error) {
				results := make([]balancescanner.BalanceScannerResult, len(addrs))
				for i := range addrs {
					balance := new(big.Int).Mul(big.NewInt(int64(i)), big.NewInt(1000000000000000000))
					results[i] = balancescanner.BalanceScannerResult{
						Success: true,
						Data:    balance.Bytes(),
					}
				}
				return results, nil
			})

		// Second chunk: accounts[10:20]
		mockBalanceScanner.EXPECT().TokenBalances(gomock.Any(), accountAddresses[10:20], tokenAddress).DoAndReturn(
			func(opts *bind.CallOpts, addrs []common.Address, token common.Address) ([]balancescanner.BalanceScannerResult, error) {
				results := make([]balancescanner.BalanceScannerResult, len(addrs))
				for i := range addrs {
					balance := new(big.Int).Mul(big.NewInt(int64(i+10)), big.NewInt(1000000000000000000))
					results[i] = balancescanner.BalanceScannerResult{
						Success: true,
						Data:    balance.Bytes(),
					}
				}
				return results, nil
			})

		// Third chunk: accounts[20:25]
		mockBalanceScanner.EXPECT().TokenBalances(gomock.Any(), accountAddresses[20:25], tokenAddress).DoAndReturn(
			func(opts *bind.CallOpts, addrs []common.Address, token common.Address) ([]balancescanner.BalanceScannerResult, error) {
				results := make([]balancescanner.BalanceScannerResult, len(addrs))
				for i := range addrs {
					balance := new(big.Int).Mul(big.NewInt(int64(i+20)), big.NewInt(1000000000000000000))
					results[i] = balancescanner.BalanceScannerResult{
						Success: true,
						Data:    balance.Bytes(),
					}
				}
				return results, nil
			})
	}

	// Test
	result, err := fetcher.FetchErc20BalancesWithBalanceScanner(ctx, accountAddresses, tokenAddresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 25)

	// Verify all addresses are present with correct balances
	for i, addr := range accountAddresses {
		assert.Contains(t, result, addr)
		assert.Len(t, result[addr], 2)
		for _, tokenAddr := range tokenAddresses {
			assert.Contains(t, result[addr], tokenAddr)
			expectedBalance := new(big.Int).Mul(big.NewInt(int64(i)), big.NewInt(1000000000000000000))
			assert.Equal(t, expectedBalance, result[addr][tokenAddr])
		}
	}
}

func TestFetchErc20BalancesWithBalanceScanner_FailedResults(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	accountAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate failed results
	mockBalanceScanner.EXPECT().TokenBalances(gomock.Any(), accountAddresses, tokenAddresses[0]).DoAndReturn(
		func(opts *bind.CallOpts, addrs []common.Address, token common.Address) ([]balancescanner.BalanceScannerResult, error) {
			results := []balancescanner.BalanceScannerResult{
				{
					Success: true,
					Data:    big.NewInt(1000000000000000000).Bytes(), // 1 token
				},
				{
					Success: false, // Failed result
					Data:    []byte{},
				},
			}
			return results, nil
		})

	// Test
	result, err := fetcher.FetchErc20BalancesWithBalanceScanner(ctx, accountAddresses, tokenAddresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	// First account should have balance, second should have zero balance due to failed result
	assert.Equal(t, big.NewInt(1000000000000000000), result[accountAddresses[0]][tokenAddresses[0]])
	assert.Equal(t, big.NewInt(0), result[accountAddresses[1]][tokenAddresses[0]])
}

func TestFetchErc20BalancesWithBalanceScanner_InvalidData(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockBalanceScanner := mock_balancescanner.NewMockBalanceScanner(ctrl)

	accountAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	tokenAddresses := []common.Address{
		common.HexToAddress(gofakeit.HexUint(160)),
	}
	atBlock := gethrpc.BlockNumber(1000)
	batchSize := 10

	// Mock expectations - simulate invalid data
	// Since len(accountAddresses) == len(tokenAddresses), it goes to the else branch and calls TokensBalance
	mockBalanceScanner.EXPECT().TokensBalance(gomock.Any(), accountAddresses[0], tokenAddresses).DoAndReturn(
		func(opts *bind.CallOpts, owner common.Address, contracts []common.Address) ([]balancescanner.BalanceScannerResult, error) {
			results := []balancescanner.BalanceScannerResult{
				{
					Success: true,
					Data:    []byte{0x01, 0x02, 0x03}, // Invalid data that can't be converted to big.Int
				},
			}
			return results, nil
		})

	// Test
	result, err := fetcher.FetchErc20BalancesWithBalanceScanner(ctx, accountAddresses, tokenAddresses, atBlock, mockBalanceScanner, batchSize)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	// The result should be a valid big.Int even with invalid data
	assert.NotNil(t, result[accountAddresses[0]][tokenAddresses[0]])
}
