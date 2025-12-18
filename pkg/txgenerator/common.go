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

func validateAndDetermineTxType(params BaseTxParams) (TxType, error) {

	if params.MaxFeePerGas != nil || params.MaxPriorityFeePerGas != nil {
		if params.ChainID == nil {
			return 0, ErrMissingChainID
		}
		if params.MaxFeePerGas == nil {
			return 0, ErrMissingMaxFeePerGas
		}
		if params.MaxPriorityFeePerGas == nil {
			return 0, ErrMissingMaxPriorityFeePerGas
		}
		if params.MaxFeePerGas.Sign() < 0 {
			return 0, ErrNegativeMaxFeePerGas
		}
		if params.MaxPriorityFeePerGas.Sign() < 0 {
			return 0, ErrNegativeMaxPriorityFeePerGas
		}
		return types.DynamicFeeTxType, nil
	}

	if params.GasPrice == nil {
		return 0, ErrMissingGasPrice
	}
	if params.GasPrice.Sign() < 0 {
		return 0, ErrNegativeGasPrice
	}
	return types.LegacyTxType, nil
}

// createTokenTransaction is a helper function to create a token transaction (ERC20/ERC721/ERC1155)
// with the appropriate transaction type (legacy or EIP-1559)
func createTokenTransaction(
	baseParams BaseTxParams,
	tokenAddress common.Address,
	data []byte,
) (*types.Transaction, error) {
	txType, err := validateAndDetermineTxType(baseParams)
	if err != nil {
		return nil, err
	}

	switch txType {
	case types.LegacyTxType:
		return types.NewTx(&types.LegacyTx{
			Nonce:    baseParams.Nonce,
			To:       &tokenAddress,
			Value:    big.NewInt(0), // no ETH value for token operations
			Gas:      baseParams.GasLimit,
			GasPrice: baseParams.GasPrice,
			Data:     data,
		}), nil

	case types.DynamicFeeTxType:
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   baseParams.ChainID,
			Nonce:     baseParams.Nonce,
			GasTipCap: baseParams.MaxPriorityFeePerGas,
			GasFeeCap: baseParams.MaxFeePerGas,
			Gas:       baseParams.GasLimit,
			To:        &tokenAddress,
			Value:     big.NewInt(0), // no ETH value for token operations
			Data:      data,
		}), nil

	default:
		return nil, errors.New("unsupported transaction type")
	}
}
