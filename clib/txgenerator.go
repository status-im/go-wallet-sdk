package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

// JSON parameter structs for each transaction type
type baseTxParamsJSON struct {
	Nonce                uint64 `json:"nonce"`
	GasLimit             uint64 `json:"gasLimit"`
	ChainID              string `json:"chainID"`
	GasPrice             string `json:"gasPrice,omitempty"`
	MaxFeePerGas         string `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas,omitempty"`
}

type transferETHParamsJSON struct {
	baseTxParamsJSON
	To    string `json:"to"`
	Value string `json:"value"`
}

type transferERC20ParamsJSON struct {
	baseTxParamsJSON
	TokenAddress string `json:"tokenAddress"`
	To           string `json:"to"`
	Amount       string `json:"amount"`
}

type approveERC20ParamsJSON struct {
	baseTxParamsJSON
	TokenAddress string `json:"tokenAddress"`
	Spender      string `json:"spender"`
	Amount       string `json:"amount"`
}

type transferERC721ParamsJSON struct {
	baseTxParamsJSON
	TokenAddress string `json:"tokenAddress"`
	From         string `json:"from"`
	To           string `json:"to"`
	TokenID      string `json:"tokenID"`
}

type approveERC721ParamsJSON struct {
	baseTxParamsJSON
	TokenAddress string `json:"tokenAddress"`
	To           string `json:"to"`
	TokenID      string `json:"tokenID"`
}

type setApprovalForAllERC721ParamsJSON struct {
	baseTxParamsJSON
	TokenAddress string `json:"tokenAddress"`
	Operator     string `json:"operator"`
	Approved     bool   `json:"approved"`
}

type transferERC1155ParamsJSON struct {
	baseTxParamsJSON
	TokenAddress string `json:"tokenAddress"`
	From         string `json:"from"`
	To           string `json:"to"`
	TokenID      string `json:"tokenID"`
	Value        string `json:"value"`
}

type batchTransferERC1155ParamsJSON struct {
	baseTxParamsJSON
	TokenAddress string   `json:"tokenAddress"`
	From         string   `json:"from"`
	To           string   `json:"to"`
	TokenIDs     []string `json:"tokenIDs"`
	Values       []string `json:"values"`
}

type setApprovalForAllERC1155ParamsJSON struct {
	baseTxParamsJSON
	TokenAddress string `json:"tokenAddress"`
	Operator     string `json:"operator"`
	Approved     bool   `json:"approved"`
}

// Helper function to parse big.Int from string
func parseBigInt(s string) (*big.Int, error) {
	if s == "" {
		return nil, nil
	}
	bi := new(big.Int)
	_, ok := bi.SetString(s, 10)
	if !ok {
		return nil, errors.New("invalid number: " + s)
	}
	return bi, nil
}

// Helper function to convert baseTxParamsJSON to BaseTxParams
func convertBaseTxParams(params baseTxParamsJSON) (txgenerator.BaseTxParams, error) {
	chainID, err := parseBigInt(params.ChainID)
	if err != nil {
		return txgenerator.BaseTxParams{}, err
	}

	var gasPrice, maxFeePerGas, maxPriorityFeePerGas *big.Int
	if params.GasPrice != "" {
		gasPrice, err = parseBigInt(params.GasPrice)
		if err != nil {
			return txgenerator.BaseTxParams{}, err
		}
	}
	if params.MaxFeePerGas != "" {
		maxFeePerGas, err = parseBigInt(params.MaxFeePerGas)
		if err != nil {
			return txgenerator.BaseTxParams{}, err
		}
	}
	if params.MaxPriorityFeePerGas != "" {
		maxPriorityFeePerGas, err = parseBigInt(params.MaxPriorityFeePerGas)
		if err != nil {
			return txgenerator.BaseTxParams{}, err
		}
	}

	return txgenerator.BaseTxParams{
		Nonce:                params.Nonce,
		GasLimit:             params.GasLimit,
		ChainID:              chainID,
		GasPrice:             gasPrice,
		MaxFeePerGas:         maxFeePerGas,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
	}, nil
}

// Helper function to serialize transaction to JSON
func txToJSON(tx *types.Transaction) (string, error) {
	txJSON, err := tx.MarshalJSON()
	if err != nil {
		return "", err
	}
	return string(txJSON), nil
}

//export GoWSK_txgenerator_TransferETH
func GoWSK_txgenerator_TransferETH(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params transferETHParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	to := common.HexToAddress(params.To)
	value, err := parseBigInt(params.Value)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tx, err := txgenerator.TransferETH(txgenerator.TransferETHParams{
		BaseTxParams: baseParams,
		To:           to,
		Value:        value,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_TransferERC20
func GoWSK_txgenerator_TransferERC20(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params transferERC20ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	to := common.HexToAddress(params.To)
	amount, err := parseBigInt(params.Amount)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tx, err := txgenerator.TransferERC20(txgenerator.TransferERC20Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		To:           to,
		Amount:       amount,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_ApproveERC20
func GoWSK_txgenerator_ApproveERC20(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params approveERC20ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	spender := common.HexToAddress(params.Spender)
	amount, err := parseBigInt(params.Amount)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tx, err := txgenerator.ApproveERC20(txgenerator.ApproveERC20Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		Spender:      spender,
		Amount:       amount,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_TransferFromERC721
func GoWSK_txgenerator_TransferFromERC721(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params transferERC721ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	from := common.HexToAddress(params.From)
	to := common.HexToAddress(params.To)
	tokenID, err := parseBigInt(params.TokenID)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tx, err := txgenerator.TransferFromERC721(txgenerator.TransferERC721Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		From:         from,
		To:           to,
		TokenID:      tokenID,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_SafeTransferFromERC721
func GoWSK_txgenerator_SafeTransferFromERC721(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params transferERC721ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	from := common.HexToAddress(params.From)
	to := common.HexToAddress(params.To)
	tokenID, err := parseBigInt(params.TokenID)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tx, err := txgenerator.SafeTransferFromERC721(txgenerator.TransferERC721Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		From:         from,
		To:           to,
		TokenID:      tokenID,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_ApproveERC721
func GoWSK_txgenerator_ApproveERC721(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params approveERC721ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	to := common.HexToAddress(params.To)
	tokenID, err := parseBigInt(params.TokenID)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tx, err := txgenerator.ApproveERC721(txgenerator.ApproveERC721Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		To:           to,
		TokenID:      tokenID,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_SetApprovalForAllERC721
func GoWSK_txgenerator_SetApprovalForAllERC721(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params setApprovalForAllERC721ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	operator := common.HexToAddress(params.Operator)

	tx, err := txgenerator.SetApprovalForAllERC721(txgenerator.SetApprovalForAllERC721Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		Operator:     operator,
		Approved:     params.Approved,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_TransferERC1155
func GoWSK_txgenerator_TransferERC1155(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params transferERC1155ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	from := common.HexToAddress(params.From)
	to := common.HexToAddress(params.To)
	tokenID, err := parseBigInt(params.TokenID)
	if err != nil {
		handleError(errOut, err)
		return nil
	}
	value, err := parseBigInt(params.Value)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tx, err := txgenerator.TransferERC1155(txgenerator.TransferERC1155Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		From:         from,
		To:           to,
		TokenID:      tokenID,
		Value:        value,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_BatchTransferERC1155
func GoWSK_txgenerator_BatchTransferERC1155(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params batchTransferERC1155ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	from := common.HexToAddress(params.From)
	to := common.HexToAddress(params.To)

	// Parse token IDs
	tokenIDs := make([]*big.Int, 0, len(params.TokenIDs))
	for _, idStr := range params.TokenIDs {
		tokenID, err := parseBigInt(idStr)
		if err != nil {
			handleError(errOut, err)
			return nil
		}
		tokenIDs = append(tokenIDs, tokenID)
	}

	// Parse values
	values := make([]*big.Int, 0, len(params.Values))
	for _, valStr := range params.Values {
		value, err := parseBigInt(valStr)
		if err != nil {
			handleError(errOut, err)
			return nil
		}
		values = append(values, value)
	}

	tx, err := txgenerator.BatchTransferERC1155(txgenerator.BatchTransferERC1155Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		From:         from,
		To:           to,
		TokenIDs:     tokenIDs,
		Values:       values,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}

//export GoWSK_txgenerator_SetApprovalForAllERC1155
func GoWSK_txgenerator_SetApprovalForAllERC1155(paramsJSON *C.char, errOut **C.char) *C.char {
	if paramsJSON == nil {
		handleError(errOut, errors.New("paramsJSON is NULL"))
		return nil
	}

	var params setApprovalForAllERC1155ParamsJSON
	if err := json.Unmarshal([]byte(C.GoString(paramsJSON)), &params); err != nil {
		handleError(errOut, err)
		return nil
	}

	baseParams, err := convertBaseTxParams(params.baseTxParamsJSON)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	tokenAddress := common.HexToAddress(params.TokenAddress)
	operator := common.HexToAddress(params.Operator)

	tx, err := txgenerator.SetApprovalForAllERC1155(txgenerator.SetApprovalForAllERC1155Params{
		BaseTxParams: baseParams,
		TokenAddress: tokenAddress,
		Operator:     operator,
		Approved:     params.Approved,
	})
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	jsonStr, err := txToJSON(tx)
	if err != nil {
		handleError(errOut, err)
		return nil
	}

	return C.CString(jsonStr)
}
