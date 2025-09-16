package multistandardfetcher_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/status-im/go-wallet-sdk/pkg/balance/multistandardfetcher"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	"github.com/status-im/go-wallet-sdk/pkg/multicall"
	mock_multicall "github.com/status-im/go-wallet-sdk/pkg/multicall/mock"
)

func TestFetchBalances_NativeBalances_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	account2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	expectedBalance1 := big.NewInt(1000000000000000000) // 1 ETH
	expectedBalance2 := big.NewInt(2000000000000000000) // 2 ETH

	// Mock the multicall execution
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildNativeBalanceCall(account1, multicall3Addr),
		multicall.BuildNativeBalanceCall(account2, multicall3Addr),
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: expectedBalance1.Bytes()},
		{Success: true, ReturnData: expectedBalance2.Bytes()},
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

	// Create config
	config := multistandardfetcher.FetchConfig{
		Native: []multistandardfetcher.AccountAddress{account1, account2},
	}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 2)

	// Check first result
	result1 := results[0]
	assert.Equal(t, multistandardfetcher.ResultTypeNative, result1.ResultType)
	nativeResult1, ok := result1.Result.(multistandardfetcher.NativeResult)
	assert.True(t, ok)
	assert.Equal(t, account1, nativeResult1.Account)
	assert.Equal(t, expectedBalance1, nativeResult1.Result)
	assert.NoError(t, nativeResult1.Err)
	assert.Equal(t, expectedBlockNumber, nativeResult1.AtBlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), nativeResult1.AtBlockHash)

	// Check second result
	result2 := results[1]
	assert.Equal(t, multistandardfetcher.ResultTypeNative, result2.ResultType)
	nativeResult2, ok := result2.Result.(multistandardfetcher.NativeResult)
	assert.True(t, ok)
	assert.Equal(t, account2, nativeResult2.Account)
	assert.Equal(t, expectedBalance2, nativeResult2.Result)
	assert.NoError(t, nativeResult2.Err)
	assert.Equal(t, expectedBlockNumber, nativeResult2.AtBlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), nativeResult2.AtBlockHash)
}

func TestFetchBalances_NativeBalances_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")
	expectedError := errors.New("network error")

	// Mock the multicall execution to return error
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildNativeBalanceCall(account, multicall3Addr),
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		Return(nil, [32]byte{}, nil, expectedError)

	// Create config
	config := multistandardfetcher.FetchConfig{
		Native: []multistandardfetcher.AccountAddress{account},
	}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, multistandardfetcher.ResultTypeNative, result.ResultType)
	nativeResult, ok := result.Result.(multistandardfetcher.NativeResult)
	assert.True(t, ok)
	assert.Equal(t, account, nativeResult.Account)
	assert.Nil(t, nativeResult.Result)
	assert.Equal(t, expectedError, nativeResult.Err)
	assert.Nil(t, nativeResult.AtBlockNumber)
	assert.Equal(t, common.Hash{}, nativeResult.AtBlockHash)
}

func TestFetchBalances_ERC20Balances_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")
	token1 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	token2 := common.HexToAddress("0x4444444444444444444444444444444444444444")

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	expectedBalance1 := big.NewInt(1000000000000000000) // 1 token
	expectedBalance2 := big.NewInt(2000000000000000000) // 2 tokens

	// Mock the multicall execution
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildERC20BalanceCall(account, token1),
		multicall.BuildERC20BalanceCall(account, token2),
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: expectedBalance1.Bytes()},
		{Success: true, ReturnData: expectedBalance2.Bytes()},
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

	// Create config
	config := multistandardfetcher.FetchConfig{
		ERC20: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
			account: {token1, token2},
		},
	}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, multistandardfetcher.ResultTypeERC20, result.ResultType)
	erc20Result, ok := result.Result.(multistandardfetcher.ERC20Result)
	assert.True(t, ok)
	assert.Equal(t, account, erc20Result.Account)
	assert.NoError(t, erc20Result.Err)
	assert.Equal(t, expectedBlockNumber, erc20Result.AtBlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), erc20Result.AtBlockHash)

	// Check token balances
	assert.Len(t, erc20Result.Results, 2)
	assert.Equal(t, expectedBalance1, erc20Result.Results[token1])
	assert.Equal(t, expectedBalance2, erc20Result.Results[token2])
}

func TestFetchBalances_ERC20Balances_MultipleAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	account2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	token1 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	token2 := common.HexToAddress("0x4444444444444444444444444444444444444444")

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	// Mock the multicall execution - all calls in one batch
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildERC20BalanceCall(account1, token1),
		multicall.BuildERC20BalanceCall(account1, token2),
		multicall.BuildERC20BalanceCall(account2, token1),
		multicall.BuildERC20BalanceCall(account2, token2),
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: big.NewInt(1000).Bytes()}, // account1, token1
		{Success: true, ReturnData: big.NewInt(2000).Bytes()}, // account1, token2
		{Success: true, ReturnData: big.NewInt(3000).Bytes()}, // account2, token1
		{Success: true, ReturnData: big.NewInt(4000).Bytes()}, // account2, token2
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

	// Create config
	config := multistandardfetcher.FetchConfig{
		ERC20: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
			account1: {token1, token2},
			account2: {token1, token2},
		},
	}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 2)

	// Find results by account
	var account1Result, account2Result multistandardfetcher.FetchResult
	for _, result := range results {
		erc20Result := result.Result.(multistandardfetcher.ERC20Result)
		switch erc20Result.Account {
		case account1:
			account1Result = result
		case account2:
			account2Result = result
		}
	}

	// Check account1 result
	assert.Equal(t, multistandardfetcher.ResultTypeERC20, account1Result.ResultType)
	erc20Result1 := account1Result.Result.(multistandardfetcher.ERC20Result)
	assert.Equal(t, account1, erc20Result1.Account)
	assert.NoError(t, erc20Result1.Err)
	assert.Equal(t, big.NewInt(1000), erc20Result1.Results[token1])
	assert.Equal(t, big.NewInt(2000), erc20Result1.Results[token2])

	// Check account2 result
	assert.Equal(t, multistandardfetcher.ResultTypeERC20, account2Result.ResultType)
	erc20Result2 := account2Result.Result.(multistandardfetcher.ERC20Result)
	assert.Equal(t, account2, erc20Result2.Account)
	assert.NoError(t, erc20Result2.Err)
	assert.Equal(t, big.NewInt(3000), erc20Result2.Results[token1])
	assert.Equal(t, big.NewInt(4000), erc20Result2.Results[token2])
}

func TestFetchBalances_ERC721Balances_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")
	nft1 := common.HexToAddress("0x5555555555555555555555555555555555555555")
	nft2 := common.HexToAddress("0x6666666666666666666666666666666666666666")

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	expectedBalance1 := big.NewInt(5) // 5 NFTs
	expectedBalance2 := big.NewInt(3) // 3 NFTs

	// Mock the multicall execution
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildERC721BalanceCall(account, nft1),
		multicall.BuildERC721BalanceCall(account, nft2),
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: expectedBalance1.Bytes()},
		{Success: true, ReturnData: expectedBalance2.Bytes()},
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

	// Create config
	config := multistandardfetcher.FetchConfig{
		ERC721: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
			account: {nft1, nft2},
		},
	}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, multistandardfetcher.ResultTypeERC721, result.ResultType)
	erc721Result, ok := result.Result.(multistandardfetcher.ERC721Result)
	assert.True(t, ok)
	assert.Equal(t, account, erc721Result.Account)
	assert.NoError(t, erc721Result.Err)
	assert.Equal(t, expectedBlockNumber, erc721Result.AtBlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), erc721Result.AtBlockHash)

	// Check NFT balances
	assert.Len(t, erc721Result.Results, 2)
	assert.Equal(t, expectedBalance1, erc721Result.Results[nft1])
	assert.Equal(t, expectedBalance2, erc721Result.Results[nft2])
}

func TestFetchBalances_ERC1155Balances_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")
	contract1 := common.HexToAddress("0x7777777777777777777777777777777777777777")
	contract2 := common.HexToAddress("0x8888888888888888888888888888888888888888")
	tokenID1 := big.NewInt(1)
	tokenID2 := big.NewInt(2)

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	expectedBalance1 := big.NewInt(10) // 10 tokens
	expectedBalance2 := big.NewInt(20) // 20 tokens

	// Mock the multicall execution
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildERC1155BalanceCall(account, contract1, tokenID1),
		multicall.BuildERC1155BalanceCall(account, contract2, tokenID2),
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: expectedBalance1.Bytes()},
		{Success: true, ReturnData: expectedBalance2.Bytes()},
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

	// Create config
	config := multistandardfetcher.FetchConfig{
		ERC1155: map[multistandardfetcher.AccountAddress][]multistandardfetcher.CollectibleID{
			account: {
				{ContractAddress: contract1, TokenID: tokenID1},
				{ContractAddress: contract2, TokenID: tokenID2},
			},
		},
	}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, multistandardfetcher.ResultTypeERC1155, result.ResultType)
	erc1155Result, ok := result.Result.(multistandardfetcher.ERC1155Result)
	assert.True(t, ok)
	assert.Equal(t, account, erc1155Result.Account)
	assert.NoError(t, erc1155Result.Err)
	assert.Equal(t, expectedBlockNumber, erc1155Result.AtBlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), erc1155Result.AtBlockHash)

	// Check collectible balances
	assert.Len(t, erc1155Result.Results, 2)

	// Find the correct hashable collectible IDs
	var collectibleID1, collectibleID2 multistandardfetcher.HashableCollectibleID
	for hashableID := range erc1155Result.Results {
		switch hashableID.ContractAddress {
		case contract1:
			collectibleID1 = hashableID
		case contract2:
			collectibleID2 = hashableID
		}
	}

	assert.Equal(t, expectedBalance1, erc1155Result.Results[collectibleID1])
	assert.Equal(t, expectedBalance2, erc1155Result.Results[collectibleID2])
}

func TestFetchBalances_MixedBalanceTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	account2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	token1 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	nft1 := common.HexToAddress("0x5555555555555555555555555555555555555555")
	contract1 := common.HexToAddress("0x7777777777777777777777777777777777777777")
	tokenID1 := big.NewInt(1)

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	// Mock the multicall execution - all calls in one batch
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildNativeBalanceCall(account1, multicall3Addr),       // Native
		multicall.BuildERC20BalanceCall(account1, token1),                // ERC20
		multicall.BuildERC721BalanceCall(account2, nft1),                 // ERC721
		multicall.BuildERC1155BalanceCall(account2, contract1, tokenID1), // ERC1155
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: true, ReturnData: big.NewInt(1000000000000000000).Bytes()}, // Native balance
		{Success: true, ReturnData: big.NewInt(2000).Bytes()},                // ERC20 balance
		{Success: true, ReturnData: big.NewInt(5).Bytes()},                   // ERC721 balance
		{Success: true, ReturnData: big.NewInt(10).Bytes()},                  // ERC1155 balance
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

	// Create config with mixed balance types
	config := multistandardfetcher.FetchConfig{
		Native: []multistandardfetcher.AccountAddress{account1},
		ERC20: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
			account1: {token1},
		},
		ERC721: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
			account2: {nft1},
		},
		ERC1155: map[multistandardfetcher.AccountAddress][]multistandardfetcher.CollectibleID{
			account2: {
				{ContractAddress: contract1, TokenID: tokenID1},
			},
		},
	}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 4)

	// Find results by type
	var nativeResult, erc20Result, erc721Result, erc1155Result multistandardfetcher.FetchResult
	for _, result := range results {
		switch result.ResultType {
		case multistandardfetcher.ResultTypeNative:
			nativeResult = result
		case multistandardfetcher.ResultTypeERC20:
			erc20Result = result
		case multistandardfetcher.ResultTypeERC721:
			erc721Result = result
		case multistandardfetcher.ResultTypeERC1155:
			erc1155Result = result
		}
	}

	// Check native result
	assert.Equal(t, multistandardfetcher.ResultTypeNative, nativeResult.ResultType)
	native := nativeResult.Result.(multistandardfetcher.NativeResult)
	assert.Equal(t, account1, native.Account)
	assert.Equal(t, big.NewInt(1000000000000000000), native.Result)
	assert.NoError(t, native.Err)

	// Check ERC20 result
	assert.Equal(t, multistandardfetcher.ResultTypeERC20, erc20Result.ResultType)
	erc20 := erc20Result.Result.(multistandardfetcher.ERC20Result)
	assert.Equal(t, account1, erc20.Account)
	assert.Equal(t, big.NewInt(2000), erc20.Results[token1])
	assert.NoError(t, erc20.Err)

	// Check ERC721 result
	assert.Equal(t, multistandardfetcher.ResultTypeERC721, erc721Result.ResultType)
	erc721 := erc721Result.Result.(multistandardfetcher.ERC721Result)
	assert.Equal(t, account2, erc721.Account)
	assert.Equal(t, big.NewInt(5), erc721.Results[nft1])
	assert.NoError(t, erc721.Err)

	// Check ERC1155 result
	assert.Equal(t, multistandardfetcher.ResultTypeERC1155, erc1155Result.ResultType)
	erc1155 := erc1155Result.Result.(multistandardfetcher.ERC1155Result)
	assert.Equal(t, account2, erc1155.Account)
	assert.Len(t, erc1155.Results, 1)
	// Find the hashable collectible ID
	for hashableID, balance := range erc1155.Results {
		if hashableID.ContractAddress == contract1 {
			assert.Equal(t, big.NewInt(10), balance)
		}
	}
	assert.NoError(t, erc1155.Err)
}

func TestFetchBalances_EmptyConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create empty config
	config := multistandardfetcher.FetchConfig{}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify no results
	assert.Empty(t, results)
}

func TestFetchBalances_ContextCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Mock the multicall execution to return context cancellation error
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildNativeBalanceCall(account, multicall3Addr),
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		DoAndReturn(func(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error) {
			cancel() // Cancel the context during the call
			return nil, [32]byte{}, nil, context.Canceled
		})

	// Create config
	config := multistandardfetcher.FetchConfig{
		Native: []multistandardfetcher.AccountAddress{account},
	}

	// Execute
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, multistandardfetcher.ResultTypeNative, result.ResultType)
	nativeResult := result.Result.(multistandardfetcher.NativeResult)
	assert.Equal(t, account, nativeResult.Account)
	assert.Error(t, nativeResult.Err)
	assert.Contains(t, nativeResult.Err.Error(), "context canceled")
}

func TestFetchBalances_CallFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCaller := mock_multicall.NewMockCaller(ctrl)

	// Test data
	multicall3Addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	account := common.HexToAddress("0x1111111111111111111111111111111111111111")
	token := common.HexToAddress("0x3333333333333333333333333333333333333333")

	expectedBlockNumber := big.NewInt(12345)
	expectedBlockHash := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	// Mock the multicall execution with one failed call
	expectedCalls := []multicall3.IMulticall3Call{
		multicall.BuildERC20BalanceCall(account, token),
	}
	expectedResults := []multicall3.IMulticall3Result{
		{Success: false, ReturnData: []byte("call failed")}, // Failed call
	}

	mockCaller.EXPECT().
		ViewTryBlockAndAggregate(
			gomock.Any(),
			false,
			expectedCalls,
		).
		Return(expectedBlockNumber, expectedBlockHash, expectedResults, nil)

	// Create config
	config := multistandardfetcher.FetchConfig{
		ERC20: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
			account: {token},
		},
	}

	// Execute
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicall3Addr, mockCaller, config, 10)

	// Collect results
	var results []multistandardfetcher.FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	// Verify results
	assert.Len(t, results, 1)
	result := results[0]
	assert.Equal(t, multistandardfetcher.ResultTypeERC20, result.ResultType)
	erc20Result := result.Result.(multistandardfetcher.ERC20Result)
	assert.Equal(t, account, erc20Result.Account)
	assert.NoError(t, erc20Result.Err) // Job-level error should be nil
	assert.Equal(t, expectedBlockNumber, erc20Result.AtBlockNumber)
	assert.Equal(t, common.Hash(expectedBlockHash), erc20Result.AtBlockHash)

	// The failed call should not appear in results (it's skipped due to nil check in processERC20JobResult)
	assert.Empty(t, erc20Result.Results)
}
