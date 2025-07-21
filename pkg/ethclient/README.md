# ethclient

A comprehensive Ethereum JSON-RPC client that uses go-ethereum's `rpc.Client` for reliable communication while providing custom types for all Ethereum JSON-RPC responses.

## Features

- **Complete Ethereum JSON-RPC Client**: All 47 Ethereum JSON-RPC methods
- **Direct RPC Implementation**: Uses go-ethereum's `rpc.Client` with custom types
- **Modern Ethereum Support**: Legacy, EIP-1559, EIP-4844 transactions
- **Event Filtering**: Comprehensive event filtering and subscription management
- **Contract Interaction**: Built-in support for contract calls and storage access
- **Compatibility**: Same method names as go-ethereum's `ethclient`

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/status-im/go-wallet-sdk/pkg/ethclient"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/rpc"
)

func main() {
    // Create client
    rpcClient, err := rpc.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
    if err != nil {
        log.Fatalf("Failed to dial RPC: %v", err)
    }
    defer rpcClient.Close()

    client := ethclient.NewClient(rpcClient)
    ctx := context.Background()

    // Get latest block number
    blockNumber, err := client.BlockNumber(ctx)
    if err != nil {
        log.Fatalf("Error getting block number: %v", err)
    }
    fmt.Printf("Latest block: %d\n", blockNumber)

    // Get balance
    address := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
    balance, err := client.BalanceAt(ctx, address, nil)
    if err != nil {
        log.Fatalf("Error getting balance: %v", err)
    }
    fmt.Printf("Balance: %s wei\n", balance.String())
}
```

## Client Methods

### Core Methods (go-ethereum ethclient compatible)

#### Blockchain Information
- `BlockNumber()` - Get latest block number
- `ChainID()` - Get chain ID
- `SuggestGasPrice()` - Get current gas price
- `SuggestGasTipCap()` - Get suggested gas tip cap

#### Account Information
- `BalanceAt()` - Get account balance at specific block
- `CodeAt()` - Get contract code at specific block
- `StorageAt()` - Get storage value at specific position
- `PendingNonceAt()` - Get account nonce

#### Contract Interaction
- `CallContract()` - Execute contract call
- `EstimateGas()` - Estimate gas for transaction

#### Transaction Information
- `TransactionByHash()` - Get transaction by hash
- `TransactionCount()` - Get transaction count in block
- `TransactionInBlock()` - Get transaction by block and index
- `TransactionReceipt()` - Get transaction receipt

#### Block Information
- `BlockByHash()` - Get block by hash (returns custom Block type)
- `BlockByNumber()` - Get block by number (returns custom Block type)

#### Transaction Submission
- `SendRawTransaction()` - Send signed transaction

### Ethereum-Specific Methods

#### Event Filtering
- `EthNewFilter()` - Create event filter (uses go-ethereum's FilterQuery)
- `EthNewBlockFilter()` - Create block filter
- `EthGetFilterChanges()` - Get filter changes (returns custom Log types)
- `EthGetFilterLogs()` - Get filter logs (returns custom Log types)
- `EthUninstallFilter()` - Remove filter
- `EthGetLogs()` - Get logs directly (uses go-ethereum's FilterQuery)

#### Transaction Information
- `EthGetTransactionByBlockHashAndIndex()` - Get transaction by block and index
- `EthGetTransactionByBlockNumberAndIndex()` - Get transaction by block and index
- `EthGetTransactionReceipt()` - Get transaction receipt (returns custom Receipt type)

#### Block Information
- `EthGetBlockTransactionCountByNumber()` - Get transaction count in block by number
- `EthGetBlockReceipts()` - Get all transaction receipts in a block

#### Uncle Blocks (PoW)
- `EthGetUncleByBlockHashAndIndex()` - Get uncle by hash and index
- `EthGetUncleByBlockNumberAndIndex()` - Get uncle by number and index
- `EthGetUncleCountByBlockHash()` - Get uncle count by hash
- `EthGetUncleCountByBlockNumber()` - Get uncle count by number

#### Mining (for mining nodes)
- `EthMining()` - Check if mining
- `EthHashrate()` - Get hashrate
- `EthGetWork()` - Get mining work
- `EthSubmitWork()` - Submit proof-of-work
- `EthSubmitHashrate()` - Submit hashrate

#### Network Status
- `EthSyncing()` - Get sync status
- `EthProtocolVersion()` - Get protocol version
- `EthCoinbase()` - Get coinbase address

#### Transaction Operations
- `EthSendTransaction()` - Send transaction
- `EthSign()` - Sign data
- `EthSignTransaction()` - Sign transaction

#### Advanced Features
- `EthFeeHistory()` - Get fee history for EIP-1559
- `EthMaxPriorityFeePerGas()` - Get max priority fee per gas
- `EthGetProof()` - Get account and storage proofs

#### Web3 Methods
- `Web3ClientVersion()` - Get client version
- `Web3Sha3()` - Hash data using Keccak-256

#### Net Methods
- `NetListening()` - Check if node is listening
- `NetPeerCount()` - Get number of connected peers
- `NetVersion()` - Get network ID

### Convenience Methods

- `GetLatestBlock()` - Get latest block (returns custom Block type)
- `GetLatestBlockNumber()` - Get latest block number
- `GetBalance()` - Get balance at latest block
- `GetTransactionByHash()` - Get transaction by hash (returns custom Transaction type)
- `GetTransactionReceipt()` - Get transaction receipt (returns custom Receipt type)
- `GetCode()` - Get contract code at latest block
- `GetNonce()` - Get account nonce
- `GetGasPrice()` - Get current gas price
- `GetChainID()` - Get chain ID
- `GetNetworkID()` - Get network ID
- `GetClientVersion()` - Get client version
- `IsConnected()` - Check connection status

## Types

### Core Blockchain Types

- **Block**: Complete block structure with all fields including post-merge attributes
- **Transaction**: Transaction structure supporting all transaction types
- **Receipt**: Transaction receipt with logs and status
- **Log**: Event log entry with topics and data
- **Withdrawal**: Validator withdrawal information (post-merge)

### RPC Response Types

- **RPCResponse**: Generic JSON-RPC response wrapper
- **RPCError**: Standard JSON-RPC error structure
- **FilterID**: Filter identifier for event subscriptions
- **AccessList**: EIP-2930 access list structure
- **WorkData**: Mining work data array

### Utility Types

- **BlockNumber**: Block number with support for "latest", "earliest", "pending"
- **FeeHistory**: Fee history data for EIP-1559
- **ProofResult**: Account and storage proof results

### Type Design

- **Exported types**: Use `*big.Int`, `uint64`, `[]byte` for clean API
- **Internal types**: Use `hexutil` types for JSON-RPC compatibility
- **Automatic conversion**: Seamless marshaling/unmarshaling between formats

## Supported Features

### Transaction Types
- **Legacy Transactions**: Standard pre-EIP-1559 transactions
- **EIP-1559 Transactions**: Fee market transactions with base fee and priority fee
- **EIP-2930 Transactions**: Access list transactions
- **EIP-4844 Transactions**: Blob transactions (Danksharding)

### Network Features
- **Proof of Work**: Full support for PoW networks (uncles, mining)
- **Proof of Stake**: Full support for PoS networks (withdrawals, beacon root)
- **EIP-1559**: Fee market with base fee and priority fee
- **EIP-4844**: Blob transactions and gas

### Advanced Features
- **Event Filtering**: Real-time event monitoring
- **Storage Proofs**: Merkle proofs for account and storage data
- **Fee History**: Historical fee data for gas estimation
- **Batch Requests**: Support for batch RPC calls
- **WebSocket**: Support for WebSocket connections

## Migration from go-ethereum's ethclient

```go
// Before (go-ethereum ethclient)
ethClient, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
blockNumber, err := ethClient.BlockNumber(ctx)
balance, err := ethClient.BalanceAt(ctx, address, nil)

// After (our client)
rpcClient, err := rpc.Dial("https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
client := ethclient.NewClient(rpcClient)
blockNumber, err := client.BlockNumber(ctx)  // Same method name!
balance, err := client.BalanceAt(ctx, address, nil)  // Same method name!
```

## Examples

See `examples/ethclient-usage/` for comprehensive usage examples including:
- Network information retrieval
- Blockchain data access
- Account balance and nonce queries
- Contract interaction patterns
- Event filtering and monitoring
- Transaction submission
- Gas estimation
- Mining information

Run the example:
```bash
cd examples/ethclient-usage
go run main.go
``` 