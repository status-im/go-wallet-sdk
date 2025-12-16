# Ethereum JSON-RPC Client Example

This example demonstrates how to use the `ethclient` package with our custom `eth.go` methods to interact with multiple Ethereum-compatible networks and nodes. Unlike go-ethereum's ethclient, our implementation is chain-agnostic and works with Arbitrum, Optimism, and other EVM chains.

## What it demonstrates

- **Multi-Network Testing**: Test connectivity and functionality across multiple RPC endpoints
- **Network Information**: Get client version, network ID, and chain ID
- **Blockchain Data**: Retrieve latest block information, transactions, and gas prices
- **Account Information**: Get account balances and nonces
- **Contract Interaction**: Check contract code and interact with smart contracts
- **Event Filtering**: Query event logs for specific contracts
- **Transaction Information**: Retrieve detailed transaction details
- **Network Status**: Check node connectivity and network version
- **Gas Estimation**: Estimate gas for contract calls

## Why Our eth.go Methods?

Our custom `eth.go` methods provide several advantages over go-ethereum's ethclient:

- **Chain-agnostic**: Works with any EVM chain (Arbitrum, Optimism, Polygon, etc.)
- **No assumptions**: Makes no assumptions about transaction types or chain-specific values
- **Universal compatibility**: Follows only the standard Ethereum JSON-RPC specification
- **Better L2 support**: Handles edge cases gracefully on non-Ethereum chains

## Run

### Prerequisites

- Go 1.23.0 or later
- Access to Ethereum RPC endpoints (public nodes, Infura, Alchemy, local node, etc.)

### Running the Example

1. **Set your RPC endpoints** (optional):
   ```bash
   export ETH_RPC_ENDPOINTS="https://mainnet.infura.io/v3/YOUR-PROJECT-ID https://optimism-rpc.publicnode.com"
   ```

2. **Run the example**:
   ```bash
   go run main.go
   ```

3. **Or build and run**:
   ```bash
   go build -o ethclient-example
   ./ethclient-example
   ```

### Example Output

```
Testing RPC endpoint: https://ethereum-rpc.publicnode.com
🚀 Ethereum JSON-RPC Client Example (using eth.go methods)
==========================================================

📡 Network Information
Client Version: Geth/v1.16.0-stable
Network ID: 1
Chain ID: 1

⛓️  Blockchain Information
Latest Block Number: 19543210
Latest Block Hash: 0x1234...
Block Number: 19543210
Found 150 Transactions
Transaction 1:
 Hash: 0x5678...
 From: 0xabcd...
 Gas: 21000
Block Timestamp: 1703123456
Gas Used: 0x1c9c380
Gas Limit: 0x1c9c380
Base Fee Per Gas: 15000000000 wei
Current Gas Price: 15000000000 wei

👤 Account Information
Balance of 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045: 1234567890000000000 wei
Balance in ETH: 1.234567890
Nonce of 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045: 42

📄 Contract Interaction
Multicall3 Contract Code Length: 1234 bytes
✅ Contract exists (has code)

🔍 Event Filtering
Found 15 Transfer events in the last 10 blocks
  Event 1: Block 19542210, Tx 0x1234...
  Event 2: Block 19542215, Tx 0x5678...
  Event 3: Block 19542220, Tx 0x9abc...

💸 Transaction Information
Transaction Hash: 0x1234...
From: 0x0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b
To: 0x1111111111111111111111111111111111111111
Value: 1000000000000000000 wei
Gas: 21000
Gas Price: 15000000000 wei

🌐 Network Status
Net Version: 1

⛽ Gas Estimation
Estimated gas for call: 21000
--------------------------------
--------------------------------
--------------------------------

Testing RPC endpoint: https://optimism-rpc.publicnode.com
[Similar output for Optimism network]

✅ Example completed successfully!
```

## Configuration

### Environment Variables

- `ETH_RPC_ENDPOINTS`: Space-separated list of Ethereum RPC endpoint URLs
  - Default: Multiple public nodes (Ethereum, Optimism, Arbitrum, Sepolia)
  - Examples:
    - Single endpoint: `https://mainnet.infura.io/v3/YOUR-PROJECT-ID`
    - Multiple endpoints: `https://mainnet.infura.io/v3/YOUR-PROJECT-ID https://optimism-rpc.publicnode.com`
    - Local nodes: `http://localhost:8545 http://localhost:8546`

### Supported Networks

The example tests multiple networks by default:

- **Ethereum Mainnet**: `https://ethereum-rpc.publicnode.com`
- **Optimism**: `https://optimism-rpc.publicnode.com`
- **Arbitrum**: `https://arbitrum-rpc.publicnode.com`
- **Sepolia Testnet**: `https://public.sepolia.rpc.status.network`

You can also use:
- **Infura**: `https://mainnet.infura.io/v3/YOUR-PROJECT-ID`
- **Alchemy**: `https://eth-mainnet.g.alchemy.com/v2/YOUR-API-KEY`
- **Local nodes**: `http://localhost:8545`

## Code Structure

The example is organized into several functions:

- `main()`: Main function that tests multiple RPC endpoints
- `testRPC()`: Tests a single RPC endpoint with comprehensive functionality
- Multiple example sections demonstrating different client capabilities

## Method Usage

The example demonstrates our custom `eth.go` methods:

```go
// Network information
client.Web3ClientVersion(ctx)      // Get client version
client.NetVersion(ctx)             // Get network ID
client.EthChainId(ctx)             // Get chain ID

// Blockchain data
client.EthBlockNumber(ctx)         // Get latest block number
client.EthGetBlockByNumberWithFullTxs(ctx, blockNum)  // Get block with transactions
client.EthGasPrice(ctx)            // Get current gas price

// Account information
client.EthGetBalance(ctx, address, nil)      // Get account balance
client.EthGetTransactionCount(ctx, address, nil)  // Get account nonce

// Contract interaction
client.EthGetCode(ctx, address, nil)         // Get contract code

// Event filtering
client.EthGetLogs(ctx, filterQuery)          // Get event logs

// Transaction information
client.EthGetTransactionByHash(ctx, hash)    // Get transaction details

// Gas estimation
client.EthEstimateGas(ctx, callMsg)          // Estimate gas for call
```

## Error Handling

The example includes comprehensive error handling for:
- RPC connection failures
- Network timeouts
- Invalid responses
- Missing data
- Network-specific errors

## Dependencies

- `github.com/ethereum/go-ethereum v1.16.0`: Core Ethereum types and RPC client
- `github.com/status-im/go-wallet-sdk`: Our custom client implementation

## Network Compatibility

This example works with any Ethereum-compatible network including:
- Ethereum Mainnet and testnets
- Layer 2 networks (Optimism, Arbitrum, Polygon)
- Other EVM-compatible chains
- Local development networks

The key advantage is that our `eth.go` methods make no assumptions about transaction types or chain-specific implementations, making them universally compatible with any EVM chain that follows the JSON-RPC specification. 