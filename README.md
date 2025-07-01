# Go Wallet SDK

A modular Go SDK for building multi-chain cryptocurrency wallets and blockchain applications. This SDK provides production-ready components for interacting with Ethereum and other EVM-compatible blockchains.

## 🚀 Overview

The Go Wallet SDK is designed to simplify the development of blockchain applications by providing:

- **Modular Components**: Independent packages that can be used separately or together
- **Multi-Chain Support**: Works with Ethereum and any EVM-compatible blockchain
- **Well Documented**: Clear documentation and examples for each package
- **Extensible**: Easy to extend and customize for specific use cases

## 📦 Available Packages

### Balance Management
- **`pkg/balance/fetcher`**: High-performance balance fetching with automatic fallback strategies
  - Batch processing for multiple addresses
  - Smart fallback between different fetching methods
  - Chain-agnostic design

### Common Utilities
- **`pkg/common`**: Shared utilities and constants used across the SDK
  - Chain identifiers and constants
  - Common data structures

## 🏗️ Project Structure

```
go-wallet-sdk/
├── pkg/                    # Core SDK packages
│   ├── balance/           # Balance-related functionality
│   └── common/            # Shared utilities
├── examples/              # Usage examples and applications
├── internal/              # Internal implementation details
└── README.md             # This file
```

## 🚀 Quick Start

### Installation

```bash
go get github.com/status-im/go-wallet-sdk
```

### Basic Usage

Each package provides its own specific functionality. See the package-specific READMEs for detailed usage examples:

- [Balance Fetcher](pkg/balance/fetcher/README.md) - Fetch balances for multiple addresses
- [Common Utilities](pkg/common/) - Shared constants and utilities

## 🌐 Examples

### Web-Based Balance Fetcher

A complete web application demonstrating the balance fetcher package:

```bash
cd examples/balance-fetcher-web
go run .
```

**Features:**
- Modern web interface
- Support for any EVM-compatible chain
- Dynamic chain management
- Modular code structure

**Access:** http://localhost:8080

See the [example README](examples/balance-fetcher-web/README.md) for detailed documentation.

## 🧪 Testing

The SDK includes comprehensive tests for all packages:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## 🔧 Development

### Prerequisites

- Go 1.23 or later
- Access to Ethereum RPC endpoints (for testing)

### Building

```bash
# Build all packages
go build ./...

# Build specific example
go build -o balance-fetcher-web examples/balance-fetcher-web/
```

## 📚 Documentation

Each package includes detailed documentation:

- [Balance Fetcher](pkg/balance/fetcher/README.md) - Balance fetching functionality
- [Web Example](examples/balance-fetcher-web/README.md) - Complete web application
- [API Documentation](https://pkg.go.dev/github.com/status-im/go-wallet-sdk) - Go package documentation

## 🤝 Contributing

We welcome contributions! Please see our contributing guidelines:

1. **Fork** the repository
2. **Create** a feature branch
3. **Make** your changes
4. **Add** tests for new functionality
5. **Update** documentation
6. **Submit** a pull request

### Code Style

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for complex logic
- Ensure all code is tested

## 📄 License

This project is licensed under the Mozilla Public License Version 2.0 - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: Report bugs and request features on [GitHub Issues](https://github.com/status-im/go-wallet-sdk/issues)
- **Discussions**: Join the conversation on [GitHub Discussions](https://github.com/status-im/go-wallet-sdk/discussions)
- **Documentation**: Check the package-specific README files for detailed usage information

## 🔗 Related Projects

- [Status](https://status.im/) - Privacy-focused messaging and Web3 browser
- [status-go](https://github.com/status/status-go) - Backbone library of Status client, consumer of Go Wallet SDK

---

**Built with ❤️ by the Status team**

