package txgenerator

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
)

// TransferERC20Params contains parameters for generating an ERC20 token transfer transaction
type TransferERC20Params struct {
	BaseTxParams
	// TokenAddress is the ERC20 token contract address
	TokenAddress common.Address
	// To is the recipient address
	To common.Address
	// Amount is the amount of tokens to transfer (in token units, not wei)
	Amount *big.Int
}

// ApproveERC20Params contains parameters for generating an ERC20 approval transaction
type ApproveERC20Params struct {
	BaseTxParams
	// TokenAddress is the ERC20 token contract address
	TokenAddress common.Address
	// Spender is the address to approve
	Spender common.Address
	// Amount is the amount of tokens to approve (in token units, not wei)
	Amount *big.Int
}

// TransferERC20 generates an unsigned transaction for transferring ERC20 tokens
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func TransferERC20(params TransferERC20Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.Amount == nil || params.Amount.Sign() < 0 {
		return nil, errors.New("amount must be non-negative")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC20 transfer function call
	// transfer(address recipient, uint256 amount)
	erc20ABI, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	data, err := erc20ABI.Pack("transfer", params.To, params.Amount)
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}

// ApproveERC20 generates an unsigned transaction for approving ERC20 token spending
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func ApproveERC20(params ApproveERC20Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.Spender == (common.Address{}) {
		return nil, errors.New("spender address cannot be zero")
	}
	if params.Amount == nil || params.Amount.Sign() < 0 {
		return nil, errors.New("amount must be non-negative")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC20 approve function call
	// approve(address spender, uint256 amount)
	erc20ABI, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	data, err := erc20ABI.Pack("approve", params.Spender, params.Amount)
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}
