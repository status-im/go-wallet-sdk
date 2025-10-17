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
│   ├── contracts/         # Smart contract bindings
│   └── common/            # Shared utilities
├── examples/              # Usage examples
│   ├── balance-fetcher-web/       # Web interface for balance fetching
│   ├── ethclient-usage/           # Ethereum client examples
│   ├── gas-comparison/            # Gas fee comparison tool
│   ├── multiclient3-usage/        # Multicall examples
│   ├── multistandardfetcher-example/  # Multi-standard balance fetching
│   └── eventfilter-example/       # Event filtering examples
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

### Example Documentation
- [Web Balance Fetcher](examples/balance-fetcher-web/README.md) - Web interface for balance fetching
- [Ethereum Client Usage](examples/ethclient-usage/README.md) - Ethereum client examples
- [Gas Comparison](examples/gas-comparison/README.md) - Gas fee comparison tool
- [Multicall Usage](examples/multiclient3-usage/README.md) - Multicall examples
- [Event Filter Example](examples/eventfilter-example/README.md) - Event filtering examples

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

