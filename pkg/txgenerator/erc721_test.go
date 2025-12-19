package txgenerator_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

func TestTransferFromERC721_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D") // BAYC
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72")
	tokenID := big.NewInt(1234)
	nonce := uint64(0)
	gasLimit := uint64(100000)
	gasPrice := big.NewInt(20000000000) // 20 gwei
	chainID := big.NewInt(1)

	params := txgenerator.TransferERC721Params{
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
	}

	tx, err := txgenerator.TransferFromERC721(params)
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

func TestTransferFromERC721_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D")
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72")
	tokenID := big.NewInt(1234)
	nonce := uint64(0)
	gasLimit := uint64(100000)
	maxFeePerGas := big.NewInt(30000000000)        // 30 gwei
	maxPriorityFeePerGas := big.NewInt(2000000000) // 2 gwei
	chainID := big.NewInt(1)

	params := txgenerator.TransferERC721Params{
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
	}

	tx, err := txgenerator.TransferFromERC721(params)
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

func TestTransferFromERC721_ZeroTokenAddress(t *testing.T) {
	params := txgenerator.TransferERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.Address{}, // zero address
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(1234),
	}

	tx, err := txgenerator.TransferFromERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token address cannot be zero")
}

func TestTransferFromERC721_ZeroFromAddress(t *testing.T) {
	params := txgenerator.TransferERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
		From:         common.Address{}, // zero address
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(1234),
	}

	tx, err := txgenerator.TransferFromERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "from address cannot be zero")
}

func TestTransferFromERC721_ZeroToAddress_Burn(t *testing.T) {
	// Zero address is valid for burning tokens
	params := txgenerator.TransferERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.Address{}, // zero address for burning
		TokenID:      big.NewInt(1234),
	}

	tx, err := txgenerator.TransferFromERC721(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestTransferFromERC721_NegativeTokenID(t *testing.T) {
	params := txgenerator.TransferERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      big.NewInt(-1), // negative token ID
	}

	tx, err := txgenerator.TransferFromERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token ID must be non-negative")
}

func TestTransferFromERC721_NilTokenID(t *testing.T) {
	params := txgenerator.TransferERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
		TokenID:      nil,
	}

	tx, err := txgenerator.TransferFromERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token ID must be non-negative")
}

func TestSafeTransferFromERC721_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D")
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72")
	tokenID := big.NewInt(1234)
	nonce := uint64(0)
	gasLimit := uint64(100000)
	gasPrice := big.NewInt(20000000000)
	chainID := big.NewInt(1)

	params := txgenerator.TransferERC721Params{
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
	}

	tx, err := txgenerator.SafeTransferFromERC721(params)
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

func TestSafeTransferFromERC721_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D")
	from := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	to := common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72")
	tokenID := big.NewInt(1234)
	nonce := uint64(0)
	gasLimit := uint64(100000)
	maxFeePerGas := big.NewInt(30000000000)
	maxPriorityFeePerGas := big.NewInt(2000000000)
	chainID := big.NewInt(1)

	params := txgenerator.TransferERC721Params{
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
	}

	tx, err := txgenerator.SafeTransferFromERC721(params)
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

func TestSafeTransferFromERC721_ZeroToAddress_Burn(t *testing.T) {
	// Zero address is valid for burning tokens
	params := txgenerator.TransferERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 100000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
		From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		To:           common.Address{}, // zero address for burning
		TokenID:      big.NewInt(1234),
	}

	tx, err := txgenerator.SafeTransferFromERC721(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestApproveERC721_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D")
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	tokenID := big.NewInt(1234)
	nonce := uint64(0)
	gasLimit := uint64(46000)
	gasPrice := big.NewInt(20000000000)
	chainID := big.NewInt(1)

	params := txgenerator.ApproveERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    nonce,
			GasLimit: gasLimit,
			ChainID:  chainID,
			GasPrice: gasPrice,
		},
		TokenAddress: tokenAddress,
		To:           to,
		TokenID:      tokenID,
	}

	tx, err := txgenerator.ApproveERC721(params)
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

func TestApproveERC721_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D")
	to := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	tokenID := big.NewInt(1234)
	nonce := uint64(0)
	gasLimit := uint64(46000)
	maxFeePerGas := big.NewInt(30000000000)
	maxPriorityFeePerGas := big.NewInt(2000000000)
	chainID := big.NewInt(1)

	params := txgenerator.ApproveERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:                nonce,
			GasLimit:             gasLimit,
			ChainID:              chainID,
			MaxFeePerGas:         maxFeePerGas,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
		},
		TokenAddress: tokenAddress,
		To:           to,
		TokenID:      tokenID,
	}

	tx, err := txgenerator.ApproveERC721(params)
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

func TestApproveERC721_ZeroTokenAddress(t *testing.T) {
	params := txgenerator.ApproveERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.Address{}, // zero address
		To:           common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		TokenID:      big.NewInt(1234),
	}

	tx, err := txgenerator.ApproveERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token address cannot be zero")
}

func TestApproveERC721_ZeroToAddress(t *testing.T) {
	params := txgenerator.ApproveERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
		To:           common.Address{}, // zero address
		TokenID:      big.NewInt(1234),
	}

	tx, err := txgenerator.ApproveERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "approval address cannot be zero")
}

func TestSetApprovalForAllERC721_LegacyTransaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D")
	operator := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	approved := true
	nonce := uint64(0)
	gasLimit := uint64(46000)
	gasPrice := big.NewInt(20000000000)
	chainID := big.NewInt(1)

	params := txgenerator.SetApprovalForAllERC721Params{
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

	tx, err := txgenerator.SetApprovalForAllERC721(params)
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

func TestSetApprovalForAllERC721_EIP1559Transaction(t *testing.T) {
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D")
	operator := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	approved := true
	nonce := uint64(0)
	gasLimit := uint64(46000)
	maxFeePerGas := big.NewInt(30000000000)
	maxPriorityFeePerGas := big.NewInt(2000000000)
	chainID := big.NewInt(1)

	params := txgenerator.SetApprovalForAllERC721Params{
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

	tx, err := txgenerator.SetApprovalForAllERC721(params)
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

func TestSetApprovalForAllERC721_RevokeApproval(t *testing.T) {
	// Test revoking approval (approved = false)
	tokenAddress := common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D")
	operator := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	approved := false
	nonce := uint64(0)
	gasLimit := uint64(46000)
	gasPrice := big.NewInt(20000000000)
	chainID := big.NewInt(1)

	params := txgenerator.SetApprovalForAllERC721Params{
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

	tx, err := txgenerator.SetApprovalForAllERC721(params)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestSetApprovalForAllERC721_ZeroTokenAddress(t *testing.T) {
	params := txgenerator.SetApprovalForAllERC721Params{
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

	tx, err := txgenerator.SetApprovalForAllERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "token address cannot be zero")
}

func TestSetApprovalForAllERC721_ZeroOperatorAddress(t *testing.T) {
	params := txgenerator.SetApprovalForAllERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  big.NewInt(1),
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
		Operator:     common.Address{}, // zero address
		Approved:     true,
	}

	tx, err := txgenerator.SetApprovalForAllERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "operator address cannot be zero")
}

func TestSetApprovalForAllERC721_MissingChainID(t *testing.T) {
	params := txgenerator.SetApprovalForAllERC721Params{
		BaseTxParams: txgenerator.BaseTxParams{
			Nonce:    0,
			GasLimit: 46000,
			ChainID:  nil, // missing chain ID
			GasPrice: big.NewInt(20000000000),
		},
		TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
		Operator:     common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
		Approved:     true,
	}

	tx, err := txgenerator.SetApprovalForAllERC721(params)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "chain ID is required")
}
