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

### Account Management
- **`pkg/accounts/extkeystore`**: Extended keystore with HD wallet support
  - BIP32 hierarchical deterministic wallet functionality
  - Encrypted storage following Web3 Secret Storage specification
  - Child account derivation using BIP44 derivation paths
  - Import/export extended keys and standard private keys
  - Full account lifecycle management (create, unlock, lock, sign, delete)

- **`pkg/accounts/mnemonic`**: BIP39 mnemonic phrase utilities
  - Generate cryptographically secure mnemonic phrases (12, 15, 18, 21, 24 words)
  - Create BIP32 extended keys from mnemonic phrases
  - BIP39 passphrase support for seed extension
  - Seamless integration with extkeystore

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

### ENS Resolution
- **`pkg/ens`**: Ethereum Name Service (ENS) resolution
  - Forward resolution (ENS name → Ethereum address)
  - Reverse resolution (Ethereum address → ENS name)
  - Chain support detection via `IsSupportedChain()`
  - Works on any chain where ENS registry is deployed

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

### Accounts Example

```bash
cd examples/accounts
go run .
```

Access: http://localhost:8081

Interactive web interface for testing extkeystore and standard keystore functionality, including mnemonic generation, account creation, derivation, import/export, and signing.

### C Application Example

```bash
# Build the shared library first
make shared-library

# Then build and run the C example
cd examples/c-app
make build
cd bin
./c-app
```

Demonstrates how to use the Go Wallet SDK from C applications using the shared library. The example includes:
- Ethereum client operations (creating clients, retrieving chain ID, fetching balances, making JSON-RPC calls)
- Multi-standard balance fetching (Native ETH, ERC20, ERC721, ERC1155)
- Account management with extended keystore and standard keystore
- Mnemonic generation and key derivation
- Account creation, import/export, signing, and derivation

## Building the C Library

The SDK can be compiled as a C library (shared or static) for use in non-Go applications:

**Shared Library:**
```bash
make shared-library
```

This creates:
- `build/libgowalletsdk.dylib` (macOS) or `build/libgowalletsdk.so` (Linux)
- `build/libgowalletsdk.h` (C header file)

**Static Library:**
```bash
make static-library
```

This creates:
- `build/libgowalletsdk.a` (static library)
- `build/libgowalletsdk.h` (C header file)

The shared library exposes core SDK functionality through a C-compatible API, including:
- Ethereum client operations (RPC calls, chain ID, balances)
- Multi-standard balance fetching (Native ETH, ERC20, ERC721, ERC1155)
- Account management (extended keystore and standard keystore)
- Mnemonic generation and key derivation utilities

See [examples/c-app](examples/c-app/README.md) for a complete C usage example.

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

### ENS Resolver Example

```bash
cd examples/ens-resolver-example

# Forward resolution (ENS name to address)
go run . -rpc https://eth.llamarpc.com -name vitalik.eth

# Reverse resolution (address to ENS name)
go run . -rpc https://eth.llamarpc.com -address 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
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
│   ├── accounts/          # Account management (extkeystore, mnemonic)
│   ├── tokens/            # Token management system
│   │   ├── types/         # Core token data structures
│   │   ├── parsers/       # Token list format parsers
│   │   ├── fetcher/       # HTTP token list fetching
│   │   ├── autofetcher/   # Automated background fetching
│   │   ├── builder/       # Incremental token collection building
│   │   └── manager/       # High-level token management
│   ├── ens/               # ENS name resolution
│   ├── contracts/         # Smart contract bindings
│   └── common/            # Shared utilities
├── clib/                 # C library bindings (shared and static)
│   ├── c.go             # Memory management utilities
│   ├── ethclient.go     # Ethereum client C bindings
│   ├── balance_multistandardfetcher.go  # Multi-standard balance fetching C bindings
│   ├── accounts_extkeystore.go  # Extended keystore C bindings
│   ├── accounts_keystore.go     # Standard keystore C bindings
│   ├── accounts_keys.go         # Key derivation and conversion C bindings
│   ├── accounts_mnemonic.go     # Mnemonic utilities C bindings
│   └── main.go          # Entry point for library build
├── examples/              # Usage examples
│   ├── balance-fetcher-web/       # Web interface for balance fetching
│   ├── ethclient-usage/           # Ethereum client examples
│   ├── gas-comparison/            # Gas fee comparison tool
│   ├── multiclient3-usage/        # Multicall examples
│   ├── multistandardfetcher-example/  # Multi-standard balance fetching
│   ├── eventfilter-example/       # Event filtering examples
│   ├── accounts/                  # Keystore management web interface
│   ├── c-app/                    # C application example
│   ├── token-builder/             # Token collection building
│   ├── token-fetcher/            # Token list fetching
│   ├── token-manager/             # Token management
│   ├── token-parser/              # Token list parsing
│   └── ens-resolver-example/     # ENS resolution CLI tool
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
- [Extended Keystore](pkg/accounts/extkeystore/README.md) - HD wallet keystore with BIP32 support
- [Mnemonic](pkg/accounts/mnemonic/README.md) - BIP39 mnemonic phrase utilities
- [Token Types](pkg/tokens/types/README.md) - Core token data structures
- [Token Parsers](pkg/tokens/parsers/README.md) - Token list format parsers
- [Token Fetcher](pkg/tokens/fetcher/README.md) - HTTP token list fetching
- [Token AutoFetcher](pkg/tokens/autofetcher/README.md) - Automated background fetching
- [Token Builder](pkg/tokens/builder/README.md) - Incremental token collection building
- [Token Manager](pkg/tokens/manager/README.md) - High-level token management
- [ENS Resolver](pkg/ens/README.md) - ENS name resolution

### Example Documentation
- [Web Balance Fetcher](examples/balance-fetcher-web/README.md) - Web interface for balance fetching
- [Ethereum Client Usage](examples/ethclient-usage/README.md) - Ethereum client examples
- [Gas Comparison](examples/gas-comparison/README.md) - Gas fee comparison tool
- [Multicall Usage](examples/multiclient3-usage/README.md) - Multicall examples
- [Event Filter Example](examples/eventfilter-example/README.md) - Event filtering examples
- [Accounts Example](examples/accounts/README.md) - Keystore management web interface
- [C Application Example](examples/c-app/README.md) - C application using the shared library
- [Token Builder](examples/token-builder/README.md) - Token collection building
- [Token Fetcher](examples/token-fetcher/README.md) - Token list fetching
- [Token Manager](examples/token-manager/README.md) - Token management
- [Token Parser](examples/token-parser/README.md) - Token list parsing
- [ENS Resolver Example](examples/ens-resolver-example/README.md) - ENS resolution CLI tool

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

