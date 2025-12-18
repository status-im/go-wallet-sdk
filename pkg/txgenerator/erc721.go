package txgenerator

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc721"
)

// TransferERC721Params contains parameters for generating an ERC721 token transfer transaction
type TransferERC721Params struct {
	BaseTxParams
	// TokenAddress is the ERC721 token contract address
	TokenAddress common.Address
	// From is the current owner address
	From common.Address
	// To is the recipient address
	To common.Address
	// TokenID is the token ID to transfer
	TokenID *big.Int
}

// ApproveERC721Params contains parameters for generating an ERC721 approval transaction
type ApproveERC721Params struct {
	BaseTxParams
	// TokenAddress is the ERC721 token contract address
	TokenAddress common.Address
	// To is the address to approve
	To common.Address
	// TokenID is the token ID to approve
	TokenID *big.Int
}

// SetApprovalForAllERC721Params contains parameters for generating an ERC721 setApprovalForAll transaction
type SetApprovalForAllERC721Params struct {
	BaseTxParams
	// TokenAddress is the ERC721 token contract address
	TokenAddress common.Address
	// Operator is the operator address to approve or revoke
	Operator common.Address
	// Approved indicates whether to approve (true) or revoke (false) the operator
	Approved bool
}

// TransferFromERC721 generates an unsigned transaction for transferring ERC721 tokens
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func TransferFromERC721(params TransferERC721Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.From == (common.Address{}) {
		return nil, errors.New("from address cannot be zero")
	}
	if params.TokenID == nil || params.TokenID.Sign() < 0 {
		return nil, errors.New("token ID must be non-negative")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC721 transfer function call
	erc721ABI, err := erc721.Erc721MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	// Use transferFrom(address from, address to, uint256 tokenId)
	data, err := erc721ABI.Pack("transferFrom", params.From, params.To, params.TokenID)
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}

// SafeTransferFromERC721 generates an unsigned transaction for transferring ERC721 tokens using safeTransferFrom
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func SafeTransferFromERC721(params TransferERC721Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.From == (common.Address{}) {
		return nil, errors.New("from address cannot be zero")
	}
	if params.TokenID == nil || params.TokenID.Sign() < 0 {
		return nil, errors.New("token ID must be non-negative")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC721 transfer function call
	erc721ABI, err := erc721.Erc721MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	// Use safeTransferFrom(address from, address to, uint256 tokenId)
	data, err := erc721ABI.Pack("safeTransferFrom", params.From, params.To, params.TokenID)
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}

// ApproveERC721 generates an unsigned transaction for approving a specific ERC721 token
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func ApproveERC721(params ApproveERC721Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.To == (common.Address{}) {
		return nil, errors.New("approval address cannot be zero")
	}
	if params.TokenID == nil || params.TokenID.Sign() < 0 {
		return nil, errors.New("token ID must be non-negative")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC721 approve function call
	// approve(address to, uint256 tokenId)
	erc721ABI, err := erc721.Erc721MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	data, err := erc721ABI.Pack("approve", params.To, params.TokenID)
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}

// SetApprovalForAllERC721 generates an unsigned transaction for setting approval for all ERC721 tokens
// The transaction type is determined by the presence of GasPrice (legacy) or MaxFeePerGas (EIP-1559)
func SetApprovalForAllERC721(params SetApprovalForAllERC721Params) (*types.Transaction, error) {
	if params.TokenAddress == (common.Address{}) {
		return nil, errors.New("token address cannot be zero")
	}
	if params.Operator == (common.Address{}) {
		return nil, errors.New("operator address cannot be zero")
	}
	if params.ChainID == nil {
		return nil, errors.New("chain ID is required")
	}

	// Encode the ERC721 setApprovalForAll function call
	// setApprovalForAll(address operator, bool approved)
	erc721ABI, err := erc721.Erc721MetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	data, err := erc721ABI.Pack("setApprovalForAll", params.Operator, params.Approved)
	if err != nil {
		return nil, err
	}

	return createTokenTransaction(params.TokenAddress, params.Nonce, params.GasLimit, params.ChainID, params.GasPrice, params.MaxFeePerGas, params.MaxPriorityFeePerGas, data)
}
