# Transaction Generator

The `txgenerator` package provides utilities for generating unsigned Ethereum transactions. It supports creating transactions for:

- **ETH transfers**: Simple native token transfers
- **ERC20 tokens**: Transfers and approvals
- **ERC721 tokens (NFTs)**: Transfers and approvals
- **ERC1155 tokens**: Single and batch transfers and approvals

## Use it when

- You want to build an unsigned `types.Transaction` for signing elsewhere.
- You need helper constructors for common token operations (ETH/ERC20/ERC721/ERC1155).
- You want automatic legacy vs EIP-1559 transaction type selection.

## Key entrypoints

### Native ETH
- `txgenerator.TransferETH(params)`

### ERC20 Tokens
- `txgenerator.TransferERC20(params)`
- `txgenerator.ApproveERC20(params)`

### ERC721 Tokens (NFTs)
- `txgenerator.TransferFromERC721(params)`
- `txgenerator.SafeTransferFromERC721(params)`
- `txgenerator.ApproveERC721(params)`
- `txgenerator.SetApprovalForAllERC721(params)`

### ERC1155 Tokens
- `txgenerator.TransferERC1155(params)`
- `txgenerator.BatchTransferERC1155(params)`
- `txgenerator.SetApprovalForAllERC1155(params)`

## Features

- Supports both legacy (type 0) and EIP-1559 (type 2) transactions
- Automatic transaction type detection based on provided gas parameters
- Returns unsigned `types.Transaction` objects ready for signing
- Comprehensive parameter validation

## Usage

### ETH Transfer

```go
import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

// Legacy transaction example
tx, err := txgenerator.TransferETH(txgenerator.TransferETHParams{
    To:       common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    Value:    big.NewInt(1000000000000000000), // 1 ETH in wei
    Nonce:    0,
    GasLimit: 21000,
    GasPrice: big.NewInt(20000000000), // 20 gwei
    ChainID:  big.NewInt(1), // Ethereum mainnet
})

// EIP-1559 transaction example
tx, err := txgenerator.TransferETH(txgenerator.TransferETHParams{
    To:                  common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    Value:               big.NewInt(1000000000000000000), // 1 ETH in wei
    Nonce:               0,
    GasLimit:            21000,
    MaxFeePerGas:        big.NewInt(30000000000), // 30 gwei
    MaxPriorityFeePerGas: big.NewInt(2000000000), // 2 gwei
    ChainID:             big.NewInt(1), // Ethereum mainnet
})
```

### ERC20 Token Transfer

```go
import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

// Legacy transaction example
tx, err := txgenerator.TransferERC20(txgenerator.TransferERC20Params{
    TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
    To:           common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    Amount:       big.NewInt(1000000), // 1 USDC (6 decimals)
    Nonce:        0,
    GasLimit:     65000,
    GasPrice:     big.NewInt(20000000000), // 20 gwei
    ChainID:      big.NewInt(1), // Ethereum mainnet
})

// EIP-1559 transaction example
tx, err := txgenerator.TransferERC20(txgenerator.TransferERC20Params{
    TokenAddress:        common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
    To:                  common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    Amount:              big.NewInt(1000000), // 1 USDC (6 decimals)
    Nonce:               0,
    GasLimit:            65000,
    MaxFeePerGas:        big.NewInt(30000000000), // 30 gwei
    MaxPriorityFeePerGas: big.NewInt(2000000000), // 2 gwei
    ChainID:             big.NewInt(1), // Ethereum mainnet
})
```

### ERC20 Token Approval

```go
import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

// Approve ERC20 token spending
tx, err := txgenerator.ApproveERC20(txgenerator.ApproveERC20Params{
    TokenAddress: common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
    Spender:      common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    Amount:       big.NewInt(1000000), // 1 USDC (6 decimals)
    Nonce:        0,
    GasLimit:     46000,
    GasPrice:     big.NewInt(20000000000), // 20 gwei
    ChainID:      big.NewInt(1), // Ethereum mainnet
})
```

### ERC721 Token Transfer (NFT)

```go
import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

// TransferFrom (basic transfer)
tx, err := txgenerator.TransferFromERC721(txgenerator.TransferERC721Params{
    TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"), // Bored Ape Yacht Club
    From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"), // Current owner
    To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"), // Recipient
    TokenID:      big.NewInt(1234), // Token ID
    Nonce:        0,
    GasLimit:     100000,
    GasPrice:     big.NewInt(20000000000), // 20 gwei
    ChainID:      big.NewInt(1), // Ethereum mainnet
})

// SafeTransferFrom (recommended - checks if recipient can handle ERC721)
tx, err := txgenerator.SafeTransferFromERC721(txgenerator.TransferERC721Params{
    TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
    From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
    TokenID:      big.NewInt(1234),
    Nonce:        0,
    GasLimit:     100000,
    MaxFeePerGas:        big.NewInt(30000000000), // 30 gwei
    MaxPriorityFeePerGas: big.NewInt(2000000000),  // 2 gwei
    ChainID:             big.NewInt(1), // Ethereum mainnet
})
```

### ERC721 Token Approval

```go
// Approve a specific NFT
tx, err := txgenerator.ApproveERC721(txgenerator.ApproveERC721Params{
    TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
    To:           common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"), // Approved address
    TokenID:      big.NewInt(1234),
    Nonce:        0,
    GasLimit:     46000,
    GasPrice:     big.NewInt(20000000000),
    ChainID:      big.NewInt(1),
})

// Set approval for all NFTs (operator)
tx, err := txgenerator.SetApprovalForAllERC721(txgenerator.SetApprovalForAllERC721Params{
    TokenAddress: common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"),
    Operator:     common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    Approved:     true, // true to approve, false to revoke
    Nonce:        0,
    GasLimit:     46000,
    GasPrice:     big.NewInt(20000000000),
    ChainID:      big.NewInt(1),
})
```

### ERC1155 Token Transfer

```go
import (
    "math/big"
    "github.com/ethereum/go-ethereum/common"
    "github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

// Single token transfer
tx, err := txgenerator.TransferERC1155(txgenerator.TransferERC1155Params{
    TokenAddress: common.HexToAddress("0x..."), // ERC1155 contract
    From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
    TokenID:      big.NewInt(1),
    Value:        big.NewInt(100), // Amount of tokens
    Nonce:        0,
    GasLimit:     100000,
    GasPrice:     big.NewInt(20000000000),
    ChainID:      big.NewInt(1),
})

// Batch transfer multiple tokens
tx, err := txgenerator.BatchTransferERC1155(txgenerator.BatchTransferERC1155Params{
    TokenAddress: common.HexToAddress("0x..."),
    From:         common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    To:           common.HexToAddress("0x8ba1f109551bd432803012645ac136ddd64dba72"),
    TokenIDs:     []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
    Values:       []*big.Int{big.NewInt(10), big.NewInt(20), big.NewInt(30)},
    Nonce:        0,
    GasLimit:     200000,
    MaxFeePerGas:        big.NewInt(30000000000),
    MaxPriorityFeePerGas: big.NewInt(2000000000),
    ChainID:             big.NewInt(1),
})
```

### ERC1155 Operator Approval

```go
// Set approval for all ERC1155 tokens
tx, err := txgenerator.SetApprovalForAllERC1155(txgenerator.SetApprovalForAllERC1155Params{
    TokenAddress: common.HexToAddress("0x..."),
    Operator:     common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"),
    Approved:     true, // true to approve, false to revoke
    Nonce:        0,
    GasLimit:     46000,
    GasPrice:     big.NewInt(20000000000),
    ChainID:      big.NewInt(1),
})
```

## Transaction Type Detection

The transaction type is automatically determined based on the provided parameters:

- **Legacy (type 0)**: If `GasPrice` is provided
- **EIP-1559 (type 2)**: If `MaxFeePerGas` or `MaxPriorityFeePerGas` is provided

For EIP-1559 transactions, both `MaxFeePerGas` and `MaxPriorityFeePerGas` must be provided.

## Signing Transactions

The generated transactions are unsigned and must be signed before being sent to the network. Use the keystore or extkeystore modules for signing:

```go
// After generating the transaction
signedTx, err := keystore.SignTx(account, tx, chainID)
if err != nil {
    // handle error
}

// Send the signed transaction
err = ethClient.SendTransaction(ctx, signedTx)
```

## Error Handling

The package returns specific errors for common issues:

- `ErrInvalidParams`: Invalid transaction parameters
- `ErrMissingGasPrice`: GasPrice required for legacy transactions
- `ErrMissingMaxFeePerGas`: MaxFeePerGas required for EIP-1559 transactions
- `ErrMissingMaxPriorityFeePerGas`: MaxPriorityFeePerGas required for EIP-1559 transactions

## Notes

- All transactions are created without signatures
- ChainID is required for all transactions
- For ERC20 transfers, the amount should be in the token's base units (consider decimals)
- For ERC721 transfers:
  - `SafeTransferFromERC721` uses `safeTransferFrom` which checks if the recipient can handle ERC721 tokens (recommended)
  - `TransferFromERC721` uses `transferFrom` which is a basic transfer without safety checks
  - The `From` address must be the current owner of the token
- For ERC1155 transfers:
  - `TransferERC1155` transfers a single token type
  - `BatchTransferERC1155` transfers multiple token types in one transaction
  - TokenIDs and Values arrays must have the same length for batch transfers
- Gas limits should be estimated separately (not included in this package)

## See Also

- [Ethereum Client](../ethclient/README.md) - RPC client for transaction operations
- [Accounts Package](../accounts/README.md) - Account management and transaction signing
- [Gas Package](../gas/README.md) - Gas price estimation and fee suggestions

## Examples

- [Transaction Generator Web Example](../../examples/txgenerator-example/README.md) - Web interface for generating transactions
