package txgenerator

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc1155"
)

// TransferERC1155Params contains parameters for generating an ERC1155 single token transfer transaction
type TransferERC1155Params struct {
	BaseTxParams
	// TokenAddress is the ERC1155 token contract address
	TokenAddress common.Address
	// From is the current owner address
	From common.Address
	// To is the recipient address
	To common.Address
	// TokenID is the token ID to transfer
	TokenID *big.Int
	// Value is the amount of tokens to transfer
	Value *big.Int
}

// BatchTransferERC1155Params contains parameters for generating an ERC1155 batch token transfer transaction
type BatchTransferERC1155Params struct {
	BaseTxParams
	// TokenAddress is the ERC1155 token contract address
	TokenAddress common.Address
	// From is the current owner address
	From common.Address
	// To is the recipient address
	To common.Address
	// TokenIDs is the array of token IDs to transfer
	TokenIDs []*big.Int
	// Values is the array of amounts to transfer (corresponding to TokenIDs)
	Values []*big.Int
}

// SetApprovalForAllERC1155Params contains parameters for generating an ERC1155 setApprovalForAll transaction
type SetApprovalForAllERC1155Params struct {
	BaseTxParams
	// TokenAddress is the ERC1155 token contract address
	TokenAddress common.Address
	// Operator is the operator address to approve or revoke
	Operator common.Address
	// Approved indicates whether to approve (true) or revoke (false) the operator
	Approved bool
}

// TransferERC1155 generates an unsigned transaction for transferring a single ERC1155 token
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func TransferERC1155(params TransferERC1155Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.From == (common.Address{}) {
		return nil, errors.New("from address cannot be zero")
	}
	if params.TokenID == nil || params.TokenID.Sign() < 0 {
		return nil, errors.New("token ID must be non-negative")
	}
	if params.Value == nil || params.Value.Sign() < 0 {
		return nil, errors.New("value must be non-negative")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC1155 safeTransferFrom function call
	// safeTransferFrom(address from, address to, uint256 id, uint256 value, bytes data)
	erc1155ABI, err := erc1155.Erc1155MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	// Pass empty bytes for data parameter
	data, err := erc1155ABI.Pack("safeTransferFrom", params.From, params.To, params.TokenID, params.Value, []byte{})
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}

// BatchTransferERC1155 generates an unsigned transaction for batch transferring ERC1155 tokens
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func BatchTransferERC1155(params BatchTransferERC1155Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.From == (common.Address{}) {
		return nil, errors.New("from address cannot be zero")
	}
	if len(params.TokenIDs) == 0 {
		return nil, errors.New("token IDs cannot be empty")
	}
	if len(params.Values) == 0 {
		return nil, errors.New("values cannot be empty")
	}
	if len(params.TokenIDs) != len(params.Values) {
		return nil, errors.New("token IDs and values arrays must have the same length")
	}
	for i, tokenID := range params.TokenIDs {
		if tokenID == nil || tokenID.Sign() < 0 {
			return nil, errors.New("token ID must be non-negative")
		}
		if params.Values[i] == nil || params.Values[i].Sign() < 0 {
			return nil, errors.New("value must be non-negative")
		}
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC1155 safeBatchTransferFrom function call
	// safeBatchTransferFrom(address from, address to, uint256[] ids, uint256[] values, bytes data)
	erc1155ABI, err := erc1155.Erc1155MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	// Pass empty bytes for data parameter
	data, err := erc1155ABI.Pack("safeBatchTransferFrom", params.From, params.To, params.TokenIDs, params.Values, []byte{})
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}

// SetApprovalForAllERC1155 generates an unsigned transaction for setting approval for all ERC1155 tokens
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func SetApprovalForAllERC1155(params SetApprovalForAllERC1155Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.Operator == (common.Address{}) {
		return nil, errors.New("operator address cannot be zero")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC1155 setApprovalForAll function call
	// setApprovalForAll(address operator, bool approved)
	erc1155ABI, err := erc1155.Erc1155MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	data, err := erc1155ABI.Pack("setApprovalForAll", params.Operator, params.Approved)
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}
