# Gas Data Generator

This program converts the functionality of `feesData.sh` into a Go program that fetches gas data from any blockchain network and generates a `data.go` file with the results using the existing SDK types.

## Run

Run the generator with the required command line arguments:

```bash
cd examples/gas-comparison/internal/data

# BSC Mainnet
go run main.go -infura-api-key YOUR_API_KEY -rpc https://bsc-mainnet.infura.io/v3/YOUR_API_KEY

# Ethereum Mainnet
go run main.go -infura-api-key YOUR_API_KEY -rpc https://mainnet.infura.io/v3/YOUR_API_KEY

# Polygon Mainnet
go run main.go -infura-api-key YOUR_API_KEY -rpc https://polygon-mainnet.infura.io/v3/YOUR_API_KEY

# Show help
go run main.go -help
```

3. The program will:
   - Create an RPC client connection to the specified network
   - Automatically detect the chain ID from the RPC endpoint
   - Use SDK's ethclient to fetch the latest block with full transactions
   - Use SDK's ethclient to fetch fee history with percentiles [0, 5, 10, ..., 95, 100]
   - Use SDK's infura client to fetch suggested gas fees for the detected chain ID
   - Generate a chain-specific package and `data.go` file based on the detected chain ID

### Command Line Arguments

- **`-infura-api-key`** (required): Infura API key for gas suggestions
- **`-rpc`** (required): RPC URL for the blockchain network
- **`-help`**: Show help message with examples

### Automatic Chain ID Detection

The generator automatically detects the chain ID by calling `eth_chainId` on the RPC endpoint, eliminating the need to manually specify it.

### Supported Networks

The generator creates human-readable package names for popular networks:

| Chain ID | Package Name | Network Name |
|----------|--------------|--------------|
| 1        | `ethereum`   | Ethereum Mainnet |
| 56       | `bsc`        | BSC Mainnet |
| 137      | `polygon`    | Polygon Mainnet |
| 42161    | `arbitrum`   | Arbitrum One |
| 10       | `optimism`   | Optimism Mainnet |
| 8453     | `base`       | Base Mainnet |
| 43114    | `avalanche`  | Avalanche C-Chain |
| 250      | `fantom`     | Fantom Opera |
| 100      | `gnosis`     | Gnosis Chain |
| 25       | `cronos`     | Cronos Mainnet |
| Other    | `chainXXX`   | Chain XXX |

For unknown networks, the package name follows the format `chainXXX` where XXX is the chain ID.

## What it replaces

This Go program replaces the shell script `feesData.sh` which made these 3 requests:

1. **Latest Block**: `eth_getBlockByNumber` with `"latest"` and `true` parameters
2. **Fee History**: `eth_feeHistory` with block count `0x400` (1024 blocks) and percentiles  
3. **Infura Suggested Fees**: GET request to `https://gas-api.metaswap.codefi.network/networks/{chainID}/suggestedGasFees`

Now supports any blockchain network through command line arguments with automatic chain ID detection.

## Generated Output

The generator creates chain-specific packages to allow multiple networks in a single program:

### **Package Structure:**
```
examples/gas-comparison/internal/data/
├── ethereum/     # Chain ID 1 (Ethereum Mainnet)
│   └── data.go
├── bsc/          # Chain ID 56 (BSC Mainnet)  
│   └── data.go
├── polygon/      # Chain ID 137 (Polygon Mainnet)
│   └── data.go
└── chain123/     # Chain ID 123 (Unknown networks use chainXXX format)
    └── data.go
```

### **Generated File Contents:**
Each `data.go` file contains:
- Chain-specific package name (e.g., `package ethereum`, `package bsc`)
- Import of the shared `data.GasData` type from `github.com/status-im/go-wallet-sdk/examples/gas-comparison/internal/data`
- A `GetGasData()` function that returns `*data.GasData` with parsed data from embedded JSON constants
- Embedded JSON data from all three API calls
- Comments indicating the specific network and chain ID

The `GasData` struct is defined once in `data/types.go` and reused across all generated packages, ensuring consistency and reducing code duplication.

### **Usage in Go Programs:**
```go
import (
    "github.com/status-im/go-wallet-sdk/examples/gas-comparison/internal/data"
    ethereumData "github.com/status-im/go-wallet-sdk/examples/gas-comparison/internal/data/ethereum"
    bscData "github.com/status-im/go-wallet-sdk/examples/gas-comparison/internal/data/bsc"
    polygonData "github.com/status-im/go-wallet-sdk/examples/gas-comparison/internal/data/polygon"
)

// Use data from different networks - all return the same *data.GasData type
var ethData, bscGasData, polygonGasData *data.GasData
var err error

ethData, err = ethereumData.GetGasData()
bscGasData, err = bscData.GetGasData()
polygonGasData, err = polygonData.GetGasData()

// All data uses the same GasData type for consistency
fmt.Printf("Ethereum base fee: %s\n", ethData.LatestBlock.BaseFeePerGas)
fmt.Printf("BSC base fee: %s\n", bscGasData.LatestBlock.BaseFeePerGas)
```

## SDK Integration

The generator now uses the SDK's client methods instead of raw HTTP requests:

### **Client Methods Used:**
- **Ethereum Client**: `ethclient.NewClient()` with RPC client
  - `client.ChainID(ctx)` - automatically detects chain ID
  - `client.GetLatestBlock(ctx)` - fetches latest block with full transactions
  - `client.FeeHistory(ctx, blockCount, lastBlock, percentiles)` - fetches fee history
- **Infura Client**: `infura.NewClient(apiKey)`
  - `client.GetGasSuggestions(ctx, chainID)` - fetches gas suggestions

### **SDK Types Used:**
- **Block Data**: `github.com/status-im/go-wallet-sdk/pkg/ethclient.BlockWithFullTxs`
- **Fee History**: `github.com/status-im/go-wallet-sdk/pkg/ethclient.FeeHistory`
- **Infura Fees**: `github.com/status-im/go-wallet-sdk/pkg/gas/infura.GasResponse`

### **Benefits:**
✅ **Proper Error Handling**: Uses SDK's built-in error handling and retries  
✅ **Type Safety**: Automatic JSON marshaling/unmarshaling with proper types  
✅ **Connection Management**: Proper RPC connection lifecycle management  
✅ **Consistency**: Uses the same client methods as the rest of the SDK  
✅ **Timeout Handling**: Built-in timeout and context management  
✅ **Auto Chain Detection**: Automatically detects chain ID from RPC endpoint  
✅ **Multi-Network Support**: Generate data for multiple networks without conflicts  
✅ **Clean Package Structure**: Each network gets its own package namespace  
✅ **Shared Type Definitions**: Single `GasData` type definition prevents duplication and ensures consistency  
✅ **Type Safety**: All generated packages return the same `*data.GasData` type for easy interoperability  

This ensures full integration with the SDK and allows you to use the gas data in your Go programs without making additional API calls.

## Important Note

If you have existing generated files from previous versions of the generator, you should regenerate them to use the new shared `data.GasData` type instead of having duplicate type definitions in each file. Simply run the generator again for each network you want to update:

```bash
cd examples/gas-comparison/internal/data/generator

# Regenerate existing data files to use shared types
go run main.go -infura-api-key YOUR_API_KEY -rpc https://mainnet.infura.io/v3/YOUR_API_KEY
go run main.go -infura-api-key YOUR_API_KEY -rpc https://bsc-mainnet.infura.io/v3/YOUR_API_KEY
```