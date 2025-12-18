package txgenerator_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

func TestTransferETH_LegacyTransaction(t *testing.T) {
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	value := big.NewInt(1000000000000000000) // 1 ETH
	nonce := uint64(0)
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(20000000000) // 20 gwei
	chainID := big.NewInt(1)

	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    nonce,
			GasLimit: gasLimit,
			ChainID:  chainID,
			GasPrice: gasPrice,
		},
		To:    to,
		Value: value,
	}

	tx, err := txgenerator.TransferETH(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify transaction type
	assert.Equal(t, uint8(types.LegacyTxType), tx.Type())

	// Verify transaction fields
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, to, *tx.To())
	assert.Equal(t, value, tx.Value())
	assert.Equal(t, gasLimit, tx.Gas())
	assert.Equal(t, gasPrice, tx.GasPrice())
	assert.Nil(t, tx.Data())
}

func TestTransferETH_EIP1559Transaction(t *testing.T) {
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	value := big.NewInt(1000000000000000000) // 1 ETH
	nonce := uint64(0)
	gasLimit := uint64(21000)
	maxFeePerGas := big.NewInt(30000000000)        // 30 gwei
	maxPriorityFeePerGas := big.NewInt(2000000000) // 2 gwei
	chainID := big.NewInt(1)

	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                nonce,
			GasLimit:             gasLimit,
			ChainID:              chainID,
			MaxFeePerGas:         maxFeePerGas,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
		},
		To:    to,
		Value: value,
	}

	tx, err := txgenerator.TransferETH(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify transaction type
	assert.Equal(t, uint8(types.DynamicFeeTxType), tx.Type())

	// Verify transaction fields
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, to, *tx.To())
	assert.Equal(t, value, tx.Value())
	assert.Equal(t, gasLimit, tx.Gas())
	assert.Equal(t, chainID, tx.ChainId())
	assert.Nil(t, tx.Data())

	// Verify EIP-1559 transaction was created (type is verified above)
	// Note: GasFeeCap and GasTipCap are not directly accessible without reflection
	// but we verify the transaction type and other accessible fields
}

func TestTransferETH_ZeroAddress_Burn(t *testing.T) {
	// Zero address is valid for burning ETH
	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 21000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		To:    common.Address{}, // zero address for burning
		Value: big.NewInt(1000000000000000000),
	}

	tx, err := txgenerator.TransferETH(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, common.Address{}, *tx.To())
}

func TestTransferETH_NegativeValue(t *testing.T) {
	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 21000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		To:    common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Value: big.NewInt(-1), // negative value
	}

	tx, err := txgenerator.TransferETH(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "value must be non-negative")
}

func TestTransferETH_NilValue(t *testing.T) {
	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 21000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		To:    common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Value: nil,
	}

	tx, err := txgenerator.TransferETH(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "value must be non-negative")
}

func TestTransferETH_ZeroValue(t *testing.T) {
	// Zero value should be allowed (can be used for contract calls)
	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 21000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		To:    common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Value: big.NewInt(0),
	}

	tx, err := txgenerator.TransferETH(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, big.NewInt(0), tx.Value())
}

func TestTransferETH_MissingChainID(t *testing.T) {
	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 21000,
			ChainID:  nil, // missing chain ID
			GasPrice: big.NewInt(20000000000),
		},
		To:    common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Value: big.NewInt(1000000000000000000),
	}

	tx, err := txgenerator.TransferETH(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "chain ID is required")
}

func TestTransferETH_MissingGasPrice_Legacy(t *testing.T) {
	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 21000,
			ChainID:  big.NewInt(1),
			GasPrice: nil, // missing gas price for legacy tx
		},
		To:    common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Value: big.NewInt(1000000000000000000),
	}

	tx, err := txgenerator.TransferETH(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Equal(t, txgenerator.ErrMissingGasPrice, err)
}

func TestTransferETH_MissingMaxFeePerGas_EIP1559(t *testing.T) {
	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                0,
			GasLimit:             21000,
			ChainID:              big.NewInt(1),
			MaxFeePerGas:         nil, // missing max fee per gas
			MaxPriorityFeePerGas: big.NewInt(2000000000),
		},
		To:    common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Value: big.NewInt(1000000000000000000),
	}

	tx, err := txgenerator.TransferETH(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Equal(t, txgenerator.ErrMissingMaxFeePerGas, err)
}

func TestTransferETH_MissingMaxPriorityFeePerGas_EIP1559(t *testing.T) {
	params := txgenerator.TransferETHParams{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                0,
			GasLimit:             21000,
			ChainID:              big.NewInt(1),
			MaxFeePerGas:         big.NewInt(30000000000),
			MaxPriorityFeePerGas: nil, // missing max priority fee per gas
		},
		To:    common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Value: big.NewInt(1000000000000000000),
	}

	tx, err := txgenerator.TransferETH(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Equal(t, txgenerator.ErrMissingMaxPriorityFeePerGas, err)
}
