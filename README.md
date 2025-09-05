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

### Ethereum Client
- **`pkg/ethclient`**: Full-featured Ethereum client with go-ethereum compatibility
  - Complete RPC method coverage (eth_*, net_*, web3_*)
  - Go-ethereum ethclient compatible interface for easy migration

### Token Lists Management
- **`pkg/tokenlists`**: Comprehensive token list management with privacy-aware fetching
  - Multi-source support (Status, Uniswap, CoinGecko, custom sources)
  - Privacy-aware automatic refresh with ETag support
  - Cross-chain token management and validation
  - Extensible parser system for custom token list formats
  - Thread-safe concurrent access with proper synchronization

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

## Testing

```bash
go test ./...
```

## Project Structure

```
go-wallet-sdk/
├── pkg/                    # Core SDK packages
│   ├── balance/           # Balance-related functionality
│   ├── ethclient/         # Ethereum client with full RPC support
│   ├── tokenlists/        # Token list management and fetching
│   └── common/            # Shared utilities
├── examples/              # Usage examples
└── README.md             # This file
```

## Documentation

- [Balance Fetcher](pkg/balance/fetcher/README.md) - Balance fetching functionality
- [Ethereum Client](pkg/ethclient/README.md) - Complete Ethereum RPC client
- [Token Lists](pkg/tokenlists/README.md) - Token list management and fetching
- [Web Example](examples/balance-fetcher-web/README.md) - Complete web application

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

