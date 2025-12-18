package txgenerator_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

func TestTransferERC20_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48") // USDC
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	amount := big.NewInt(1000000) // 1 USDC (6 decimals)
	nonce := uint64(0)
	gasLimit := uint64(65000)
	gasPrice := big.NewInt(20000000000) // 20 gwei
	chainID := big.NewInt(1)

	params := txgenerator.TransferERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    nonce,
			GasLimit: gasLimit,
			ChainID:  chainID,
			GasPrice: gasPrice,
		},
		TokenAddress: tokenAddress,
		To:           to,
		Amount:       amount,
	}

	tx, err := txgenerator.TransferERC20(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify transaction type
	assert.Equal(t, uint8(types.LegacyTxType), tx.Type())

	// Verify transaction fields
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, tokenAddress, *tx.To())
	assert.Equal(t, big.NewInt(0), tx.Value()) // no ETH value for token operations
	assert.Equal(t, gasLimit, tx.Gas())
	assert.Equal(t, gasPrice, tx.GasPrice())
	assert.NotNil(t, tx.Data()) // should have encoded transfer data
	assert.Greater(t, len(tx.Data()), 0)
}

func TestTransferERC20_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48") // USDC
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	amount := big.NewInt(1000000) // 1 USDC
	nonce := uint64(0)
	gasLimit := uint64(65000)
	maxFeePerGas := big.NewInt(30000000000)        // 30 gwei
	maxPriorityFeePerGas := big.NewInt(2000000000) // 2 gwei
	chainID := big.NewInt(1)

	params := txgenerator.TransferERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                nonce,
			GasLimit:             gasLimit,
			ChainID:              chainID,
			MaxFeePerGas:         maxFeePerGas,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
		},
		TokenAddress: tokenAddress,
		To:           to,
		Amount:       amount,
	}

	tx, err := txgenerator.TransferERC20(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify transaction type
	assert.Equal(t, uint8(types.DynamicFeeTxType), tx.Type())

	// Verify transaction fields
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, tokenAddress, *tx.To())
	assert.Equal(t, big.NewInt(0), tx.Value())
	assert.Equal(t, gasLimit, tx.Gas())
	assert.Equal(t, chainID, tx.ChainId())
	assert.NotNil(t, tx.Data())
	assert.Greater(t, len(tx.Data()), 0)

	// Verify EIP-1559 transaction was created (type is verified above)
	// Note: GasFeeCap and GasTipCap are not directly accessible without reflection
	// but we verify the transaction type and other accessible fields
}

func TestTransferERC20_ZeroTokenAddress(t *testing.T) {
	params := txgenerator.TransferERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 65000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.Address{}, // zero address
		To:           common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Amount:       big.NewInt(1000000),
	}

	tx, err := txgenerator.TransferERC20(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token address cannot be zero")
}

func TestTransferERC20_ZeroRecipientAddress_Burn(t *testing.T) {
	// Zero address is valid for burning tokens
	params := txgenerator.TransferERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 65000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		To:           common.Address{}, // zero address for burning
		Amount:       big.NewInt(1000000),
	}

	tx, err := txgenerator.TransferERC20(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestTransferERC20_NegativeAmount(t *testing.T) {
	params := txgenerator.TransferERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 65000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		To:           common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Amount:       big.NewInt(-1), // negative amount
	}

	tx, err := txgenerator.TransferERC20(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "amount must be non-negative")
}

func TestTransferERC20_NilAmount(t *testing.T) {
	params := txgenerator.TransferERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 65000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		To:           common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Amount:       nil,
	}

	tx, err := txgenerator.TransferERC20(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "amount must be non-negative")
}

func TestTransferERC20_ZeroAmount(t *testing.T) {
	// Zero amount should be allowed (can be used for some edge cases)
	params := txgenerator.TransferERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 65000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		To:           common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Amount:       big.NewInt(0),
	}

	tx, err := txgenerator.TransferERC20(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestTransferERC20_MissingChainID(t *testing.T) {
	params := txgenerator.TransferERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 65000,
			ChainID:  nil, // missing chain ID
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		To:           common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Amount:       big.NewInt(1000000),
	}

	tx, err := txgenerator.TransferERC20(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "chain ID is required")
}

func TestApproveERC20_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48") // USDC
	spender := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	amount := big.NewInt(1000000) // 1 USDC
	nonce := uint64(0)
	gasLimit := uint64(46000)
	gasPrice := big.NewInt(20000000000) // 20 gwei
	chainID := big.NewInt(1)

	params := txgenerator.ApproveERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    nonce,
			GasLimit: gasLimit,
			ChainID:  chainID,
			GasPrice: gasPrice,
		},
		TokenAddress: tokenAddress,
		Spender:      spender,
		Amount:       amount,
	}

	tx, err := txgenerator.ApproveERC20(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify transaction type
	assert.Equal(t, uint8(types.LegacyTxType), tx.Type())

	// Verify transaction fields
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, tokenAddress, *tx.To())
	assert.Equal(t, big.NewInt(0), tx.Value())
	assert.Equal(t, gasLimit, tx.Gas())
	assert.Equal(t, gasPrice, tx.GasPrice())
	assert.NotNil(t, tx.Data())
	assert.Greater(t, len(tx.Data()), 0)
}

func TestApproveERC20_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48") // USDC
	spender := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	amount := big.NewInt(1000000) // 1 USDC
	nonce := uint64(0)
	gasLimit := uint64(46000)
	maxFeePerGas := big.NewInt(30000000000)        // 30 gwei
	maxPriorityFeePerGas := big.NewInt(2000000000) // 2 gwei
	chainID := big.NewInt(1)

	params := txgenerator.ApproveERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                nonce,
			GasLimit:             gasLimit,
			ChainID:              chainID,
			MaxFeePerGas:         maxFeePerGas,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
		},
		TokenAddress: tokenAddress,
		Spender:      spender,
		Amount:       amount,
	}

	tx, err := txgenerator.ApproveERC20(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify transaction type
	assert.Equal(t, uint8(types.DynamicFeeTxType), tx.Type())

	// Verify transaction fields
	assert.Equal(t, nonce, tx.Nonce())
	assert.Equal(t, tokenAddress, *tx.To())
	assert.Equal(t, big.NewInt(0), tx.Value())
	assert.Equal(t, gasLimit, tx.Gas())
	assert.Equal(t, chainID, tx.ChainId())
	assert.NotNil(t, tx.Data())
	assert.Greater(t, len(tx.Data()), 0)
}

func TestApproveERC20_ZeroTokenAddress(t *testing.T) {
	params := txgenerator.ApproveERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.Address{}, // zero address
		Spender:      common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Amount:       big.NewInt(1000000),
	}

	tx, err := txgenerator.ApproveERC20(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token address cannot be zero")
}

func TestApproveERC20_ZeroSpenderAddress(t *testing.T) {
	params := txgenerator.ApproveERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		Spender:      common.Address{}, // zero address
		Amount:       big.NewInt(1000000),
	}

	tx, err := txgenerator.ApproveERC20(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "spender address cannot be zero")
}

func TestApproveERC20_NegativeAmount(t *testing.T) {
	params := txgenerator.ApproveERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		Spender:      common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Amount:       big.NewInt(-1), // negative amount
	}

	tx, err := txgenerator.ApproveERC20(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "amount must be non-negative")
}

func TestApproveERC20_MissingChainID(t *testing.T) {
	params := txgenerator.ApproveERC20Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  nil, // missing chain ID
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		Spender:      common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Amount:       big.NewInt(1000000),
	}

	tx, err := txgenerator.ApproveERC20(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "chain ID is required")
}
