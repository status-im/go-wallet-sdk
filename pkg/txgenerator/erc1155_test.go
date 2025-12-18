package txgenerator_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

func TestTransferERC1155_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72")
	tokenID := big.NewInt(1)
	value := big.NewInt(100)
	nonce := uint64(0)
	gasLimit := uint64(100000)
	gasPrice := big.NewInt(20000000000) // 20 gwei
	chainID := big.NewInt(1)

	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    nonce,
			GasLimit: gasLimit,
			ChainID:  chainID,
			GasPrice: gasPrice,
		},
		TokenAddress: tokenAddress,
		From:         from,
		To:           to,
		TokenID:      tokenID,
		Value:        value,
	}

	tx, err := txgenerator.TransferERC1155(params)
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

func TestTransferERC1155_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72")
	tokenID := big.NewInt(1)
	value := big.NewInt(100)
	nonce := uint64(0)
	gasLimit := uint64(100000)
	maxFeePerGas := big.NewInt(30000000000)        // 30 gwei
	maxPriorityFeePerGas := big.NewInt(2000000000) // 2 gwei
	chainID := big.NewInt(1)

	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                nonce,
			GasLimit:             gasLimit,
			ChainID:              chainID,
			MaxFeePerGas:         maxFeePerGas,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
		},
		TokenAddress: tokenAddress,
		From:         from,
		To:           to,
		TokenID:      tokenID,
		Value:        value,
	}

	tx, err := txgenerator.TransferERC1155(params)
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

func TestTransferERC1155_ZeroTokenAddress(t *testing.T) {
	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.Address{}, // zero address
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(1),
		Value:        big.NewInt(100),
	}

	tx, err := txgenerator.TransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token address cannot be zero")
}

func TestTransferERC1155_ZeroFromAddress(t *testing.T) {
	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.Address{}, // zero address
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(1),
		Value:        big.NewInt(100),
	}

	tx, err := txgenerator.TransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "from address cannot be zero")
}

func TestTransferERC1155_ZeroToAddress_Burn(t *testing.T) {
	// Zero address is valid for burning tokens
	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.Address{}, // zero address for burning
		TokenID:      big.NewInt(1),
		Value:        big.NewInt(100),
	}

	tx, err := txgenerator.TransferERC1155(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestTransferERC1155_NegativeTokenID(t *testing.T) {
	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(-1), // negative token ID
		Value:        big.NewInt(100),
	}

	tx, err := txgenerator.TransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token ID must be non-negative")
}

func TestTransferERC1155_NegativeValue(t *testing.T) {
	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(1),
		Value:        big.NewInt(-1), // negative value
	}

	tx, err := txgenerator.TransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "value must be non-negative")
}

func TestTransferERC1155_ZeroValue(t *testing.T) {
	// Zero value should be allowed
	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(1),
		Value:        big.NewInt(0),
	}

	tx, err := txgenerator.TransferERC1155(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestTransferERC1155_MissingChainID(t *testing.T) {
	params := txgenerator.TransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  nil, // missing chain ID
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(1),
		Value:        big.NewInt(100),
	}

	tx, err := txgenerator.TransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "chain ID is required")
}

func TestBatchTransferERC1155_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72")
	tokenIDs := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	values := []*big.Int{big.NewInt(10), big.NewInt(20), big.NewInt(30)}
	nonce := uint64(0)
	gasLimit := uint64(200000)
	gasPrice := big.NewInt(20000000000)
	chainID := big.NewInt(1)

	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    nonce,
			GasLimit: gasLimit,
			ChainID:  chainID,
			GasPrice: gasPrice,
		},
		TokenAddress: tokenAddress,
		From:         from,
		To:           to,
		TokenIDs:     tokenIDs,
		Values:       values,
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
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

func TestBatchTransferERC1155_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72")
	tokenIDs := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	values := []*big.Int{big.NewInt(10), big.NewInt(20), big.NewInt(30)}
	nonce := uint64(0)
	gasLimit := uint64(200000)
	maxFeePerGas := big.NewInt(30000000000)
	maxPriorityFeePerGas := big.NewInt(2000000000)
	chainID := big.NewInt(1)

	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                nonce,
			GasLimit:             gasLimit,
			ChainID:              chainID,
			MaxFeePerGas:         maxFeePerGas,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
		},
		TokenAddress: tokenAddress,
		From:         from,
		To:           to,
		TokenIDs:     tokenIDs,
		Values:       values,
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
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

func TestBatchTransferERC1155_EmptyTokenIDs(t *testing.T) {
	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 200000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenIDs:     []*big.Int{}, // empty token IDs
		Values:       []*big.Int{big.NewInt(10)},
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token IDs cannot be empty")
}

func TestBatchTransferERC1155_EmptyValues(t *testing.T) {
	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 200000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenIDs:     []*big.Int{big.NewInt(1)},
		Values:       []*big.Int{}, // empty values
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "values cannot be empty")
}

func TestBatchTransferERC1155_MismatchedArrayLengths(t *testing.T) {
	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 200000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenIDs:     []*big.Int{big.NewInt(1), big.NewInt(2)},
		Values:       []*big.Int{big.NewInt(10)}, // different length
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token IDs and values arrays must have the same length")
}

func TestBatchTransferERC1155_NegativeTokenID(t *testing.T) {
	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 200000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenIDs:     []*big.Int{big.NewInt(-1)}, // negative token ID
		Values:       []*big.Int{big.NewInt(10)},
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token ID must be non-negative")
}

func TestBatchTransferERC1155_NegativeValue(t *testing.T) {
	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 200000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenIDs:     []*big.Int{big.NewInt(1)},
		Values:       []*big.Int{big.NewInt(-1)}, // negative value
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "value must be non-negative")
}

func TestBatchTransferERC1155_NilTokenID(t *testing.T) {
	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 200000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenIDs:     []*big.Int{nil}, // nil token ID
		Values:       []*big.Int{big.NewInt(10)},
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token ID must be non-negative")
}

func TestBatchTransferERC1155_NilValue(t *testing.T) {
	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 200000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenIDs:     []*big.Int{big.NewInt(1)},
		Values:       []*big.Int{nil}, // nil value
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "value must be non-negative")
}

func TestBatchTransferERC1155_ZeroToAddress_Burn(t *testing.T) {
	// Zero address is valid for burning tokens
	params := txgenerator.BatchTransferERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 200000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.Address{}, // zero address for burning
		TokenIDs:     []*big.Int{big.NewInt(1), big.NewInt(2)},
		Values:       []*big.Int{big.NewInt(10), big.NewInt(20)},
	}

	tx, err := txgenerator.BatchTransferERC1155(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestSetApprovalForAllERC1155_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	operator := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	approved := true
	nonce := uint64(0)
	gasLimit := uint64(46000)
	gasPrice := big.NewInt(20000000000)
	chainID := big.NewInt(1)

	params := txgenerator.SetApprovalForAllERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    nonce,
			GasLimit: gasLimit,
			ChainID:  chainID,
			GasPrice: gasPrice,
		},
		TokenAddress: tokenAddress,
		Operator:     operator,
		Approved:     approved,
	}

	tx, err := txgenerator.SetApprovalForAllERC1155(params)
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

func TestSetApprovalForAllERC1155_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	operator := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	approved := true
	nonce := uint64(0)
	gasLimit := uint64(46000)
	maxFeePerGas := big.NewInt(30000000000)
	maxPriorityFeePerGas := big.NewInt(2000000000)
	chainID := big.NewInt(1)

	params := txgenerator.SetApprovalForAllERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                nonce,
			GasLimit:             gasLimit,
			ChainID:              chainID,
			MaxFeePerGas:         maxFeePerGas,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
		},
		TokenAddress: tokenAddress,
		Operator:     operator,
		Approved:     approved,
	}

	tx, err := txgenerator.SetApprovalForAllERC1155(params)
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

func TestSetApprovalForAllERC1155_RevokeApproval(t *testing.T) {
	// Test revoking approval (approved = false)
	tokenAddress := common.HexToAddress("0x1234567890123456789012345678901234567890")
	operator := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	approved := false
	nonce := uint64(0)
	gasLimit := uint64(46000)
	gasPrice := big.NewInt(20000000000)
	chainID := big.NewInt(1)

	params := txgenerator.SetApprovalForAllERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    nonce,
			GasLimit: gasLimit,
			ChainID:  chainID,
			GasPrice: gasPrice,
		},
		TokenAddress: tokenAddress,
		Operator:     operator,
		Approved:     approved,
	}

	tx, err := txgenerator.SetApprovalForAllERC1155(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestSetApprovalForAllERC1155_ZeroTokenAddress(t *testing.T) {
	params := txgenerator.SetApprovalForAllERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.Address{}, // zero address
		Operator:     common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Approved:     true,
	}

	tx, err := txgenerator.SetApprovalForAllERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token address cannot be zero")
}

func TestSetApprovalForAllERC1155_ZeroOperatorAddress(t *testing.T) {
	params := txgenerator.SetApprovalForAllERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Operator:     common.Address{}, // zero address
		Approved:     true,
	}

	tx, err := txgenerator.SetApprovalForAllERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "operator address cannot be zero")
}

func TestSetApprovalForAllERC1155_MissingChainID(t *testing.T) {
	params := txgenerator.SetApprovalForAllERC1155Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  nil, // missing chain ID
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Operator:     common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Approved:     true,
	}

	tx, err := txgenerator.SetApprovalForAllERC1155(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "chain ID is required")
}
