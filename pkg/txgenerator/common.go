package txgenerator

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TxType represents the type of transaction to create
type TxType = byte

// BaseTxParams contains common parameters for all transaction types
type BaseTxParams struct {
	// Nonce is the transaction nonce
	Nonce uint64
	// GasLimit is the gas limit for the transaction
	GasLimit uint64
	// ChainID is the chain ID for the transaction
	ChainID *big.Int

	// For legacy transactions (TxType = LegacyTxType):
	// GasPrice is the gas price in wei
	GasPrice *big.Int

	// For EIP-1559 transactions (TxType = DynamicFeeTxType):
	// MaxFeePerGas is the maximum fee per gas in wei
	MaxFeePerGas *big.Int
	// MaxPriorityFeePerGas is the maximum priority fee per gas in wei
	MaxPriorityFeePerGas *big.Int
}

// createTokenTransaction is a helper function to create a token transaction (ERC20/ERC721/ERC1155)
// with the appropriate transaction type (legacy or EIP-1559)
func createTokenTransaction(
	tokenAddress common.Address,
	nonce uint64,
	gasLimit uint64,
	chainID *big.Int,
	gasPrice *big.Int,
	maxFeePerGas *big.Int,
	maxPriorityFeePerGas *big.Int,
	data []byte,
) (*types.Transaction, error) {
	// Determine transaction type based on provided fields
	var txType TxType
	if maxFeePerGas != nil || maxPriorityFeePerGas != nil {
		txType = types.DynamicFeeTxType
		if maxFeePerGas == nil {
			return nil, ErrMissingMaxFeePerGas
		}
		if maxPriorityFeePerGas == nil {
			return nil, ErrMissingMaxPriorityFeePerGas
		}
	} else {
		txType = types.LegacyTxType
		if gasPrice == nil {
			return nil, ErrMissingGasPrice
		}
	}

	switch txType {
	case types.LegacyTxType:
		return types.NewTransaction(
			nonce,
			tokenAddress,
			big.NewInt(0), // no ETH value for token operations
			gasLimit,
			gasPrice,
			data,
		), nil

	case types.DynamicFeeTxType:
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			GasTipCap: maxPriorityFeePerGas,
			GasFeeCap: maxFeePerGas,
			Gas:       gasLimit,
			To:        &tokenAddress,
			Value:     big.NewInt(0), // no ETH value for token operations
			Data:      data,
		}), nil

	default:
		return nil, errors.New("unsupported transaction type")
	}
}
