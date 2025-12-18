package txgenerator

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TransferETHParams contains parameters for generating an ETH transfer transaction
type TransferETHParams struct {
	BaseTxParams
	// To is the recipient address
	To common.Address
	// Value is the amount to transfer in wei
	Value *big.Int
}

// TransferETH generates an unsigned transaction for transferring ETH
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func TransferETH(params TransferETHParams) (*types.Transaction, error) {
	if params.Value == nil || params.Value.Sign() < 0 {
		return nil, errors.New("value must be non-negative")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Determine transaction type based on provided fields
	var txType TxType
	if params.MaxFeePerGas != nil || params.MaxPriorityFeePerGas != nil {
		txType = types.DynamicFeeTxType
		if params.MaxFeePerGas == nil {
			return nil, ErrMissingMaxFeePerGas
		}
		if params.MaxPriorityFeePerGas == nil {
			return nil, ErrMissingMaxPriorityFeePerGas
		}
	} else {
		txType = types.LegacyTxType
		if params.GasPrice == nil {
			return nil, ErrMissingGasPrice
		}
	}

	switch txType {
	case types.LegacyTxType:
		return types.NewTransaction(
			params.Nonce,
			params.To,
			params.Value,
			params.GasLimit,
			params.GasPrice,
			nil, // no data for ETH transfer
		), nil

	case types.DynamicFeeTxType:
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   params.ChainID,
			Nonce:     params.Nonce,
			GasTipCap: params.MaxPriorityFeePerGas,
			GasFeeCap: params.MaxFeePerGas,
			Gas:       params.GasLimit,
			To:        &params.To,
			Value:     params.Value,
			Data:      nil, // no data for ETH transfer
		}), nil

	default:
		return nil, errors.New("unsupported transaction type")
	}
}
