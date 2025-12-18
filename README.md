[![Go Reference](https://pkg.go.dev/badge/github.com/status-im/go-wallet-sdk.svg)](https://pkg.go.dev/github.com/status-im/go-wallet-sdk) [![Checks](https://github.com/status-im/go-wallet-sdk/actions/workflows/checks.yml/badge.svg)](https://github.com/status-im/go-wallet-sdk/actions/workflows/checks.yml) [![codecov](https://codecov.io/gh/status-im/go-wallet-sdk/graph/badge.svg?branch=main)](https://app.codecov.io/gh/status-im/go-wallet-sdk) [![Go Report Card](https://goreportcard.com/badge/github.com/status-im/go-wallet-sdk)](https://goreportcard.com/report/github.com/status-im/go-wallet-sdk)

# Go Wallet SDK

A modular Go SDK for building multi-chain cryptocurrency wallets and blockchain applications.

## Requirements

- **Go**: 1.24.0 or higher
- **Optional**: RPC endpoint (Infura, Alchemy, or your own node) for examples

## Installation

```bash
go get github.com/status-im/go-wallet-sdk
```

For development and contributing:

```bash
git clone https://github.com/status-im/go-wallet-sdk
cd go-wallet-sdk
go mod download
```

## API reference

- GoDoc / API: https://pkg.go.dev/github.com/status-im/go-wallet-sdk
- Package READMEs: see the links in the table below

## Available Packages

| Area | Package | Use it when… | Key entrypoints (start here) |
|---|---|---|---|
| RPC client | [`pkg/ethclient`](pkg/ethclient/README.md) | You need chain-agnostic JSON-RPC (or a go-ethereum compatible surface) | `NewClient`, `Eth*` methods, `BalanceAt` |
| Balances | [`pkg/balance/fetcher`](pkg/balance/fetcher/README.md) | You need fast native/ERC20 balance reads with fallback strategies | `FetchNativeBalances`, `FetchErc20Balances` |
| Batching | [`pkg/multicall`](pkg/multicall/README.md) | You want to batch thousands of contract reads via Multicall3 | `Build*Call`, `RunSync`, `RunAsync` |
| Multi-standard balances | [`pkg/balance/multistandardfetcher`](pkg/balance/multistandardfetcher/README.md) | You want native+ERC20+ERC721+ERC1155 balances via one API | `FetchBalances`, `FetchConfig` |
| Gas | [`pkg/gas`](pkg/gas/README.md) | You need fee suggestions + inclusion estimates across L1/L2s | `GetTxSuggestions`, `GetChainSuggestions` |
| Transaction generation | [`pkg/txgenerator`](pkg/txgenerator/README.md) | You need to generate unsigned transactions for ETH/ERC20/ERC721/ERC1155 | `TransferETH`, `TransferERC20`, `ApproveERC20`, `TransferFromERC721`, `SafeTransferFromERC721`, `TransferERC1155` |
| Transfers | [`pkg/eventfilter`](pkg/eventfilter/README.md) | You need to efficiently query ERC20/721/1155 transfers via `eth_getLogs` | `FilterTransfers`, `TransferQueryConfig` |
| Log parsing | [`pkg/eventlog`](pkg/eventlog/README.md) | You need to detect/parse standard token events | `ParseLog`, `Event` |
| Accounts | [`pkg/accounts/extkeystore`](pkg/accounts/extkeystore/README.md) | You need HD (BIP32) keystore + signing | `NewKeyStore`, `DeriveWithPassphrase`, `SignHash` |
| Mnemonics | [`pkg/accounts/mnemonic`](pkg/accounts/mnemonic/README.md) | You need BIP39 mnemonics + seeds/extended keys | `CreateRandomMnemonic`, `CreateExtendedKeyFromMnemonic` |
| Token types | [`pkg/tokens/types`](pkg/tokens/types/README.md) | You need core token data structures and key generation | `Token`, `TokenList`, `TokenKey`, `IsNative` |
| Token parsers | [`pkg/tokens/parsers`](pkg/tokens/parsers/README.md) | You need to parse token lists from various formats | `StandardTokenListParser`, `StatusTokenListParser`, `CoinGeckoAllTokensParser` |
| Token fetcher | [`pkg/tokens/fetcher`](pkg/tokens/fetcher/README.md) | You need HTTP fetching with ETag caching and validation | `New`, `Fetch`, `FetchConcurrent` |
| Token autofetcher | [`pkg/tokens/autofetcher`](pkg/tokens/autofetcher/README.md) | You need automated background refresh of token lists | `NewAutofetcherFromTokenLists`, `NewAutofetcherFromRemoteListOfTokenLists` |
| Token builder | [`pkg/tokens/builder`](pkg/tokens/builder/README.md) | You need to incrementally build and merge token collections | `New`, `AddTokenList`, `AddNativeTokenList` |
| Token manager | [`pkg/tokens/manager`](pkg/tokens/manager/README.md) | You need high-level token management with auto-refresh | `New`, `Start`, `GetTokenByChainAddress`, `UniqueTokens` |
| ENS | [`pkg/ens`](pkg/ens/README.md) | You need forward/reverse ENS resolution | `NewResolver`, `AddressOf`, `GetName`, `IsSupportedChain` |

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

The C library exposes core SDK functionality through a C-compatible API, including:
- Ethereum client operations (RPC calls, chain ID, balances)
- Multi-standard balance fetching (Native ETH, ERC20, ERC721, ERC1155)
- Transaction generation (ETH transfers, ERC20, ERC721, ERC1155 operations)
- Account management (extended keystore and standard keystore)
- Mnemonic generation and key derivation utilities

See [examples/c-app](examples/c-app/README.md) for a complete C usage example.

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

### Transaction Generator Example

```bash
cd examples/txgenerator-example
go run .
```

Access: http://localhost:8080

Web interface for generating unsigned Ethereum transactions for ETH transfers, ERC20, ERC721, and ERC1155 operations.

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

## Documentation

### Package Documentation
- [Balance Fetcher](pkg/balance/fetcher/README.md) - Balance fetching functionality
- [Multicall](pkg/multicall/README.md) - Efficient contract call batching
- [Ethereum Client](pkg/ethclient/README.md) - Complete Ethereum RPC client
- [Gas Estimation](pkg/gas/README.md) - Gas fee estimation and suggestions
- [Transaction Generator](pkg/txgenerator/README.md) - Generate unsigned transactions for ETH/ERC20/ERC721/ERC1155
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
- [Transaction Generator Example](examples/txgenerator-example/README.md) - Web interface for generating transactions
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

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Code style and conventions
- Testing requirements
- Pull request process
- Development workflow

## License

Mozilla Public License Version 2.0 - see [LICENSE](LICENSE)

---

Built by the Status team.

