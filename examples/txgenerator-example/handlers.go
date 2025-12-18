package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

// TransactionRequest represents the request from the frontend
type TransactionRequest struct {
	TxType               string            `json:"txType"`
	UseEIP1559           bool              `json:"useEIP1559"`
	Nonce                string            `json:"nonce"`
	GasLimit             string            `json:"gasLimit"`
	ChainID              string            `json:"chainID"`
	GasPrice             string            `json:"gasPrice"`
	MaxFeePerGas         string            `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string            `json:"maxPriorityFeePerGas"`
	Params               map[string]string `json:"params"`
}

// GenerateTransaction generates a transaction based on the request
func GenerateTransaction(req TransactionRequest) (*types.Transaction, error) {
	// Parse common parameters
	nonce, err := parseUint64(req.Nonce)
	if err != nil {
		return nil, err
	}

	gasLimit, err := parseUint64(req.GasLimit)
	if err != nil {
		return nil, err
	}

	chainID, err := parseBigInt(req.ChainID)
	if err != nil {
		return nil, err
	}

	// Parse fee parameters
	var gasPrice, maxFeePerGas, maxPriorityFeePerGas *big.Int
	if req.UseEIP1559 {
		maxFeePerGas, err = parseBigInt(req.MaxFeePerGas)
		if err != nil {
			return nil, err
		}
		maxPriorityFeePerGas, err = parseBigInt(req.MaxPriorityFeePerGas)
		if err != nil {
			return nil, err
		}
	} else {
		gasPrice, err = parseBigInt(req.GasPrice)
		if err != nil {
			return nil, err
		}
	}

	// Generate transaction based on type
	switch req.TxType {
	case "transferETH":
		to := common.HexToAddress(req.Params["to"])
		value, err := parseBigInt(req.Params["value"])
		if err != nil {
			return nil, err
		}
		return txgenerator.TransferETH(txgenerator.TransferETHParams{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			To:    to,
			Value: value,
		})

	case "transferERC20":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		to := common.HexToAddress(req.Params["to"])
		amount, err := parseBigInt(req.Params["amount"])
		if err != nil {
			return nil, err
		}
		return txgenerator.TransferERC20(txgenerator.TransferERC20Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			To:           to,
			Amount:       amount,
		})

	case "approveERC20":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		spender := common.HexToAddress(req.Params["spender"])
		amount, err := parseBigInt(req.Params["amount"])
		if err != nil {
			return nil, err
		}
		return txgenerator.ApproveERC20(txgenerator.ApproveERC20Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			Spender:      spender,
			Amount:       amount,
		})

	case "transferFromERC721":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		from := common.HexToAddress(req.Params["from"])
		to := common.HexToAddress(req.Params["to"])
		tokenID, err := parseBigInt(req.Params["tokenID"])
		if err != nil {
			return nil, err
		}
		return txgenerator.TransferFromERC721(txgenerator.TransferERC721Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			From:         from,
			To:           to,
			TokenID:      tokenID,
		})

	case "safeTransferFromERC721":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		from := common.HexToAddress(req.Params["from"])
		to := common.HexToAddress(req.Params["to"])
		tokenID, err := parseBigInt(req.Params["tokenID"])
		if err != nil {
			return nil, err
		}
		return txgenerator.SafeTransferFromERC721(txgenerator.TransferERC721Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			From:         from,
			To:           to,
			TokenID:      tokenID,
		})

	case "approveERC721":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		to := common.HexToAddress(req.Params["to"])
		tokenID, err := parseBigInt(req.Params["tokenID"])
		if err != nil {
			return nil, err
		}
		return txgenerator.ApproveERC721(txgenerator.ApproveERC721Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			To:           to,
			TokenID:      tokenID,
		})

	case "setApprovalForAllERC721":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		operator := common.HexToAddress(req.Params["operator"])
		approved := strings.ToLower(req.Params["approved"]) == "true"
		return txgenerator.SetApprovalForAllERC721(txgenerator.SetApprovalForAllERC721Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			Operator:     operator,
			Approved:     approved,
		})

	case "transferERC1155":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		from := common.HexToAddress(req.Params["from"])
		to := common.HexToAddress(req.Params["to"])
		tokenID, err := parseBigInt(req.Params["tokenID"])
		if err != nil {
			return nil, err
		}
		value, err := parseBigInt(req.Params["value"])
		if err != nil {
			return nil, err
		}
		return txgenerator.TransferERC1155(txgenerator.TransferERC1155Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			From:         from,
			To:           to,
			TokenID:      tokenID,
			Value:        value,
		})

	case "batchTransferERC1155":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		from := common.HexToAddress(req.Params["from"])
		to := common.HexToAddress(req.Params["to"])

		// Parse tokenIDs (comma-separated)
		tokenIDStrs := strings.Split(req.Params["tokenIDs"], ",")
		tokenIDs := make([]*big.Int, 0, len(tokenIDStrs))
		for _, idStr := range tokenIDStrs {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}
			tokenID, err := parseBigInt(idStr)
			if err != nil {
				return nil, err
			}
			tokenIDs = append(tokenIDs, tokenID)
		}

		// Parse values (comma-separated)
		valueStrs := strings.Split(req.Params["values"], ",")
		values := make([]*big.Int, 0, len(valueStrs))
		for _, valStr := range valueStrs {
			valStr = strings.TrimSpace(valStr)
			if valStr == "" {
				continue
			}
			value, err := parseBigInt(valStr)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}

		return txgenerator.BatchTransferERC1155(txgenerator.BatchTransferERC1155Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			From:         from,
			To:           to,
			TokenIDs:     tokenIDs,
			Values:       values,
		})

	case "setApprovalForAllERC1155":
		tokenAddress := common.HexToAddress(req.Params["tokenAddress"])
		operator := common.HexToAddress(req.Params["operator"])
		approved := strings.ToLower(req.Params["approved"]) == "true"
		return txgenerator.SetApprovalForAllERC1155(txgenerator.SetApprovalForAllERC1155Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:                nonce,
				GasLimit:             gasLimit,
				ChainID:              chainID,
				GasPrice:             gasPrice,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: maxPriorityFeePerGas,
			},
			TokenAddress: tokenAddress,
			Operator:     operator,
			Approved:     approved,
		})

	default:
		return nil, fmt.Errorf("unsupported transaction type: %s", req.TxType)
	}
}

// TransactionToJSON converts a transaction to JSON format using the built-in MarshalJSON
func TransactionToJSON(tx *types.Transaction) (map[string]interface{}, error) {
	// Use the built-in MarshalJSON method
	txJSON, err := tx.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// Parse the JSON to add the raw RLP bytes
	var txMap map[string]interface{}
	if err := json.Unmarshal(txJSON, &txMap); err != nil {
		return nil, err
	}

	// Get raw transaction bytes (RLP encoded)
	rawBytes, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	txMap["raw"] = "0x" + common.Bytes2Hex(rawBytes)

	return txMap, nil
}

// Helper functions
func parseUint64(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

func parseBigInt(s string) (*big.Int, error) {
	if s == "" {
		return nil, nil
	}
	bi := new(big.Int)
	_, ok := bi.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid number: %s", s)
	}
	return bi, nil
}
