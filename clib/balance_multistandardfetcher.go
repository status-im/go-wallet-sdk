package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"runtime/cgo"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/status-im/go-wallet-sdk/pkg/balance/multistandardfetcher"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

// internal type used for JSON marshaling/unmarshaling of multistandardfetcher.FetchConfig
type fetchConfigJSON struct {
	Native  []multistandardfetcher.AccountAddress            `json:"native"`
	ERC20   map[multistandardfetcher.AccountAddress][]string `json:"erc20"`
	ERC721  map[multistandardfetcher.AccountAddress][]string `json:"erc721"`
	ERC1155 map[multistandardfetcher.AccountAddress][]string `json:"erc1155"`
}

// internal type used for JSON marshaling/unmarshaling of multistandardfetcher.FetchResult
type fetchResultJSON struct {
	ResultType    multistandardfetcher.ResultType     `json:"resultType"`
	Account       multistandardfetcher.AccountAddress `json:"account"`
	Result        *hexutil.Big                        `json:"result,omitempty"`
	Results       map[string]*hexutil.Big             `json:"results,omitempty"`
	Err           *string                             `json:"err,omitempty"`
	AtBlockNumber *hexutil.Big                        `json:"atBlockNumber,omitempty"`
	AtBlockHash   hexutil.Bytes                       `json:"atBlockHash,omitempty"`
}

func collectibleIDToString(id multistandardfetcher.CollectibleID) string {
	return fmt.Sprintf("%s:%s", id.ContractAddress.String(), id.TokenID.String())
}

func hashableCollectibleIDToString(id multistandardfetcher.HashableCollectibleID) string {
	return collectibleIDToString(id.ToCollectibleID())
}

func collectibleIDFromString(s string) (multistandardfetcher.CollectibleID, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return multistandardfetcher.CollectibleID{}, fmt.Errorf("invalid collectible ID format")
	}
	if len(parts[0]) == 0 {
		return multistandardfetcher.CollectibleID{}, fmt.Errorf("contract address cannot be empty")
	}
	if len(parts[1]) == 0 {
		return multistandardfetcher.CollectibleID{}, fmt.Errorf("token ID cannot be empty")
	}
	contractAddress := common.HexToAddress(parts[0])
	tokenID, success := new(big.Int).SetString(parts[1], 10)
	if !success {
		return multistandardfetcher.CollectibleID{}, fmt.Errorf("invalid token ID format")
	}

	return multistandardfetcher.CollectibleID{
		ContractAddress: contractAddress,
		TokenID:         tokenID,
	}, nil
}

func hashableCollectibleIDFromString(s string) (multistandardfetcher.HashableCollectibleID, error) {
	cid, err := collectibleIDFromString(s)
	if err != nil {
		return multistandardfetcher.HashableCollectibleID{}, err
	}
	return cid.ToHashableCollectibleID(), nil
}

// errorToString converts an error to a string pointer for JSON serialization
func errorToString(err error) *string {
	if err == nil {
		return nil
	}
	errStr := err.Error()
	return &errStr
}

//export GoWSK_balance_multistandardfetcher_FetchBalances
func GoWSK_balance_multistandardfetcher_FetchBalances(ethClientHandle C.uintptr_t, chainIDC C.ulong, batchSizeC C.ulong, fetchConfigJSONCStr *C.char, cancelHandleOut *C.uintptr_t, errOut **C.char) *C.char {
	// Set up fetcher
	chainID := int64(chainIDC)

	multicallAddress, exists := multicall3.GetMulticall3Address(chainID)
	if !exists {
		handleError(errOut, errors.New("multicall3 not supported on chain"))
		return nil
	}

	h := cgo.Handle(ethClientHandle)
	c := castToEthClient(h)
	if c == nil {
		handleError(errOut, errors.New("invalid client handle"))
		return nil
	}

	multicallCaller, err := multicall3.NewMulticall3Caller(multicallAddress, c)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	// Get fetch configuration
	fetchConfigJSONTmp := fetchConfigJSON{}
	err = json.Unmarshal([]byte(C.GoString(fetchConfigJSONCStr)), &fetchConfigJSONTmp)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	fetchConfig := multistandardfetcher.FetchConfig{
		Native:  fetchConfigJSONTmp.Native,
		ERC20:   make(map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress),
		ERC721:  make(map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress),
		ERC1155: make(map[multistandardfetcher.AccountAddress][]multistandardfetcher.CollectibleID),
	}
	for account, erc20s := range fetchConfigJSONTmp.ERC20 {
		for _, erc20 := range erc20s {
			erc20Address := common.HexToAddress(erc20)
			fetchConfig.ERC20[account] = append(fetchConfig.ERC20[account], erc20Address)
		}
	}
	for account, erc721s := range fetchConfigJSONTmp.ERC721 {
		for _, erc721 := range erc721s {
			erc721Address := common.HexToAddress(erc721)
			fetchConfig.ERC721[account] = append(fetchConfig.ERC721[account], erc721Address)
		}
	}
	for account, erc1155s := range fetchConfigJSONTmp.ERC1155 {
		for _, erc1155 := range erc1155s {
			erc1155ID, err := collectibleIDFromString(erc1155)
			if err != nil {
				handleError(errOut, err)
				return nil
			}
			fetchConfig.ERC1155[account] = append(fetchConfig.ERC1155[account], erc1155ID)
		}
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Store the cancel function in a cgo.Handle and return it to C
	cancelHandle := cgo.NewHandle(cancel)
	if cancelHandleOut != nil {
		*cancelHandleOut = C.uintptr_t(cancelHandle)
	}

	// Call fetcher and collect results
	ch := multistandardfetcher.FetchBalances(ctx, multicallAddress, multicallCaller, fetchConfig, int(batchSizeC))
	results := make([]fetchResultJSON, 0)

	// Collect results, checking for context cancellation
resultLoop:
	for {
		select {
		case result, ok := <-ch:
			if !ok {
				// Channel closed, all results received
				break resultLoop
			}
			resJSON := fetchResultJSON{
				ResultType: result.ResultType,
			}
			switch result.ResultType {
			case multistandardfetcher.ResultTypeNative:
				nativeResult := result.Result.(multistandardfetcher.NativeResult)
				resJSON.Account = nativeResult.Account
				resJSON.Result = (*hexutil.Big)(nativeResult.Result)
				resJSON.Err = errorToString(nativeResult.Err)
				resJSON.AtBlockNumber = (*hexutil.Big)(nativeResult.AtBlockNumber)
				resJSON.AtBlockHash = nativeResult.AtBlockHash[:]
			case multistandardfetcher.ResultTypeERC20:
				erc20Result := result.Result.(multistandardfetcher.ERC20Result)
				resJSON.Account = erc20Result.Account
				resJSON.Results = make(map[string]*hexutil.Big)
				for k, v := range erc20Result.Results {
					resJSON.Results[k.String()] = (*hexutil.Big)(v)
				}
				resJSON.Err = errorToString(erc20Result.Err)
				resJSON.AtBlockNumber = (*hexutil.Big)(erc20Result.AtBlockNumber)
				resJSON.AtBlockHash = erc20Result.AtBlockHash[:]
			case multistandardfetcher.ResultTypeERC721:
				erc721Result := result.Result.(multistandardfetcher.ERC721Result)
				resJSON.Account = erc721Result.Account
				resJSON.Results = make(map[string]*hexutil.Big)
				for k, v := range erc721Result.Results {
					resJSON.Results[k.String()] = (*hexutil.Big)(v)
				}
				resJSON.Err = errorToString(erc721Result.Err)
				resJSON.AtBlockNumber = (*hexutil.Big)(erc721Result.AtBlockNumber)
				resJSON.AtBlockHash = erc721Result.AtBlockHash[:]
			case multistandardfetcher.ResultTypeERC1155:
				erc1155Result := result.Result.(multistandardfetcher.ERC1155Result)
				resJSON.Account = erc1155Result.Account
				resJSON.Results = make(map[string]*hexutil.Big)
				for hashableID, balance := range erc1155Result.Results {
					resJSON.Results[hashableCollectibleIDToString(hashableID)] = (*hexutil.Big)(balance)
				}
				resJSON.Err = errorToString(erc1155Result.Err)
				resJSON.AtBlockNumber = (*hexutil.Big)(erc1155Result.AtBlockNumber)
				resJSON.AtBlockHash = erc1155Result.AtBlockHash[:]
			}
			results = append(results, resJSON)
		case <-ctx.Done():
			// Context was cancelled, return partial results
			break resultLoop
		}
	}

	// Convert results to JSON
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(string(resultsJSON))
}

//export GoWSK_balance_multistandardfetcher_CancelFetchBalances
func GoWSK_balance_multistandardfetcher_CancelFetchBalances(cancelHandle C.uintptr_t) {
	h := cgo.Handle(cancelHandle)
	cancel, ok := h.Value().(context.CancelFunc)
	if ok {
		cancel()
	}
}

//export GoWSK_balance_multistandardfetcher_FreeCancelHandle
func GoWSK_balance_multistandardfetcher_FreeCancelHandle(cancelHandle C.uintptr_t) {
	h := cgo.Handle(cancelHandle)
	h.Delete()
}
