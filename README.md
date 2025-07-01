# Go Wallet SDK

A modular Go SDK for building multi-chain cryptocurrency wallets and blockchain applications.

## Quick Start

```bash
go get github.com/status-im/go-wallet-sdk
```

## Available Packages

### Balance Management
- **`pkg/balance/fetcher`**: High-performance balance fetching with automatic fallback strategies
  - Batch processing for multiple addresses
  - Smart fallback between different fetching methods
  - Chain-agnostic design

### Common Utilities
- **`pkg/common`**: Shared utilities and constants used across the SDK

## Examples

### Web-Based Balance Fetcher

```bash
cd examples/balance-fetcher-web
go run .
```

Access: http://localhost:8080

## Testing

```bash
go test ./...
```

## Project Structure

```
go-wallet-sdk/
├── pkg/                    # Core SDK packages
│   ├── balance/           # Balance-related functionality
│   └── common/            # Shared utilities
├── examples/              # Usage examples
└── README.md             # This file
```

## Documentation

- [Balance Fetcher](pkg/balance/fetcher/README.md) - Balance fetching functionality
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

