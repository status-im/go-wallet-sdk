# Go Wallet SDK

A modular Go SDK for building multi-chain cryptocurrency wallets and blockchain applications.

## Quick Start

```bash
go get github.com/status-im/go-wallet-sdk
```

## Available Packages

### Balance Management
- **`pkg/balance/fetcher`**: High-performance balance fetching with automatic fallback strategies
  - Native token (ETH) balance fetching for multiple addresses
  - ERC20 token balance fetching for multiple addresses and tokens
  - Smart fallback between different fetching methods
  - Chain-agnostic design

- **`pkg/multicall`**: Efficient batching of contract calls using Multicall3
  - Batch multiple contract calls into single transactions
  - Support for ETH, ERC20, ERC721, and ERC1155 balance queries
  - Automatic chunking and error handling
  - Synchronous and asynchronous execution modes

### Ethereum Client
- **`pkg/ethclient`**: Full-featured Ethereum client with go-ethereum compatibility
  - Complete RPC method coverage (eth_*, net_*, web3_*)
  - Go-ethereum ethclient compatible interface for easy migration
  - Chain-agnostic methods for any EVM-compatible network

### Gas Estimation
- **`pkg/gas`**: Comprehensive gas fee estimation and suggestions
  - Smart fee estimation for three priority levels (low, medium, high)
  - Transaction inclusion time predictions
  - Multi-chain support (L1, Arbitrum, Optimism, Linea)
  - Network congestion analysis for L1 chains
  - Chain-specific optimizations

### Event Filtering & Parsing
- **`pkg/eventfilter`**: Efficient event filtering for ERC20, ERC721, and ERC1155 transfers
  - Minimizes eth_getLogs API calls
  - Direction-based filtering (send, receive, both)
  - Concurrent query processing

- **`pkg/eventlog`**: Automatic event log detection and parsing
  - Type-safe access to parsed event data
  - Support for Transfer, Approval, and other standard events
  - Works seamlessly with eventfilter

### Token Management
- **`pkg/tokens/types`**: Core data structures for tokens and token lists
  - Unified token representation with cross-chain support
  - Token list metadata and versioning
  - Type-safe address handling and validation

- **`pkg/tokens/parsers`**: Token list parsing from multiple formats
  - Standard token list format (Uniswap-style)
  - Status-specific format with chain grouping
  - CoinGecko API format with platform mappings
  - List-of-token-lists metadata parsing

- **`pkg/tokens/fetcher`**: HTTP-based token list fetching
  - Concurrent fetching with goroutines
  - HTTP ETag caching for bandwidth efficiency
  - JSON schema validation support
  - Robust error handling and timeout management

- **`pkg/tokens/autofetcher`**: Automated background token list management
  - Configurable refresh intervals
  - Thread-safe operations with context support
  - Pluggable storage backends
  - Error reporting via channels

- **`pkg/tokens/builder`**: Incremental token collection building
  - Builder pattern for stateful construction
  - Automatic deduplication by chain ID and address
  - Native token generation for supported chains
  - Multiple format support through parsers

- **`pkg/tokens/manager`**: High-level token management interface
  - Multi-source token integration (native, remote, local, custom)
  - Thread-safe concurrent access
  - Rich query capabilities by chain, address, or list ID
  - Automatic refresh and state management

### Smart Contract Bindings
- **`pkg/contracts`**: Go bindings for smart contracts
  - Multicall3 with 200+ chain deployments
  - ERC20, ERC721, and ERC1155 token standards
  - Automated deployment address management

### Common Utilities
- **`pkg/common`**: Shared utilities and constants used across the SDK

## Examples

### Web-Based Balance Fetcher

```bash
cd examples/balance-fetcher-web
go run .
```

Access: http://localhost:8080

### Ethereum Client Usage

```bash
cd examples/ethclient-usage
go run .
```

### Gas Comparison Tool

```bash
cd examples/gas-comparison

# Test with local mock data
go run . -fake

# Test with real networks (requires Infura API key)
go run . -infura-api-key YOUR_API_KEY
```

### Multicall Usage

```bash
cd examples/multiclient3-usage
go run .
```

### Event Filter Example

```bash
cd examples/eventfilter-example
go run . -account 0xYourAddress -start 19000000 -end 19100000
```

### Token Management Examples

```bash
# Token Builder - Incremental token collection building
cd examples/token-builder
go run .

# Token Fetcher - HTTP-based token list fetching
cd examples/token-fetcher
go run .

# Token Manager - High-level token management
cd examples/token-manager
go run .

# Token Parser - Parse different token list formats
cd examples/token-parser
go run .
```

## Testing

```bash
go test ./...
```

## Project Structure

```
go-wallet-sdk/
├── pkg/                    # Core SDK packages
│   ├── balance/           # Balance fetching functionality
│   ├── multicall/         # Multicall3 batching
│   ├── ethclient/         # Ethereum client with full RPC support
│   ├── gas/               # Gas estimation and fee suggestions
│   ├── eventfilter/       # Event filtering for transfers
│   ├── eventlog/          # Event log parsing
│   ├── tokens/            # Token management system
│   │   ├── types/         # Core token data structures
│   │   ├── parsers/       # Token list format parsers
│   │   ├── fetcher/       # HTTP token list fetching
│   │   ├── autofetcher/   # Automated background fetching
│   │   ├── builder/       # Incremental token collection building
│   │   └── manager/       # High-level token management
│   ├── contracts/         # Smart contract bindings
│   └── common/            # Shared utilities
├── examples/              # Usage examples
│   ├── balance-fetcher-web/       # Web interface for balance fetching
│   ├── ethclient-usage/           # Ethereum client examples
│   ├── gas-comparison/            # Gas fee comparison tool
│   ├── multiclient3-usage/        # Multicall examples
│   ├── multistandardfetcher-example/  # Multi-standard balance fetching
│   ├── eventfilter-example/       # Event filtering examples
│   ├── token-builder/             # Token collection building
│   ├── token-fetcher/            # Token list fetching
│   ├── token-manager/             # Token management
│   └── token-parser/              # Token list parsing
└── README.md              # This file
```

## Documentation

### Package Documentation
- [Balance Fetcher](pkg/balance/fetcher/README.md) - Balance fetching functionality
- [Multicall](pkg/multicall/README.md) - Efficient contract call batching
- [Ethereum Client](pkg/ethclient/README.md) - Complete Ethereum RPC client
- [Gas Estimation](pkg/gas/README.md) - Gas fee estimation and suggestions
- [Event Filter](pkg/eventfilter/README.md) - Event filtering for transfers
- [Event Log Parser](pkg/eventlog/README.md) - Event log parsing
- [Token Types](pkg/tokens/types/README.md) - Core token data structures
- [Token Parsers](pkg/tokens/parsers/README.md) - Token list format parsers
- [Token Fetcher](pkg/tokens/fetcher/README.md) - HTTP token list fetching
- [Token AutoFetcher](pkg/tokens/autofetcher/README.md) - Automated background fetching
- [Token Builder](pkg/tokens/builder/README.md) - Incremental token collection building
- [Token Manager](pkg/tokens/manager/README.md) - High-level token management

### Example Documentation
- [Web Balance Fetcher](examples/balance-fetcher-web/README.md) - Web interface for balance fetching
- [Ethereum Client Usage](examples/ethclient-usage/README.md) - Ethereum client examples
- [Gas Comparison](examples/gas-comparison/README.md) - Gas fee comparison tool
- [Multicall Usage](examples/multiclient3-usage/README.md) - Multicall examples
- [Event Filter Example](examples/eventfilter-example/README.md) - Event filtering examples
- [Token Builder](examples/token-builder/README.md) - Token collection building
- [Token Fetcher](examples/token-fetcher/README.md) - Token list fetching
- [Token Manager](examples/token-manager/README.md) - Token management
- [Token Parser](examples/token-parser/README.md) - Token list parsing

### Specifications
- [Technical Specifications](docs/specs.md) - Complete SDK specifications and architecture

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

Mozilla Public License Version 2.0 - see [LICENSE](LICENSE)

---

**Built with ❤️ by the Status team**

