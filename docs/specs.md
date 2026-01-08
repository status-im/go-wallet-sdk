# Go Wallet SDK – Technical Specs

This document is a deep-dive spec for the SDK architecture and public surfaces.

Quick navigation:

- Primary entrypoint and module index: `README.md`
- GoDoc / API reference: https://pkg.go.dev/github.com/status-im/go-wallet-sdk
- Package READMEs:
    - `pkg/ethclient/README.md`
    - `pkg/balance/fetcher/README.md`
    - `pkg/balance/multistandardfetcher/README.md`
    - `pkg/multicall/README.md`
- `pkg/gas/README.md`
- `pkg/txgenerator/README.md`
- `pkg/eventfilter/README.md`
- `pkg/eventlog/README.md`
    - `pkg/accounts/extkeystore/README.md`
    - `pkg/accounts/mnemonic/README.md`
    - `pkg/tokens/types/README.md`
    - `pkg/tokens/parsers/README.md`
    - `pkg/tokens/fetcher/README.md`
    - `pkg/tokens/autofetcher/README.md`
    - `pkg/tokens/builder/README.md`
    - `pkg/tokens/manager/README.md`
    - `pkg/ens/README.md`

## 1. Overview and Goals

Go Wallet SDK is a modular Go library intended to support the development of multi‑chain cryptocurrency wallets and blockchain applications. The SDK exposes self‑contained packages for common wallet tasks such as fetching account balances across many EVM chains and interacting with Ethereum JSON‑RPC.

### 1.1 Main Repository Components

| Component             | Purpose                                                                                                                                                                                                                                                    |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `pkg/balance/fetcher` | High‑performance balance fetcher for EVM‑compatible chains.  The package can fetch native token balances or ERC‑20 balances for multiple addresses in batches using smart fallback strategies. It includes automatic fallback strategies (Multicall3 contract or standard RPC batching) and exposes simple APIs to fetch balances for many addresses or tokens                                                             |
| `pkg/multicall`       | Efficient batching of multiple Ethereum contract calls into single transactions using Multicall3. Supports native ETH, ERC20, ERC721, and ERC1155 balance queries with chunked processing, error handling, and both synchronous and asynchronous execution modes. |
| `pkg/ethclient`       | Chain‑agnostic Ethereum JSON‑RPC client.  It provides two method sets: a drop‑in replacement compatible with go‑ethereum's `ethclient` and a custom implementation that follows the Ethereum JSON‑RPC specification without assuming chain‑specific types. It supports JSON‑RPC methods covering `eth_`, `net_` and `web3_` namespace |
| `pkg/gas`             | Comprehensive gas estimation and fee suggestion package for Ethereum and L2 networks. Provides smart fee estimation with priority fees, base fees, max fees, and inclusion time estimates. Supports multiple chain classes including L1 (Ethereum, Polygon, BSC), Arbitrum Stack, Optimism Stack, and Linea Stack with chain-specific optimizations. |
| `pkg/txgenerator`     | Transaction generation package for creating unsigned Ethereum transactions. Supports ETH transfers, ERC20 transfers and approvals, ERC721 transfers (transferFrom and safeTransferFrom), approvals, and operator management, as well as ERC1155 single and batch transfers and operator management. Automatically detects transaction type (legacy or EIP-1559) based on provided gas parameters. |
| `pkg/eventfilter`     | Efficient filtering for Ethereum transfer events across ERC20, ERC721, and ERC1155 tokens. Minimizes `eth_getLogs` API calls while capturing all relevant transfers involving specified addresses with optimized query generation and direction-based filtering. |
| `pkg/eventlog`        | Ethereum event log parser for ERC20, ERC721, and ERC1155 events. Automatically detects and parses token events with type-safe access to event data, supporting Transfer, Approval, and other standard token events. |
| `pkg/common`          | Shared types and constants. Such as canonical chain IDs (e.g., Ethereum Mainnet, Optimism, Arbitrum, BSC, Base). Developers use these values when configuring the SDK or examples.                               |
| `pkg/contracts/`      | Solidity contracts and Go bindings for smart contract interactions. Includes Multicall3, ERC20, ERC721, and ERC1155 contracts with deployment addresses for multiple chains. |
| `pkg/accounts/extkeystore` | Extended keystore for Ethereum accounts with BIP32 hierarchical deterministic (HD) wallet support. Stores BIP32 extended keys instead of just private keys, enabling derivation of child accounts from parent keys. Provides encrypted storage following Web3 Secret Storage specification, account management (create, unlock, lock, sign, delete), and import/export functionality for both extended keys and standard private keys. |
| `pkg/accounts/mnemonic` | Utilities for generating BIP39 mnemonic phrases and creating extended keys from them. Provides functions to create random mnemonics (12, 15, 18, 21, or 24 words) and derive BIP32 extended keys from existing phrases with optional BIP39 passphrase support. |
| `pkg/ens`             | Ethereum Name Service (ENS) resolution package. Supports forward resolution (ENS name to Ethereum address) and reverse resolution (Ethereum address to ENS name). Uses go-ens/v3 library internally. Provides `IsSupportedChain()` to check if ENS is available on Mainnet, Sepolia, or Holesky. |
| `clib/`               | C library bindings (shared and static) that expose core SDK functionality to non-Go applications. Provides C-compatible functions for Ethereum client operations including creating clients, fetching chain IDs, retrieving account balances, fetching multi-standard token balances (Native ETH, ERC20, ERC721, ERC1155), and generating unsigned transactions for ETH, ERC20, ERC721, and ERC1155 operations. The package can be compiled as a shared library (.so/.dylib) or static library (.a) with a generated C header file. |
| `examples/`           | Demonstrations of SDK usage.  Includes `balance-fetcher-web` (a web interface for batch balance fetching), `ethclient‑usage` (an example that exercises the Ethereum client across multiple RPC endpoints), `multiclient3-usage` (demonstrates multicall functionality), `multistandardfetcher-example` (shows multi-standard balance fetching across all token types), `eventfilter-example` (shows event filtering and parsing capabilities), `gas-comparison` (compares gas estimation implementations across multiple networks), `txgenerator-example` (a web interface for generating unsigned Ethereum transactions for ETH, ERC20, ERC721, and ERC1155 operations), `accounts` (an interactive web interface for testing extkeystore and standard keystore functionality including mnemonic generation, account creation, derivation, import/export, and signing), `c-app` (a C application example demonstrating how to use the shared library from C code), and `ens-resolver-example` (a CLI tool for ENS forward and reverse resolution). |

## 2. Architecture

### 2.1 High‑level Structure

Go Wallet SDK follows a modular architecture where each package encapsulates a specific functional domain. There is no central runtime; instead applications import only the packages they need. The SDK currently focuses on EVM‑compatible chains, leaving room for additional chain types in the future. The packages are:
- **Balance Fetcher** – Provides efficient methods to retrieve account balances (native or ERC‑20) across many addresses and tokens. It abstracts over RPC batch calls and Multicall3 contract calls. Developers supply a minimal RPC client interface (`RPCClient` and optionally `BatchCaller`) and the package returns a map of balances
- **Multicall** – Efficiently batches multiple Ethereum contract calls into single transactions using Multicall3. Supports native ETH, ERC20, ERC721, and ERC1155 balance queries with automatic chunking, error handling, and both synchronous and asynchronous execution modes. Provides call builders and result processors for different token types.
- **Ethereum Client** – Exposes the full Ethereum JSON‑RPC API. It wraps a standard RPC client and offers two sets of methods: chain‑agnostic versions prefixed with `Eth*` and a drop‑in `BalanceAt`, `BlockNumber` etc. that mirror go‑ethereum's ethclient. The client covers methods including network info, block and transaction queries, account state, contract code and gas estimation
- **Gas Estimation** – Provides comprehensive gas fee estimation and suggestions for Ethereum and L2 networks. Analyzes historical fee data to suggest optimal priority fees, base fees, and max fees for three priority levels (low, medium, high). Estimates transaction inclusion time based on network congestion and chain parameters. Supports multiple chain classes with specific optimizations for L1, Arbitrum Stack, Optimism Stack, and Linea Stack.
- **Transaction Generator** – Provides utilities for generating unsigned Ethereum transactions for ETH transfers, ERC20 operations (transfers and approvals), ERC721 operations (transfers, approvals, operator management), and ERC1155 operations (single and batch transfers, operator management). Automatically detects transaction type (legacy or EIP-1559) based on provided gas parameters and returns unsigned `types.Transaction` objects ready for signing.
- **Event Filter** – Efficiently filters Ethereum transfer events across ERC20, ERC721, and ERC1155 tokens. Minimizes `eth_getLogs` API calls through optimized query generation and supports direction-based filtering (send, receive, or both). Uses intelligent query merging to reduce the number of RPC calls required.
- **Event Log Parser** – Automatically detects and parses Ethereum event logs for ERC20, ERC721, and ERC1155 tokens. Provides type-safe access to event data with support for Transfer, Approval, and other standard token events. Works seamlessly with the Event Filter package.
- **Extended Keystore** – An enhanced keystore that stores BIP32 extended keys instead of just private keys, enabling hierarchical deterministic (HD) wallet functionality. Supports derivation of child accounts from parent keys using BIP44 derivation paths, encrypted storage following Web3 Secret Storage specification, and full account lifecycle management (create, unlock, lock, sign, delete). Can import/export both extended keys and standard private keys, making it compatible with existing keystore implementations.
- **Mnemonic Utilities** – Simple package for working with BIP39 mnemonic seed phrases to generate deterministic wallets. Provides functions to create random mnemonics (12, 15, 18, 21, or 24 words) and derive BIP32 extended keys from existing phrases with optional BIP39 passphrase support. Designed to work seamlessly with the Extended Keystore package.
- **Common Utilities** – Houses shared types (e.g., `ChainID`) and enumerated constants for well‑known networks. This allows examples and client code to refer to network IDs without hard‑coding numbers.
- **Contract Bindings** – Provides Go bindings for smart contracts including Multicall3, ERC20, ERC721, and ERC1155. Includes deployment addresses for multiple chains and utilities for contract interaction.
- **C Library** – Exposes core SDK functionality to non-Go applications through C-compatible bindings. The `clib` package can be compiled as a shared library (.so/.dylib) or static library (.a) with a generated C header file, enabling integration with C, C++, and other languages that can call C functions. Provides memory-safe wrappers for Ethereum client operations, multi-standard balance fetching, account management (extended keystore and standard keystore), mnemonic generation, and key derivation utilities with proper resource management. String conversion utilities for CollectibleID types are implemented internally for JSON serialization.
- **Token Types** – Core data structures for tokens and token lists with unified representation, cross-chain support, type-safe address handling, and validation. Provides Token and TokenList types that serve as the foundation for all token-related operations.
- **Token Parsers** – Token list parsing implementations for multiple formats including Standard (Uniswap-style), Status-specific with chain grouping, CoinGecko API with platform mappings, and list-of-token-lists metadata parsing. Supports chain filtering and validation with extensible parser architecture.
- **Token Fetcher** – HTTP-based token list fetching with concurrent operations, HTTP ETag caching for bandwidth efficiency, JSON schema validation support, and robust error handling with timeout management. Designed for production use with configurable HTTP client settings.
- **Token AutoFetcher** – Automated background token list management with configurable refresh intervals, thread-safe operations with context support, pluggable storage backends, and error reporting via channels. Supports both direct token list fetching and remote list-of-token-lists discovery patterns.
- **Token Builder** – Incremental token collection building using the Builder pattern with automatic deduplication by chain ID and address, native token generation for supported chains, and multiple format support through parsers. Provides stateful construction with deterministic ordering.
- **Token Manager** – High-level token management interface providing multi-source token integration (native, remote, local, custom), thread-safe concurrent access, rich query capabilities by chain/address/list ID, and automatic refresh with state management. Centralizes token operations for wallet applications.
- **ENS Resolver** – Ethereum Name Service resolution package for converting between ENS names and Ethereum addresses. Supports both forward resolution (name to address) and reverse resolution (address to name). Uses go-ens/v3 library and provides dynamic chain support detection via contract existence check.

The SDK emphasises chain agnosticism: methods do not assume particular transaction formats or gas pricing models and therefore work with Ethereum, L2 networks (Optimism, Arbitrum, Polygon), and other EVM‑compatible chains. Each package hides chain‑specific details behind simple interfaces.

### 2.2 Balance Fetcher Design

The balance fetcher is designed to efficiently query balances for many addresses and tokens. Its design includes:

- **Dual fetch strategies** – The package first attempts to use Multicall3 contract calls to retrieve multiple balances in a single transaction. If Multicall3 is unavailable on a given chain or the call fails, it falls back to batch RPC calls that iterate through addresses/tokens. Both strategies are exposed transparently through the same API.
- **Batching and concurrency** – When using Multicall3, the fetcher groups requests into batches (configurable `batchSize`) to reduce the number of round‑trips. When falling back to RPC, it also groups requests into batches and processes them in parallel when possible, aggregating results into a map keyed by address and token.
- **Chain‑agnostic** – The logic is unaware of specific chain parameters; it accepts any RPC endpoint and optionally a block number. A `ChainID` from `pkg/common` can be used to label results, but the fetcher does not require it.

### 2.3 Multicall Design

The multicall package is designed to efficiently batch multiple Ethereum contract calls into single transactions using Multicall3. Its design includes:

- **Call Builders** – Provides functions to build calls for different token types (native ETH, ERC20, ERC721, ERC1155). Each builder creates properly encoded call data for the specific contract function.
- **Job-based System** – Uses a flexible job system where each job contains a set of calls and a result processing function. Supports both synchronous (`RunSync`) and asynchronous (`RunAsync`) execution modes. The system automatically chunks large call sets into manageable batches to avoid transaction size limits.
- **Error Handling** – Graceful failure handling with detailed error reporting. Individual call failures don't cause the entire batch to fail, allowing partial results to be processed. Each job can have its own error handling strategy.
- **Result Processing** – Each job specifies its own result processing function (`CallResultFn`) that decodes the raw return data into appropriate Go types. Provides dedicated result processors for each token type that decode the raw return data into appropriate Go types (`*big.Int` for balances).
- **Chain Support** – Works with any EVM-compatible chain that has Multicall3 deployed, with automatic address resolution based on chain ID.

### 2.4 Ethereum Client Design

The Ethereum client package (`pkg/ethclient`) wraps a generic RPC client and exposes two categories of methods:

- **Go‑ethereum‑compatible methods** – Methods such as `BlockNumber`, `BalanceAt` and `TransactionByHash` mimic the ethclient interface from go‑ethereum so existing applications can switch to this SDK with minimal changes. These methods require a go‑ethereum RPC client (because they call underlying types) and may not work on Layer 2 chains that diverge from Ethereum’s API.
- **Chain‑agnostic methods** – Methods prefixed with Eth* correspond directly to Ethereum JSON‑RPC calls and accept/return standard Go types. Examples include `EthBlockNumber`, `EthGetBalance`, `EthGasPrice`, `EthGetBlockByNumberWithFullTxs`, `EthGetLogs`, and `EthEstimateGas`. These functions rely only on the JSON‑RPC specification and therefore support any EVM‑compatible chain.

Internally, the client stores a reference to an RPC client and implements each method by calling `rpcClient.CallContext` with the appropriate RPC method name and parameters (see eth.go). It deserialises responses into exported Go types or custom structs (e.g., `BlockWithTxHashes`, `BlockWithFullTxs`). The design includes convenience functions for converting block numbers to RPC arguments and decoding hex‑encoded values.

### 2.5 Common Utilities

The `pkg/common` package defines shared types and enumerations. The main export is `type ChainID uint64` with constants for well‑known networks such as `EthereumMainnet`, `EthereumSepolia`, `OptimismMainnet`, `ArbitrumMainnet`, `BSCMainnet`, `BaseMainnet`, `BaseSepolia` and a custom `StatusNetworkSepolia`. These constants allow the examples to pre‑populate supported chains and label results without repeating numeric IDs.

### 2.6 Event Filter Design

The event filter package (`pkg/eventfilter`) is designed to efficiently query Ethereum transfer events while minimizing API calls:

- **Multi-Token Support** – Supports ERC20, ERC721, and ERC1155 transfer events through a unified interface. Uses standardized event signatures from the `eventlog` package for consistent event detection.
- **Direction-Based Filtering** – Allows filtering by transfer direction (send, receive, or both) by constructing appropriate topic filters. Send direction queries match addresses in the 'from' field, while receive direction queries match addresses in the 'to' field.
- **Query Optimization** – Intelligently merges compatible queries to minimize the number of `eth_getLogs` calls. ERC20 and ERC721 transfers share the same event signature and can be combined in single queries. The package uses OR operations to merge multiple event types when possible.
- **Topic Structure Optimization** – Constructs efficient topic filters by omitting empty trailing topics and using appropriate topic positions for different event types. ERC20/ERC721 transfers use 2-3 topics while ERC1155 transfers use 3-4 topics depending on direction.
- **Chain-Agnostic** – Works with any EVM-compatible chain by using standard event signatures and topic structures. No chain-specific logic or assumptions.

### 2.7 Event Log Parser Design

The event log parser package (`pkg/eventlog`) provides automatic detection and parsing of Ethereum event logs:

- **Multi-Contract Support** – Automatically detects and parses events from ERC20, ERC721, and ERC1155 contracts using a registry-based approach. Each contract type has dedicated parsers that handle their specific event structures.
- **Type-Safe Access** – Provides strongly-typed access to parsed event data through the `Unpacked` field. Each event type is parsed into its corresponding Go struct (e.g., `Erc20Transfer`, `Erc721Transfer`, `Erc1155TransferSingle`).
- **Event Detection** – Uses event signatures and topic patterns to identify event types. Supports all standard token events including Transfer, Approval, ApprovalForAll, and URI events.
- **Integration** – Designed to work seamlessly with the Event Filter package. The `FilterTransfers` function returns parsed events ready for application use.
- **Error Handling** – Gracefully handles unknown or malformed events by returning empty slices. Safe to use with any log data without causing panics.

### 2.8 Extended Keystore Design

The `pkg/accounts/extkeystore` package provides an enhanced keystore with hierarchical deterministic (HD) wallet support:

- **BIP32 Extended Keys** – Stores BIP32 extended keys instead of just private keys, enabling derivation of child accounts from parent keys. This allows a single master key to generate unlimited child accounts following BIP44 derivation paths.
- **Encrypted Storage** – Keys are stored as encrypted JSON files following the Web3 Secret Storage specification. Supports both light scrypt parameters (fast, for development) and standard scrypt parameters (secure, for production).
- **Child Account Derivation** – Derives child accounts from parent keys using BIP44 derivation paths. Supports both ephemeral derivation (for signing without storing) and pinned derivation (saves derived keys to keystore).
- **Import/Export** – Can import extended keys or standard private keys; can export in both formats. This allows compatibility with existing keystore implementations and seamless migration between different keystore types.
- **Account Management** – Full lifecycle support: create new accounts, unlock/lock accounts with optional timeout, sign transactions and messages, and delete accounts with passphrase confirmation.
- **Based on go-ethereum** – Derived from go-ethereum's keystore implementation, modified to store extended keys instead of private keys while maintaining API compatibility where possible.

### 2.9 Mnemonic Utilities Design

The `pkg/accounts/mnemonic` package provides utilities for working with BIP39 mnemonic phrases:

- **Random Generation** – Generates cryptographically secure mnemonic phrases with configurable lengths (12, 15, 18, 21, or 24 words). Each length corresponds to a specific entropy strength following BIP39 specification.
- **Extended Key Creation** – Creates BIP32 master extended keys from mnemonic phrases using the BIP39 seed derivation process. Supports optional passphrase (BIP39 seed extension) for additional security.
- **Integration** – Designed to work seamlessly with the Extended Keystore package. Typical workflow: generate mnemonic → create extended key → import into keystore → derive child accounts.

### 2.10 Contract Bindings

The `pkg/contracts` package provides Go bindings for smart contracts and deployment utilities:

- **Multicall3** – Provides bindings for the Multicall3 contract with deployment addresses for 200+ chains. Includes utilities for address resolution and chain support checking.
- **Token Standards** – ERC20, ERC721, and ERC1155 contract bindings with standard interface implementations.
- **Deployment Management** – Automated deployment address management with utilities to regenerate addresses from official deployment lists.

### 2.11 Gas Estimation Design

The `pkg/gas` package provides comprehensive gas fee estimation and suggestions for Ethereum and L2 networks:

- **Multi-Chain Support** – Supports four chain classes with specific optimization strategies:
  - **L1 (Ethereum, Polygon, BSC)**: Uses congestion-based base fee multipliers for dynamic fee adjustment (1.025x base with 10x congestion factor for medium/high)
  - **ArbStack (Arbitrum)**: Fast 0.25s block times with fixed multipliers (1.025x, 4.1x, 10.25x) for L2 optimization
  - **OPStack (Optimism, Base)**: Fixed base fee multipliers (1.025x, 4.1x, 10.25x) for predictable fees
  - **LineaStack (Linea)**: Uses dedicated `linea_estimateGas` RPC method with 2x base fee for all levels

- **Fee Calculation** – Analyzes historical fee data from `eth_feeHistory` to calculate three priority levels:
  - **Low Priority**: Uses 10th percentile of historical priority fees, base fee with 1.025x multiplier (no congestion adjustment on L1)
  - **Medium Priority**: Uses 45th percentile with configurable base fee multipliers (1.025x for L1, 4.1x for L2) and optional congestion adjustment (10x factor on L1)
  - **High Priority**: Uses 90th percentile with higher base fee multipliers (1.025x for L1, 10.25x for L2) and optional congestion adjustment (10x factor on L1)

- **Inclusion Time Estimation** – Estimates transaction inclusion time based on:
  - Historical base fees and priority fees from recent blocks
  - Chain-specific block times (12s for Ethereum, 2s for L2s, 0.25s for Arbitrum)
  - Fee competitiveness relative to network conditions
  - Returns min/max blocks and min/max seconds until inclusion

- **Network Congestion** – Calculates congestion score (0-1 scale) for L1 chains by analyzing:
  - Average base fee trends
  - Average priority fee levels
  - Gas usage ratios across recent blocks
  - Weighted scoring of priority fees (70%) and gas usage (30%)

- **Configurable Parameters** – Developers can customize:
  - Number of blocks for congestion analysis (default: 10)
  - Number of blocks for gas price estimation (default: 10 for L1, 50 for L2)
  - Reward percentiles for low/medium/high priority (10/45/90)
  - Base fee multipliers per priority level (default: 1.025 for L1, 1.025/4.1/10.25 for L2)
  - Congestion-based adjustment factors for L1 chains (default: 0.0/10.0/10.0)
  - Fine-grained control over fee calculation for each priority level

### 2.12 Token Types Design

The `pkg/tokens/types` package provides the foundational data structures for the entire token management system:

- **Unified Token Representation** – The `Token` struct represents tokens across all blockchain networks with cross-chain support through `CrossChainID`, allowing tokens to be grouped across different chains (e.g., USDC on Ethereum and BSC).
- **Type-Safe Address Handling** – Uses `gethcommon.Address` for Ethereum addresses with automatic validation and normalization, ensuring addresses are properly formatted and checksummed.
- **Token List Metadata** – The `TokenList` struct includes comprehensive metadata including name, timestamp, version, source URL, and schema information for validation and origin tracking.
- **Custom Token Support** – Distinguishes between official tokens from curated lists and user-added custom tokens through the `CustomToken` flag, enabling personalized token management.
- **Deterministic Key Generation** – Tokens are uniquely identified using `TokenKey(chainID, address)` which creates consistent keys for deduplication and lookup operations.

### 2.13 Token Parsers Design

The `pkg/tokens/parsers` package provides extensible parsing for multiple token list formats:

- **Multi-Format Support** – Supports Standard (Uniswap-style), Status-specific with chain grouping, CoinGecko API with platform mappings, and list-of-token-lists metadata parsing through pluggable parser interfaces.
- **Chain Filtering** – All parsers accept a `supportedChains` parameter to filter tokens by blockchain network, enabling applications to focus on relevant chains only.
- **Address Validation** – Comprehensive Ethereum address validation including checksummed, lowercase, and uppercase formats with proper error reporting for invalid addresses.
- **Extensible Architecture** – Parser interfaces (`TokenListParser`, `ListOfTokenListsParser`) allow easy addition of new formats without modifying existing code.
- **Error Resilience** – Graceful handling of malformed data, missing fields, and invalid JSON with detailed error messages for debugging.

### 2.14 Token Fetcher Design

The `pkg/tokens/fetcher` package provides production-ready HTTP fetching capabilities:

- **Concurrent Operations** – Uses goroutines for parallel fetching of multiple token lists with proper synchronization and error aggregation.
- **HTTP ETag Caching** – Implements efficient caching using HTTP ETags to minimize bandwidth usage, returning empty data for 304 Not Modified responses.
- **Configurable HTTP Client** – Customizable timeout, connection pooling, compression settings, and idle connection management for optimal performance.
- **JSON Schema Validation** – Optional validation against JSON schemas with support for both inline schemas and remote schema URLs.
- **Context Support** – Full context cancellation and timeout support for proper resource management and request lifecycle control.

### 2.15 Token AutoFetcher Design

The `pkg/tokens/autofetcher` package provides automated background token list management:

- **Two Operation Modes** – Supports direct token list fetching for known lists and remote list-of-token-lists discovery for dynamic list management.
- **Configurable Refresh Logic** – Flexible refresh intervals with separate check intervals, allowing fine-grained control over when refreshes occur.
- **Thread-Safe Operations** – All operations are safe for concurrent access with proper synchronization and atomic state updates.
- **Pluggable Storage** – ContentStore interface allows integration with various storage backends (memory, database, file system) for caching fetched content.
- **Error Reporting** – Real-time error notifications via channels, enabling applications to monitor refresh operations and handle failures appropriately.

### 2.16 Token Builder Design

The `pkg/tokens/builder` package implements the Builder pattern for incremental token collection construction:

- **Incremental Building** – Start with empty state and progressively add token lists, maintaining internal state throughout the building process.
- **Automatic Deduplication** – Prevents duplicate tokens using chain ID and address combinations, ensuring each unique token appears only once in the final collection.
- **Native Token Generation** – Automatically generates native tokens (ETH, BNB, etc.) for supported blockchain networks, ensuring comprehensive token coverage.
- **Multiple Format Support** – Integrates with parsers to handle various token list formats, providing a unified interface regardless of source format.
- **Stateful Construction** – Maintains both individual token lists and unified token collection, enabling applications to track origin and build complex token hierarchies.

### 2.17 Token Manager Design

The `pkg/tokens/manager` package provides a high-level interface for comprehensive token management:

- **Multi-Source Integration** – Combines tokens from native generation, remote token lists, local token lists, and custom user tokens into a unified collection.
- **Thread-Safe Concurrent Access** – Optimized for concurrent read operations using RWMutex, allowing multiple goroutines to access token data simultaneously.
- **Rich Query Capabilities** – Provides methods to find tokens by chain ID, address, list ID, or token keys, enabling efficient token discovery and lookup.
- **Automatic Refresh Management** – Integrates with AutoFetcher to provide background refresh capabilities with configurable intervals and error handling.
- **Deterministic Processing Order** – Processes token sources in a consistent order (native → main list → additional lists → custom tokens) ensuring predictable results across runs.
- **Error Resilience** – Graceful handling of network failures, parsing errors, and storage issues with fallback mechanisms to maintain core functionality.

### 2.18 ENS Resolver Design

The `pkg/ens` package provides Ethereum Name Service resolution capabilities:

- **Forward Resolution** – Converts ENS names (e.g., `vitalik.eth`) to Ethereum addresses using the `AddressOf()` method. Names are normalized to lowercase before resolution.
- **Reverse Resolution** – Converts Ethereum addresses to their primary ENS names using the `GetName()` method. Returns the name associated with an address's reverse record.
- **Chain Support Detection** – The `IsSupportedChain()` function checks if the given chain ID is one of the supported ENS chains: Ethereum Mainnet (1), Sepolia (11155111), or Holesky (17000).
- **Minimal Validation** – Performs basic structural validation on ENS names (must contain a dot, cannot start/end with a dot) while delegating full validation to the go-ens library for ENSIP-15 compliance including unicode support.
- **Thin Wrapper** – Designed as a lightweight wrapper around go-ens/v3, passing through errors directly without additional wrapping to maintain transparency.

### 2.19 Transaction Generator Design

The `pkg/txgenerator` package provides utilities for generating unsigned Ethereum transactions:

- **Multi-Token Standard Support** – Supports generating transactions for native ETH transfers, ERC20 operations (transfers and approvals), ERC721 operations (transfers via transferFrom and safeTransferFrom, approvals, and operator management), and ERC1155 operations (single and batch transfers, operator management).
- **Automatic Transaction Type Detection** – Automatically determines whether to create a legacy (type 0) or EIP-1559 (type 2) transaction based on the provided gas parameters. If `GasPrice` is provided, creates a legacy transaction. If `MaxFeePerGas` or `MaxPriorityFeePerGas` is provided, creates an EIP-1559 transaction.
- **Parameter Validation** – Comprehensive validation of all transaction parameters including address validation (non-zero addresses), amount validation (non-negative values), and chain ID requirements. Returns specific errors for common issues such as missing gas parameters or invalid addresses.
- **Contract ABI Integration** – Uses contract bindings from `pkg/contracts` to properly encode function calls for ERC20, ERC721, and ERC1155 operations, ensuring compatibility with standard token interfaces.
- **Unsigned Transactions** – All generated transactions are unsigned and ready for signing using keystore or extkeystore modules. Returns standard `types.Transaction` objects that can be serialized, signed, and broadcast to the network.
- **Chain Agnostic** – Works with any EVM-compatible chain by accepting chain ID as a parameter. Does not make assumptions about chain-specific transaction formats beyond legacy vs EIP-1559 distinction.

## 3. API Description

### 3.1 Multicall API (`pkg/multicall`)

The multicall package provides efficient batching of multiple Ethereum contract calls into single transactions using Multicall3.

#### 3.1.1 Call Builders

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `BuildNativeBalanceCall(accountAddress, multicall3Address)` | Builds a call to get native ETH balance | `accountAddress`: `common.Address`, `multicall3Address`: `common.Address` | `multicall3.IMulticall3Call` |
| `BuildERC20BalanceCall(accountAddress, tokenAddress)` | Builds a call to get ERC20 token balance | `accountAddress`: `common.Address`, `tokenAddress`: `common.Address` | `multicall3.IMulticall3Call` |
| `BuildERC721BalanceCall(accountAddress, tokenAddress)` | Builds a call to get ERC721 NFT balance | `accountAddress`: `common.Address`, `tokenAddress`: `common.Address` | `multicall3.IMulticall3Call` |
| `BuildERC1155BalanceCall(accountAddress, tokenAddress, tokenID)` | Builds a call to get ERC1155 token balance | `accountAddress`: `common.Address`, `tokenAddress`: `common.Address`, `tokenID`: `*big.Int` | `multicall3.IMulticall3Call` |

#### 3.1.2 Execution Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `RunSync(ctx, jobs, atBlock, caller, batchSize)` | Execute jobs synchronously | `ctx`: `context.Context`, `jobs`: `[]Job`, `atBlock`: `*big.Int`, `caller`: `Caller`, `batchSize`: `int` | `[]JobResult` |
| `RunAsync(ctx, jobs, atBlock, caller, batchSize)` | Execute jobs asynchronously | `ctx`: `context.Context`, `jobs`: `[]Job`, `atBlock`: `*big.Int`, `caller`: `Caller`, `batchSize`: `int` | `<-chan JobsResult` |

#### 3.1.3 Result Processing

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `ProcessNativeBalanceResult(result)` | Parse ETH balance from result | `result`: `multicall3.IMulticall3Result` | `*big.Int`, `error` |
| `ProcessERC20BalanceResult(result)` | Parse ERC20 balance from result | `result`: `multicall3.IMulticall3Result` | `*big.Int`, `error` |
| `ProcessERC721BalanceResult(result)` | Parse ERC721 balance from result | `result`: `multicall3.IMulticall3Result` | `*big.Int`, `error` |
| `ProcessERC1155BalanceResult(result)` | Parse ERC1155 balance from result | `result`: `multicall3.IMulticall3Result` | `*big.Int`, `error` |

#### 3.1.4 Types

```go
type Job struct {
    Calls        []multicall3.IMulticall3Call
    CallResultFn func(multicall3.IMulticall3Result) (any, error)
}

type CallResult struct {
    Value any
    Err   error
}

type JobResult struct {
    Results     []CallResult
    Err         error
    BlockNumber *big.Int
    BlockHash   common.Hash
}

type JobsResult struct {
    JobIdx    int
    JobResult JobResult
}

type Caller interface {
    ViewTryBlockAndAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) (*big.Int, [32]byte, []multicall3.IMulticall3Result, error)
    ViewTryAggregate(opts *bind.CallOpts, requireSuccess bool, calls []multicall3.IMulticall3Call) ([]multicall3.IMulticall3Result, error)
}
```

### 3.2 Balance Fetcher API (`pkg/balance/fetcher`)

The balance fetcher exposes two primary functions:

| Function                                                                            | Purpose                                                                                                                                                                                                | Parameters                                                                                                                                                                                                                                              | Returns                                                                                                                                                     |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FetchNativeBalances(ctx, addresses, atBlock, rpcClient, batchSize)`                | Retrieves native token balances (e.g., ETH) for multiple addresses.  The function first tries to use Multicall3 contract calls; if unavailable it sends batched `eth_getBalance` RPC calls.         | `ctx`: context; `addresses`: slice of addresses; `atBlock`: block number or `nil` for latest; `rpcClient`: implements `RPCClient`; `batchSize`: maximum addresses per batch.                                                                            | A map `map[common.Address]*big.Int` associating each address with its balance.  Errors indicate network issues or RPC failures.                             |
| `FetchErc20Balances(ctx, addresses, tokenAddresses, atBlock, rpcClient, batchSize)` | Retrieves ERC‑20 token balances for multiple addresses and tokens.  Uses Multicall3 contract calls when available or falls back to batched `eth_call` of `balanceOf` for each (address, token) pair. | `ctx`: context; `addresses`: slice of account addresses; `tokenAddresses`: slice of ERC‑20 contract addresses; `atBlock`: block number or `nil`; `rpcClient`: implements `RPCClient` and `BatchCaller`; `batchSize`: maximum number of calls per batch. | A nested map `map[address]map[token]*big.Int` where `balances[account][token]` is the token balance.  Errors indicate RPC failures or contract call errors. |

More specific functions are also available:

| Function                                                                                                          | Description                                                                                                                                                                                                                  |
| ----------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FetchNativeBalancesWithMulticall(ctx, addresses, atBlock, multicallCaller, multicallAddress, batchSize)`        | Batches addresses into groups (`batchSize`) and uses Multicall3 to retrieve native balances; decodes results into big.Int balances.                                                                                          |
| `FetchErc20BalancesWithMulticall(ctx, accountAddresses, tokenAddresses, atBlock, multicallCaller, batchSize)`    | Uses Multicall3 to batch ERC20 balance calls for multiple accounts and tokens.                                                                                                       |
| `FetchNativeBalancesStandard(ctx, addresses, atBlock, batchCaller, batchSize)`                                    | Constructs `eth_getBalance` batch requests using the provided `BatchCaller`; decodes hex strings into big.Int balances.                                                                                                      |
| `FetchErc20BalancesStandard(ctx, addresses, tokenAddresses, atBlock, batchCaller, batchSize)`                     | Builds `eth_call` requests for each account/token pair using the ERC‑20 ABI and sends them in batches.                                                                                                                       |

**Multicall3 Deployments**

Multicall3 is deployed on 200+ chains and is the primary method for batching contract calls. The BalanceFetcher uses the following addresses:

| Chain | ChainID | Address |
| ----- | ------- | ------- |
| Ethereum Mainnet | 1 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Optimism Mainnet | 10 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Arbitrum Mainnet | 42161 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Base Mainnet | 8453 | 0xca11bde05977b3631167028862be2a173976ca11 |
| BSC Mainnet | 56 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Polygon Mainnet | 137 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Ethereum Sepolia | 11155111 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Arbitrum Sepolia | 421614 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Optimism Sepolia | 11155420 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Base Sepolia | 84532 | 0xca11bde05977b3631167028862be2a173976ca11 |
| Status Network Sepolia | 1660990954 | 0xca11bde05977b3631167028862be2a173976ca11 |
| BSC Testnet | 97 | 0xca11bde05977b3631167028862be2a173976ca11 |

These are defined at `pkg/contracts/multicall3/deployments.go` and can be accessed via `multicall3.GetMulticall3Address(chainID)`.

**ERC-20 ABI Usage**

- Uses generated bindings in `pkg/contracts/erc20` or packs `balanceOf(address)` via the ABI for `eth_call` in standard mode.
- In standard mode, `balanceOf` is encoded with `abi.Pack("balanceOf", accountAddress)` and sent as `input` to the token contract `to` address.

**Regenerating Bindings**

The SDK uses auto-generated Go bindings to interact with smart contracts. These bindings provide type-safe method calls and handle ABI encoding/decoding automatically.

Use `abigen` to regenerate bindings when contract sources are updated:

```bash
# ERC-20 from Solidity interface
abigen --sol pkg/contracts/erc20/IERC20.sol --pkg erc20 --out pkg/contracts/erc20/erc20.go

# ERC-721 from Solidity interface
abigen --sol pkg/contracts/erc721/IERC721.sol --pkg erc721 --out pkg/contracts/erc721/erc721.go

# ERC-1155 from Solidity interface
abigen --sol pkg/contracts/erc1155/IERC1155.sol --pkg erc1155 --out pkg/contracts/erc1155/erc1155.go

# Alternative: Generate from ABI JSON (if available)
abigen --abi IERC20.abi.json --pkg erc20 --out pkg/contracts/erc20/erc20.go
```

Ensure the ABI/Solidity sources match the deployed contract versions. Regeneration is needed when contracts are upgraded, ABIs change, or new functionality is added.

### 3.2 Ethereum Client API (`pkg/ethclient`)

This package exports a `Client` type that wraps a lower‑level RPC client and provides both go‑ethereum‑compatible methods and chain‑agnostic methods. Developers construct a client using `NewClient(rpcClient)`. If the provided RPC client is a go‑ethereum `rpc.Client`, the SDK internally also creates a `gethEthClient` for compatibility with existing `ethclient` code

The Ethereum client exposes a large set of methods. They can be grouped into several categories. All methods follow the same pattern of accepting a context and returning typed data or errors.

| Method                   | Description                                          | Example                                                                                |
| ------------------------ | ---------------------------------------------------- | -------------------------------------------------------------------------------------- |
| `Web3ClientVersion(ctx)` | Returns the version of the Ethereum client software | `client.Web3ClientVersion(ctx)` returns a string like `"Geth/v1.16.0-stable/linux"` |

**Net Namespace**

| Method               | Description                                                 | Example                                                                          |
| -------------------- | ----------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `NetListening(ctx)`  | Returns whether the client is actively listening for peers | `client.NetListening(ctx)` returns `true` if listening                          |
| `NetPeerCount(ctx)`  | Returns the number of connected peers                       | `client.NetPeerCount(ctx)` returns `uint64` peer count                          |
| `NetVersion(ctx)`    | Returns the network ID as a string                          | `client.NetVersion(ctx)` returns `"1"` for Mainnet, `"11155111"` for Sepolia   |

**Eth Namespace - Node/Network Information**

| Method                           | Description                                                  | Example                                                                    |
| -------------------------------- | ------------------------------------------------------------ | -------------------------------------------------------------------------- |
| `EthProtocolVersion(ctx)`        | Returns the Ethereum protocol version                        | `client.EthProtocolVersion(ctx)` returns `"0x41"` (protocol version 65)   |
| `EthChainId(ctx)`                | Returns the chain ID as a big integer                       | `client.EthChainId(ctx)` returns `*big.Int` with value `1` for Mainnet    |
| `EthSyncing(ctx)`                | Returns sync status or false if not syncing                 | `client.EthSyncing(ctx)` returns `*ethereum.SyncProgress` or `false`      |
| `EthCoinbase(ctx)`               | Returns the coinbase address (mining reward recipient)      | `client.EthCoinbase(ctx)` returns `common.Address`                        |
| `EthMining(ctx)`                 | Returns whether the client is mining                        | `client.EthMining(ctx)` returns `true` if mining                          |
| `EthHashrate(ctx)`               | Returns the mining hashrate in hashes per second            | `client.EthHashrate(ctx)` returns `uint64` hashrate                       |
| `EthMaxPriorityFeePerGas(ctx)`   | Returns suggested priority fee for EIP‑1559 transactions    | `client.EthMaxPriorityFeePerGas(ctx)` returns `*big.Int` in wei           |
| `EthBlobBaseFee(ctx)`            | Returns the base fee for blob transactions (EIP‑4844)       | `client.EthBlobBaseFee(ctx)` returns `*big.Int` base fee in wei           |

**Eth Namespace - Blocks**

| Method                                          | Description                                                          | Example                                                                        |
| ----------------------------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `EthBlockNumber(ctx)`                           | Returns the number of the most recent block                         | `client.EthBlockNumber(ctx)` returns `uint64` block number                    |
| `EthGetBlockByHashWithTxHashes(ctx, hash)`      | Fetches a block by hash with transaction hashes only                | `client.EthGetBlockByHashWithTxHashes(ctx, blockHash)`                        |
| `EthGetBlockByNumberWithTxHashes(ctx, number)`  | Fetches a block by number with transaction hashes only              | `client.EthGetBlockByNumberWithTxHashes(ctx, big.NewInt(19543210))`           |
| `EthGetBlockByHashWithFullTxs(ctx, hash)`       | Fetches a block by hash with full transaction objects               | `client.EthGetBlockByHashWithFullTxs(ctx, blockHash)`                         |
| `EthGetBlockByNumberWithFullTxs(ctx, number)`   | Fetches a block by number with full transaction objects             | `client.EthGetBlockByNumberWithFullTxs(ctx, big.NewInt(19543210))`            |
| `EthGetBlockReceipts(ctx, number)`              | Returns all transaction receipts for a given block                  | `client.EthGetBlockReceipts(ctx, big.NewInt(19543210))`                       |
| `EthGetBlockTransactionCountByHash(ctx, hash)`  | Returns the number of transactions in a block by hash               | `client.EthGetBlockTransactionCountByHash(ctx, blockHash)`                    |
| `EthGetBlockTransactionCountByNumber(ctx, num)` | Returns the number of transactions in a block by number             | `client.EthGetBlockTransactionCountByNumber(ctx, big.NewInt(19543210))`       |
| `EthGetUncleByBlockHashAndIndex(ctx, hash, i)`  | Returns uncle block by block hash and uncle index                   | `client.EthGetUncleByBlockHashAndIndex(ctx, blockHash, 0)`                    |
| `EthGetUncleByBlockNumberAndIndex(ctx, num, i)` | Returns uncle block by block number and uncle index                 | `client.EthGetUncleByBlockNumberAndIndex(ctx, big.NewInt(19543210), 0)`       |
| `EthGetUncleCountByBlockHash(ctx, hash)`        | Returns the number of uncles in a block by hash                     | `client.EthGetUncleCountByBlockHash(ctx, blockHash)`                          |
| `EthGetUncleCountByBlockNumber(ctx, number)`    | Returns the number of uncles in a block by number                   | `client.EthGetUncleCountByBlockNumber(ctx, big.NewInt(19543210))`             |

**Eth Namespace - Transactions**

| Method                                                | Description                                                   | Example                                                                    |
| ----------------------------------------------------- | ------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `EthSendRawTransaction(ctx, rawTx)`                   | Submits a signed transaction to the network                  | `client.EthSendRawTransaction(ctx, signedTxBytes)`                         |
| `EthSendTransaction(ctx, tx)`                         | Submits a transaction using a managed account                | `client.EthSendTransaction(ctx, txObject)` (requires unlocked account)    |
| `EthGetTransactionByHash(ctx, hash)`                  | Returns transaction details by transaction hash              | `client.EthGetTransactionByHash(ctx, txHash)`                             |
| `EthGetTransactionByBlockHashAndIndex(ctx, hash, i)`  | Returns transaction by block hash and transaction index      | `client.EthGetTransactionByBlockHashAndIndex(ctx, blockHash, 0)`          |
| `EthGetTransactionByBlockNumberAndIndex(ctx, num, i)` | Returns transaction by block number and transaction index    | `client.EthGetTransactionByBlockNumberAndIndex(ctx, big.NewInt(123), 0)`  |
| `EthGetTransactionReceipt(ctx, hash)`                 | Returns the receipt of a transaction by hash                 | `client.EthGetTransactionReceipt(ctx, txHash)`                            |
| `EthGetTransactionCount(ctx, address, atBlock)`       | Returns the nonce (transaction count) for an account        | `client.EthGetTransactionCount(ctx, myAddress, nil)`                      |
| `EthSign(ctx, addr, data)`                            | Signs arbitrary data with an account's private key          | `client.EthSign(ctx, myAddress, dataToSign)`                              |
| `EthSignTransaction(ctx, tx)`                         | Signs a transaction without sending it                       | `client.EthSignTransaction(ctx, txObject)`                                |

**Eth Namespace - Account/State**

| Method                                        | Description                                                     | Example                                                           |
| --------------------------------------------- | --------------------------------------------------------------- | ----------------------------------------------------------------- |
| `EthGetBalance(ctx, address, atBlock)`        | Returns the balance of an account at a given block             | `client.EthGetBalance(ctx, myAddress, nil)`                      |
| `EthGetCode(ctx, address, atBlock)`           | Returns the contract code at an address                        | `client.EthGetCode(ctx, contractAddr, nil)`                      |
| `EthGetStorageAt(ctx, address, key, atBlock)` | Returns the value from a storage position at an address        | `client.EthGetStorageAt(ctx, contractAddr, storageKey, nil)`     |
| `EthGetProof(ctx, address, keys, atBlock)`    | Returns account and storage proofs for Merkle verification     | `client.EthGetProof(ctx, myAddress, []string{storageKey}, nil)`  |

**Eth Namespace - Gas**

| Method                                                   | Description                                                    | Example                                                           |
| -------------------------------------------------------- | -------------------------------------------------------------- | ----------------------------------------------------------------- |
| `EthGasPrice(ctx)`                                       | Returns the current gas price in wei                          | `client.EthGasPrice(ctx)` returns `*big.Int`                     |
| `EthEstimateGas(ctx, callMsg)`                           | Estimates the gas required to execute a transaction           | `client.EthEstimateGas(ctx, callMsg)` returns `uint64`           |
| `EthFeeHistory(ctx, count, lastBlock, rewardPercentiles)` | Returns historical base fee and priority fee data             | `client.EthFeeHistory(ctx, 10, nil, []float64{25, 50, 75})`      |

**Eth Namespace - Call/Logs/Filters**

| Method                              | Description                                                   | Example                                                      |
| ----------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------ |
| `EthCall(ctx, callMsg, atBlock)`    | Executes a read‑only contract call without creating a tx     | `client.EthCall(ctx, callMsg, nil)`                         |
| `EthGetLogs(ctx, filterQuery)`      | Returns event logs matching a filter query                   | `client.EthGetLogs(ctx, filterQuery)`                       |
| `EthNewFilter(ctx, filterQuery)`    | Creates a new log filter and returns its ID                  | `client.EthNewFilter(ctx, filterQuery)` returns filter ID   |
| `EthNewBlockFilter(ctx)`            | Creates a new block filter and returns its ID                | `client.EthNewBlockFilter(ctx)` returns filter ID           |
| `EthGetFilterLogs(ctx, filterID)`   | Returns all logs for a filter (only for log filters)        | `client.EthGetFilterLogs(ctx, filterID)`                    |
| `EthGetFilterChanges(ctx, filterID)`| Returns new entries since last poll for any filter type      | `client.EthGetFilterChanges(ctx, filterID)`                 |
| `EthUninstallFilter(ctx, filterID)` | Uninstalls a filter and stops polling                        | `client.EthUninstallFilter(ctx, filterID)` returns `bool`   |

The chain‑agnostic methods (prefixed with `Eth*`, `Net*`, `Web3*`) correspond directly to Ethereum JSON‑RPC calls and accept/return standard Go types, making them compatible with any EVM‑compatible chain. For backward compatibility, the package also exports go‑ethereum compatible methods such as `BlockNumber(ctx)`, `BalanceAt(ctx, address, nil)`, etc., which call the same RPC methods but use go‑ethereum types.

**RPC Parameter Translation Helpers**

The Ethereum client includes several critical helper functions that bridge the gap between Go types and the specific JSON-RPC parameter formats required by Ethereum nodes. These helpers are essential because:

1. **Ethereum JSON-RPC has strict formatting requirements** - Parameters must be properly encoded as hex strings, structured objects, or special sentinel values
2. **Go types don't directly match RPC expectations** - Standard Go types like `*big.Int`, `ethereum.CallMsg`, and `ethereum.FilterQuery` need transformation
3. **Chain compatibility requires consistent encoding** - Different Ethereum clients expect the same standardized parameter formats

```go
// Block number encoder handling negative sentinel values for latest/finalized/etc.
func toBlockNumArg(number *big.Int) string

// Call and filter translators to RPC args
func toCallArg(msg ethereum.CallMsg) interface{}
func toFilterArg(q ethereum.FilterQuery) (interface{}, error)
```

**Block Number Translation (`toBlockNumArg`)**

Converts Go `*big.Int` block numbers into proper JSON-RPC format:
- `nil` → `"latest"` (most recent block)
- Positive numbers → hex-encoded strings (e.g., `big.NewInt(12345)` → `"0x3039"`)
- Special negative values → sentinel strings:
  - `-1` → `"pending"` (pending block)
  - `-2` → `"latest"` (latest mined block)
  - `-3` → `"finalized"` (finalized block)
  - `-4` → `"safe"` (safe block)
  - `-5` → `"earliest"` (genesis block)

This is used by all block-parameter methods like `EthGetBalance`, `EthGetCode`, `EthCall`, etc.

**Call Message Translation (`toCallArg`)**

Converts Go `ethereum.CallMsg` structs into JSON-RPC call objects with proper hex encoding:
- Addresses → hex strings
- Data/input → hex-encoded bytes
- Gas values → hex-encoded numbers
- Wei amounts → hex-encoded big integers
- EIP-1559 fee fields → properly formatted fee caps
- Access lists and blob parameters → structured objects

This ensures `EthCall`, `EthEstimateGas`, and transaction methods send correctly formatted parameters.

**Filter Query Translation (`toFilterArg`)**

Converts Go `ethereum.FilterQuery` structs into JSON-RPC filter objects:
- Address lists → arrays of hex-encoded addresses
- Topics → arrays of topic hashes with proper null handling
- Block ranges → properly formatted block parameters using `toBlockNumArg`
- Validates mutually exclusive parameters (block hash vs. block range)

This enables `EthGetLogs`, `EthNewFilter`, and other event filtering methods to work correctly across all EVM chains.

### 3.3 Gas Estimation API (`pkg/gas`)

The gas package provides comprehensive gas fee estimation and transaction inclusion time predictions for Ethereum and L2 networks.

#### 3.3.1 Core Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `GetChainSuggestions(ctx, ethClient, params, config, account)` | Get fee suggestions for a specific account without transaction details | `ctx`: `context.Context`, `ethClient`: `EthClient`, `params`: `ChainParameters`, `config`: `SuggestionsConfig`, `account`: `common.Address` | `*FeeSuggestions`, `error` |
| `GetTxSuggestions(ctx, ethClient, params, config, callMsg)` | Get fee suggestions and gas limit for a transaction | `ctx`: `context.Context`, `ethClient`: `EthClient`, `params`: `ChainParameters`, `config`: `SuggestionsConfig`, `callMsg`: `*ethereum.CallMsg` | `*TxSuggestions`, `error` |
| `EstimateInclusion(ctx, ethClient, params, config, fee)` | Estimate inclusion time for a specific fee | `ctx`: `context.Context`, `ethClient`: `EthClient`, `params`: `ChainParameters`, `config`: `SuggestionsConfig`, `fee`: `Fee` | `*Inclusion`, `error` |
| `DefaultConfig(chainClass)` | Get default configuration for a chain class | `chainClass`: `ChainClass` | `SuggestionsConfig` |

#### 3.3.2 Chain Parameters

```go
type ChainParameters struct {
    ChainClass       ChainClass  // L1, ArbStack, OPStack, or LineaStack
    NetworkBlockTime float64     // Average block time in seconds
}
```

| ChainClass | Description | Example Networks | Block Time |
|------------|-------------|------------------|------------|
| `ChainClassL1` | Ethereum L1 chains | Ethereum, Polygon, BSC | 12s (Ethereum), 2.25s (Polygon), 0.75s (BSC) |
| `ChainClassArbStack` | Arbitrum-based chains | Arbitrum One, Arbitrum Nova | 0.25s |
| `ChainClassOPStack` | Optimism-based chains | Optimism, Base, OP Sepolia | 2s |
| `ChainClassLineaStack` | Linea-based chains | Linea Mainnet, Status Network | 2s |

#### 3.3.3 Configuration

```go
type SuggestionsConfig struct {
    NetworkCongestionBlocks           int     // Blocks to analyze for congestion (default: 10)
    GasPriceEstimationBlocks          int     // Blocks for gas price estimation (default: 10 for L1, 50 for L2)
    LowRewardPercentile               float64 // Percentile for low priority (default: 10)
    MediumRewardPercentile            float64 // Percentile for medium priority (default: 45)
    HighRewardPercentile              float64 // Percentile for high priority (default: 90)
    LowBaseFeeMultiplier              float64 // Base fee multiplier for low level (default: 1.025)
    MediumBaseFeeMultiplier           float64 // Base fee multiplier for medium level (default: 1.025 for L1, 4.1 for L2)
    HighBaseFeeMultiplier             float64 // Base fee multiplier for high level (default: 1.025 for L1, 10.25 for L2)
    LowBaseFeeCongestionMultiplier    float64 // Congestion adjustment for low (L1 only, default: 0.0)
    MediumBaseFeeCongestionMultiplier float64 // Congestion adjustment for medium (L1 only, default: 10.0)
    HighBaseFeeCongestionMultiplier   float64 // Congestion adjustment for high (L1 only, default: 10.0)
}
```

#### 3.3.4 Response Types

```go
type TxSuggestions struct {
    GasLimit       *big.Int        // Estimated gas limit for the transaction
    FeeSuggestions *FeeSuggestions // Fee suggestions for three priority levels
}

type FeeSuggestions struct {
    // Fee suggestions for three priority levels
    Low    Fee // Low priority fee
    Medium Fee // Medium priority fee
    High   Fee // High priority fee

    // Inclusion time estimates
    LowInclusion    Inclusion // Low priority inclusion estimate
    MediumInclusion Inclusion // Medium priority inclusion estimate
    HighInclusion   Inclusion // High priority inclusion estimate

    // Network state
    EstimatedBaseFee      *big.Int // Next block's estimated base fee (wei)
    PriorityFeeLowerBound *big.Int // Recommended min priority fee (wei)
    PriorityFeeUpperBound *big.Int // Recommended max priority fee (wei)
    NetworkCongestion     float64  // Congestion score 0-1 (L1 only)
}

type Fee struct {
    MaxPriorityFeePerGas *big.Int // Max priority fee per gas (wei)
    MaxFeePerGas         *big.Int // Max fee per gas (wei)
}

type Inclusion struct {
    MinBlocksUntilInclusion int     // Minimum blocks until inclusion
    MaxBlocksUntilInclusion int     // Maximum blocks until inclusion
    MinTimeUntilInclusion   float64 // Minimum time in seconds
    MaxTimeUntilInclusion   float64 // Maximum time in seconds
}
```

#### 3.3.5 EthClient Interface

```go
type EthClient interface {
    FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int,
               rewardPercentiles []float64) (*ethereum.FeeHistory, error)
    EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
    LineaEstimateGas(ctx context.Context, msg ethereum.CallMsg) (*LineaEstimateGasResult, error)
}
```

#### 3.3.6 Usage Example

```go
import "github.com/status-im/go-wallet-sdk/pkg/gas"

// Define chain parameters
params := gas.ChainParameters{
    ChainClass:       gas.ChainClassL1,
    NetworkBlockTime: 12.0, // Ethereum block time
}

// Get default config or customize
config := gas.DefaultConfig(params.ChainClass)

// Create transaction call message
callMsg := &ethereum.CallMsg{
    From:  common.HexToAddress("0x..."),
    To:    &toAddress,
    Value: big.NewInt(0),
    Data:  txData,
}

// Get fee suggestions
suggestions, err := gas.GetTxSuggestions(ctx, ethClient, params, config, callMsg)
if err != nil {
    return err
}

// Use medium priority fees
maxPriorityFee := suggestions.FeeSuggestions.Medium.MaxPriorityFeePerGas
maxFee := suggestions.FeeSuggestions.Medium.MaxFeePerGas
gasLimit := suggestions.GasLimit

// Check estimated wait time
minWait := suggestions.FeeSuggestions.MediumInclusion.MinTimeUntilInclusion
maxWait := suggestions.FeeSuggestions.MediumInclusion.MaxTimeUntilInclusion

// Estimate inclusion for custom fee
customFee := gas.Fee{
    MaxPriorityFeePerGas: big.NewInt(2000000000), // 2 gwei
    MaxFeePerGas:         big.NewInt(30000000000), // 30 gwei
}
inclusion, err := gas.EstimateInclusion(ctx, ethClient, params, config, customFee)
```

#### 3.3.7 GetChainSuggestions Method

The `GetChainSuggestions` method provides fee suggestions for a specific account without requiring transaction details. This is useful for getting general fee recommendations based on network conditions and account-specific factors.

**Method Signature:**
```go
func GetChainSuggestions(
    ctx context.Context,
    ethClient EthClient,
    params ChainParameters,
    config SuggestionsConfig,
    account common.Address,
) (*FeeSuggestions, error)
```

**Key Features:**
- **Account-Specific**: Uses the account address for chain-specific optimizations (especially for LineaStack)
- **No Transaction Required**: Provides fee suggestions without needing a specific transaction call message
- **Network-Aware**: Analyzes current network conditions and historical fee data
- **Chain-Optimized**: Uses different strategies based on chain class (L1, ArbStack, OPStack, LineaStack)

**Usage Example:**
```go
// Get general fee suggestions for an account
account := common.HexToAddress("0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6")
suggestions, err := gas.GetChainSuggestions(ctx, ethClient, params, config, account)
if err != nil {
    return err
}

// Access fee suggestions
lowFee := suggestions.Low.MaxFeePerGas
mediumFee := suggestions.Medium.MaxFeePerGas
highFee := suggestions.High.MaxFeePerGas

// Check inclusion time estimates
minWait := suggestions.MediumInclusion.MinTimeUntilInclusion
maxWait := suggestions.MediumInclusion.MaxTimeUntilInclusion

// Access network state
baseFee := suggestions.EstimatedBaseFee
congestion := suggestions.NetworkCongestion // L1 only
```

**Chain-Specific Behavior:**
- **L1 Chains**: Uses historical fee data and congestion analysis
- **ArbStack/OPStack**: Uses fixed multipliers with historical percentiles
- **LineaStack**: Uses `linea_estimateGas` RPC method with account-specific gas price suggestions

#### 3.3.8 Chain-Specific Behavior

| Chain Class | Base Fee Strategy | Priority Fee Source | Congestion Analysis |
|-------------|-------------------|---------------------|---------------------|
| L1 | Dynamic (1.025x base for all levels + congestion multiplier for medium/high) | Historical percentiles (10/45/90) | Yes (0-1 scale, 0x/10x/10x factors) |
| ArbStack | Fixed (1.025x, 4.1x, 10.25x multipliers) | Historical percentiles (10/45/90) | No (0x/0x/0x factors) |
| OPStack | Fixed (1.025x, 4.1x, 10.25x multipliers) | Historical percentiles (10/45/90) | No (0x/0x/0x factors) |
| LineaStack | 2x base fee for all levels | `linea_estimateGas` RPC | No |

### 3.4 Event Filter API (`pkg/eventfilter`)

The event filter package provides efficient filtering for Ethereum transfer events across ERC20, ERC721, and ERC1155 tokens with concurrent processing capabilities.

#### 3.4.1 Configuration

| Type | Description | Values |
|------|-------------|---------|
| `TransferType` | Token type to filter | `TransferTypeERC20`, `TransferTypeERC721`, `TransferTypeERC1155` |
| `Direction` | Transfer direction | `Send`, `Receive`, `Both` |
| `TransferQueryConfig` | Main configuration struct | See below |

#### 3.4.2 TransferQueryConfig

```go
type TransferQueryConfig struct {
    FromBlock         *big.Int           // Start block number
    ToBlock           *big.Int           // End block number
    ContractAddresses []common.Address   // Optional contract addresses to filter
    Accounts          []common.Address   // Addresses to filter for
    TransferTypes     []TransferType     // Token types to include
    Direction         Direction          // Transfer direction filter
}
```

#### 3.4.3 Core Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `FilterTransfers(ctx, client, config)` | Filter and parse transfer events with concurrent processing | `ctx`: `context.Context`, `client`: `FilterClient`, `config`: `TransferQueryConfig` | `[]eventlog.Event`, `error` |
| `config.ToFilterQueries()` | Generate optimized filter queries | `config`: `TransferQueryConfig` | `[]ethereum.FilterQuery` |

#### 3.4.4 FilterClient Interface

```go
type FilterClient interface {
    FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
}
```

#### 3.4.5 Concurrent Processing

The `FilterTransfers` function provides concurrent processing of filter queries:

- **Parallel Execution**: Multiple filter queries are executed concurrently using goroutines
- **Error Aggregation**: All query errors are collected and returned as a single joined error
- **Event Collection**: Results from all queries are merged into a single event slice
- **Resource Management**: Proper cleanup of goroutines and channels

#### 3.4.6 Query Optimization

The package minimizes API calls through intelligent query merging:

- **Single Transfer Types**: 1-2 queries (Send + Receive)
- **Mixed Transfer Types**: 2-3 queries maximum
- **Event Signature Merging**: Multiple event types in single query using OR operations
- **Topic Structure Optimization**: Merges compatible queries by omitting empty trailing topics

### 3.5 Event Log Parser API (`pkg/eventlog`)

The event log parser package provides automatic detection and parsing of Ethereum event logs.

#### 3.5.1 Core Types

```go
type Event struct {
    ContractKey ContractKey  // "erc20", "erc721", or "erc1155"
    ContractABI *abi.ABI     // Full contract ABI
    EventKey    EventKey     // Specific event type
    ABIEvent    *abi.Event   // ABI event definition
    Unpacked    any          // Type-safe parsed event data
}

type ContractKey string
type EventKey string
```

#### 3.5.2 Core Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `ParseLog(log)` | Parse a single log into events | `log`: `types.Log` | `[]Event` |

#### 3.5.3 Supported Events

| Contract | Event Types | Unpacked Type |
|----------|-------------|---------------|
| ERC20 | Transfer, Approval | `erc20.Erc20Transfer`, `erc20.Erc20Approval` |
| ERC721 | Transfer, Approval, ApprovalForAll | `erc721.Erc721Transfer`, `erc721.Erc721Approval`, `erc721.Erc721ApprovalForAll` |
| ERC1155 | TransferSingle, TransferBatch, ApprovalForAll, URI | `erc1155.Erc1155TransferSingle`, `erc1155.Erc1155TransferBatch`, `erc1155.Erc1155ApprovalForAll`, `erc1155.Erc1155URI` |

#### 3.5.4 Event Signatures

Uses standardized signatures from the `eventlog` package:
- **ERC20/ERC721 Transfer**: `0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef`
- **ERC1155 TransferSingle**: `0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62`
- **ERC1155 TransferBatch**: `0x4a39dc06d4c0dbc64b70af90fd698a233a518aa5d07e595d983b8c0526c8f7fb`

### 3.6 Extended Keystore API (`pkg/accounts/extkeystore`)

The extended keystore package provides HD wallet functionality with BIP32 extended key storage.

#### 3.6.1 Core Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `NewKeyStore(keydir, scryptN, scryptP)` | Create a new keystore instance | `keydir`: `string`, `scryptN`: `int`, `scryptP`: `int` | `*KeyStore` |
| `ImportExtendedKey(extKey, passphrase)` | Import a BIP32 extended key | `extKey`: `*extkeys.ExtendedKey`, `passphrase`: `string` | `accounts.Account`, `error` |
| `DeriveWithPassphrase(account, path, pin, passphrase, newPassphrase)` | Derive a child account from parent | `account`: `accounts.Account`, `path`: `accounts.DerivationPath`, `pin`: `bool`, `passphrase`: `string`, `newPassphrase`: `string` | `accounts.Account`, `error` |
| `Import(keyJSON, passphrase, newPassphrase)` | Import an encrypted key JSON | `keyJSON`: `[]byte`, `passphrase`: `string`, `newPassphrase`: `string` | `accounts.Account`, `error` |
| `ExportExt(account, passphrase, newPassphrase)` | Export extended key as JSON | `account`: `accounts.Account`, `passphrase`: `string`, `newPassphrase`: `string` | `[]byte`, `error` |
| `ExportPriv(account, passphrase, newPassphrase)` | Export as standard private key JSON | `account`: `accounts.Account`, `passphrase`: `string`, `newPassphrase`: `string` | `[]byte`, `error` |
| `SignHash(account, hash)` | Sign a hash with unlocked account | `account`: `accounts.Account`, `hash`: `[]byte` | `[]byte`, `error` |
| `SignHashWithPassphrase(account, passphrase, hash)` | Sign a hash with passphrase | `account`: `accounts.Account`, `passphrase`: `string`, `hash`: `[]byte` | `[]byte`, `error` |
| `TimedUnlock(account, passphrase, timeout)` | Unlock account with timeout | `account`: `accounts.Account`, `passphrase`: `string`, `timeout`: `time.Duration` | `error` |
| `Lock(address)` | Lock an account | `address`: `common.Address` | `error` |
| `Delete(account, passphrase)` | Delete an account | `account`: `accounts.Account`, `passphrase`: `string` | `error` |

#### 3.6.2 Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `LightScryptN` | `1 << 12` | Fast scrypt N parameter for development |
| `LightScryptP` | `6` | Fast scrypt P parameter for development |
| `StandardScryptN` | `1 << 18` | Standard scrypt N parameter for production |
| `StandardScryptP` | `1` | Standard scrypt P parameter for production |
| `KeyStoreScheme` | `"extkeystore"` | URL scheme for extended keystore accounts |

### 3.7 Mnemonic API (`pkg/accounts/mnemonic`)

The mnemonic package provides utilities for BIP39 mnemonic phrases.

#### 3.7.1 Core Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `CreateRandomMnemonic(length)` | Generate random mnemonic phrase | `length`: `int` (12, 15, 18, 21, or 24) | `string`, `error` |
| `CreateRandomMnemonicWithDefaultLength()` | Generate 12-word mnemonic | None | `string`, `error` |
| `CreateExtendedKeyFromMnemonic(phrase, passphrase)` | Create BIP32 extended key from mnemonic | `phrase`: `string`, `passphrase`: `string` | `*extkeys.ExtendedKey`, `error` |
| `LengthToEntropyStrength(length)` | Convert word count to entropy strength | `length`: `int` | `extkeys.EntropyStrength`, `error` |

#### 3.7.2 Usage Example

```go
import (
    "github.com/status-im/go-wallet-sdk/pkg/accounts/mnemonic"
    "github.com/status-im/go-wallet-sdk/pkg/accounts/extkeystore"
)

// Generate mnemonic and import into keystore
phrase, _ := mnemonic.CreateRandomMnemonic(12)
extKey, _ := mnemonic.CreateExtendedKeyFromMnemonic(phrase, "")
keystore := extkeystore.NewKeyStore("/path/to/keystore",
    extkeystore.LightScryptN, extkeystore.LightScryptP)
account, _ := keystore.ImportExtendedKey(extKey, "passphrase")
```

### 3.8 Token Types API (`pkg/tokens/types`)

The token types package provides core data structures for the token management system.

#### 3.8.1 Core Types

```go
type Token struct {
    CrossChainID string             `json:"crossChainId"` // Cross-chain identifier
    ChainID      uint64             `json:"chainId"`      // Blockchain network ID
    Address      gethcommon.Address `json:"address"`      // Contract address
    Decimals     uint               `json:"decimals"`     // Number of decimal places
    Name         string             `json:"name"`         // Full token name
    Symbol       string             `json:"symbol"`       // Token symbol/ticker
    LogoURI      string             `json:"logoUri"`      // URL to token logo
    CustomToken  bool               `json:"custom"`       // Whether this is a custom user token
}

type TokenList struct {
    ID               string                 `json:"id"`               // Token list ID
    Name             string                 `json:"name"`             // Display name
    Timestamp        time.Time              `json:"timestamp"`        // Last update time
    Version          *types.Version         `json:"version"`          // Version information
    SourceURL        string                 `json:"sourceUrl"`        // Source URL
    Schema           string                 `json:"schema"`            // JSON schema URL
    Tokens           []*Token               `json:"tokens"`            // List of tokens
}
```

#### 3.8.2 Utility Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `TokenKey(chainID, addr)` | Creates unique token key | `chainID`: `uint64`, `addr`: `common.Address` | `string` |
| `ChainAndAddressFromTokenKey(key)` | Extracts chain ID and address from key | `key`: `string` | `uint64`, `common.Address`, `bool` |
| `token.IsNative()` | Checks if token is native | `token`: `*Token` | `bool` |

### 3.9 Token Parsers API (`pkg/tokens/parsers`)

The token parsers package provides parsing for multiple token list formats.

#### 3.9.1 Parser Interfaces

```go
type TokenListParser interface {
    Parse(raw []byte, supportedChains []uint64) (*types.TokenList, error)
}

type ListOfTokenListsParser interface {
    Parse(raw []byte) (*types.ListOfTokenLists, error)
}
```

#### 3.9.2 Available Parsers

| Parser | Format | Use Case |
|--------|--------|----------|
| `StandardTokenListParser` | Uniswap-style | General purpose, most common |
| `StatusTokenListParser` | Status-specific with chain grouping | Multi-chain optimization |
| `CoinGeckoAllTokensParser` | CoinGecko API with platform mappings | Cross-platform discovery |
| `StatusListOfTokenListsParser` | List-of-token-lists metadata | Managing multiple token list sources |

#### 3.9.3 Usage Example

```go
parser := &parsers.StandardTokenListParser{}
tokenList, err := parser.Parse(jsonData, supportedChains)
if err != nil {
    return err
}
```

### 3.10 Token Fetcher API (`pkg/tokens/fetcher`)

The token fetcher package provides HTTP-based token list fetching.

#### 3.10.1 Configuration

```go
type Config struct {
    Timeout            time.Duration // Request timeout (default: 5s)
    IdleConnTimeout    time.Duration // Connection idle timeout (default: 90s)
    MaxIdleConns       int           // Max idle connections (default: 10)
    DisableCompression bool          // Disable gzip compression (default: false)
}
```

#### 3.10.2 Core Interface

```go
type Fetcher interface {
    Fetch(ctx context.Context, details FetchDetails) (FetchedData, error)
    FetchConcurrent(ctx context.Context, details []FetchDetails) ([]FetchedData, error)
}
```

#### 3.10.3 Data Types

```go
type FetchDetails struct {
    types.ListDetails  // Embedded: ID, SourceURL, Schema
    Etag string             // HTTP ETag for caching
}

type FetchedData struct {
    FetchDetails           // Original fetch details
    Fetched  time.Time     // Timestamp when fetched
    JsonData []byte        // Raw JSON data (nil if 304 Not Modified)
    Error    error         // Error that occurred during fetch
}
```

### 3.11 Token AutoFetcher API (`pkg/tokens/autofetcher`)

The token autofetcher package provides automated background token list management.

#### 3.11.1 Core Interface

```go
type AutoFetcher interface {
    Start(ctx context.Context) (refreshCh chan error)
    Stop()
}
```

#### 3.11.2 Storage Interface

```go
type ContentStore interface {
    GetEtag(id string) (string, error)
    Get(id string) (Content, error)
    Set(id string, content Content) error
    GetAll() (map[string]Content, error)
}
```

#### 3.11.3 Configuration Types

```go
type Config struct {
    LastUpdate               time.Time     // When data was last updated
    AutoRefreshInterval      time.Duration // How often to refresh
    AutoRefreshCheckInterval time.Duration // How often to check if refresh is needed
}

type ConfigTokenLists struct {
    Config
    TokenLists []types.ListDetails
}

type ConfigRemoteListOfTokenLists struct {
    Config
    RemoteListOfTokenListsFetchDetails types.ListDetails
    RemoteListOfTokenListsParser       parsers.ListOfTokenListsParser
}
```

### 3.12 Token Builder API (`pkg/tokens/builder`)

The token builder package implements the Builder pattern for incremental token collection construction.

#### 3.12.1 Core Interface

```go
type Builder struct {
    // Internal state - not directly accessible
}

func New(supportedChains []uint64) *Builder
func (b *Builder) AddNativeTokenList() error
func (b *Builder) AddTokenList(id string, tokenList *types.TokenList)
func (b *Builder) AddRawTokenList(id string, rawData []byte, sourceURL string, timestamp time.Time, parser parsers.TokenListParser) error
func (b *Builder) GetTokens() map[string]*types.Token
func (b *Builder) GetTokenLists() map[string]*types.TokenList
```

#### 3.12.2 Usage Example

```go
builder := builder.New([]uint64{1, 56, 10, 137}) // Ethereum, BSC, Optimism, Polygon
builder.AddNativeTokenList()
builder.AddTokenList("uniswap", uniswapList)
tokens := builder.GetTokens()
```

### 3.13 Token Manager API (`pkg/tokens/manager`)

The token manager package provides a high-level interface for comprehensive token management.

#### 3.13.1 Core Interface

```go
type Manager interface {
    // Lifecycle Management
    Start(ctx context.Context, enableAutoRefresh bool, notifyCh chan struct{}) error
    Stop()
    EnableAutoRefresh(ctx context.Context) error
    DisableAutoRefresh(ctx context.Context) error
    TriggerRefresh(ctx context.Context) error

    // Token Operations
    UniqueTokens() map[string]*types.Token
    GetTokenByChainAddress(chainID uint64, address common.Address) (*types.Token, bool)
    GetTokensByChain(chainID uint64) []*types.Token
    GetTokensByKeys(keys []string) []*types.Token
    TokenLists() map[string]*types.TokenList
    TokenList(id string) (*types.TokenList, bool)
}
```

#### 3.13.2 Configuration

```go
type Config struct {
    MainListID      string                                    // Primary token list ID
    InitialLists    map[string][]byte                         // Initial token list data
    CustomParsers   map[string]parsers.TokenListParser       // Custom parsers
    Chains          []uint64                                  // Supported chain IDs
    AutoFetcherConfig interface{}                             // AutoFetcher configuration
}
```

#### 3.13.3 Constructor

```go
func New(config *Config, fetcher fetcher.Fetcher, contentStore autofetcher.ContentStore, customTokenStore CustomTokenStore) (Manager, error)
```

### 3.14 ENS Resolver API (`pkg/ens`)

The ENS package provides Ethereum Name Service resolution.

#### 3.14.1 Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `NewResolver(client)` | Creates a new ENS resolver | `client`: `*ethclient.Client` | `*Resolver`, `error` |
| `IsSupportedChain(chainID)` | Checks if the chain ID supports ENS | `chainID`: `uint64` | `bool` |

#### 3.14.2 Resolver Methods

| Method | Purpose | Parameters | Returns |
|--------|---------|------------|---------|
| `AddressOf(name)` | Forward resolution (name → address) | `name`: `string` | `common.Address`, `error` |
| `GetName(address)` | Reverse resolution (address → name) | `address`: `common.Address` | `string`, `error` |

#### 3.14.3 Errors

| Error | Description |
|-------|-------------|
| `ErrInvalidName` | ENS name has invalid format (missing dot, starts/ends with dot) |
| `ErrEmptyName` | ENS name is empty |
| `ErrInvalidAddress` | Ethereum address is invalid (zero address) |

#### 3.14.4 Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `ENSRegistryAddress` | `0x00000000000C2E074eC69A0dFb2997BA6C7d2e1e` | ENS registry contract address (same on all supported chains) |

### 3.15 Transaction Generator API (`pkg/txgenerator`)

The transaction generator package provides functions for creating unsigned Ethereum transactions for various token operations.

#### 3.15.1 Core Functions

| Function | Purpose | Parameters | Returns |
|----------|---------|------------|---------|
| `TransferETH(params)` | Generate ETH transfer transaction | `params`: `TransferETHParams` | `*types.Transaction`, `error` |
| `TransferERC20(params)` | Generate ERC20 token transfer transaction | `params`: `TransferERC20Params` | `*types.Transaction`, `error` |
| `ApproveERC20(params)` | Generate ERC20 approval transaction | `params`: `ApproveERC20Params` | `*types.Transaction`, `error` |
| `TransferFromERC721(params)` | Generate ERC721 transferFrom transaction | `params`: `TransferERC721Params` | `*types.Transaction`, `error` |
| `SafeTransferFromERC721(params)` | Generate ERC721 safeTransferFrom transaction | `params`: `TransferERC721Params` | `*types.Transaction`, `error` |
| `ApproveERC721(params)` | Generate ERC721 approval transaction | `params`: `ApproveERC721Params` | `*types.Transaction`, `error` |
| `SetApprovalForAllERC721(params)` | Generate ERC721 setApprovalForAll transaction | `params`: `SetApprovalForAllERC721Params` | `*types.Transaction`, `error` |
| `TransferERC1155(params)` | Generate ERC1155 single token transfer transaction | `params`: `TransferERC1155Params` | `*types.Transaction`, `error` |
| `BatchTransferERC1155(params)` | Generate ERC1155 batch transfer transaction | `params`: `BatchTransferERC1155Params` | `*types.Transaction`, `error` |
| `SetApprovalForAllERC1155(params)` | Generate ERC1155 setApprovalForAll transaction | `params`: `SetApprovalForAllERC1155Params` | `*types.Transaction`, `error` |

#### 3.15.2 Parameter Types

```go
type BaseTxParams struct {
    Nonce                uint64
    GasLimit             uint64
    ChainID              *big.Int
    GasPrice             *big.Int  // For legacy transactions
    MaxFeePerGas         *big.Int  // For EIP-1559 transactions
    MaxPriorityFeePerGas *big.Int  // For EIP-1559 transactions
}

type TransferETHParams struct {
    BaseTxParams
    To    common.Address
    Value *big.Int
}

type TransferERC20Params struct {
    BaseTxParams
    TokenAddress common.Address
    To           common.Address
    Amount       *big.Int
}

type ApproveERC20Params struct {
    BaseTxParams
    TokenAddress common.Address
    Spender      common.Address
    Amount       *big.Int
}

type TransferERC721Params struct {
    BaseTxParams
    TokenAddress common.Address
    From         common.Address
    To           common.Address
    TokenID      *big.Int
}

type ApproveERC721Params struct {
    BaseTxParams
    TokenAddress common.Address
    To           common.Address
    TokenID      *big.Int
}

type SetApprovalForAllERC721Params struct {
    BaseTxParams
    TokenAddress common.Address
    Operator     common.Address
    Approved     bool
}

type TransferERC1155Params struct {
    BaseTxParams
    TokenAddress common.Address
    From         common.Address
    To           common.Address
    TokenID      *big.Int
    Value        *big.Int
}

type BatchTransferERC1155Params struct {
    BaseTxParams
    TokenAddress common.Address
    From         common.Address
    To           common.Address
    TokenIDs     []*big.Int
    Values       []*big.Int
}

type SetApprovalForAllERC1155Params struct {
    BaseTxParams
    TokenAddress common.Address
    Operator     common.Address
    Approved     bool
}
```

#### 3.15.3 Transaction Type Detection

The transaction type is automatically determined based on the provided parameters:

- **Legacy (type 0)**: If `GasPrice` is provided in `BaseTxParams`
- **EIP-1559 (type 2)**: If `MaxFeePerGas` or `MaxPriorityFeePerGas` is provided in `BaseTxParams`

For EIP-1559 transactions, both `MaxFeePerGas` and `MaxPriorityFeePerGas` must be provided.

#### 3.15.4 Error Types

| Error | Description |
|-------|-------------|
| `ErrInvalidParams` | Invalid transaction parameters (e.g., zero address, negative amount) |
| `ErrMissingGasPrice` | GasPrice required for legacy transactions |
| `ErrMissingMaxFeePerGas` | MaxFeePerGas required for EIP-1559 transactions |
| `ErrMissingMaxPriorityFeePerGas` | MaxPriorityFeePerGas required for EIP-1559 transactions |

## 4. Example Applications

### 4.1 Multicall Usage Example

The `examples/multiclient3-usage` folder demonstrates how to use the multicall package for efficient batch contract calls:

```go
package main

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/status-im/go-wallet-sdk/pkg/multicall"
    "github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

func main() {
    ctx := context.Background()

    // Get Multicall3 address for Ethereum Mainnet
    multicallAddr, exists := multicall3.GetMulticall3Address(1)
    if !exists {
        panic("Multicall3 not available on Ethereum Mainnet")
    }

    // Create caller (you would use your RPC client here)
    // caller := multicall3.NewMulticall3Caller(multicallAddr, rpcClient)

    // Build calls for multiple accounts and tokens
    accounts := []common.Address{
        common.HexToAddress("0x1234..."),
        common.HexToAddress("0x5678..."),
    }

    tokens := []common.Address{
        common.HexToAddress("0xA0b86a33E6441b8C4C8C0C4C0C4C0C4C0C4C0C4C0"), // USDC
        common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"), // DAI
    }

    // Create native balance job
    nativeCalls := make([]multicall3.IMulticall3Call, 0, len(accounts))
    for _, account := range accounts {
        nativeCalls = append(nativeCalls, multicall.BuildNativeBalanceCall(account, multicallAddr))
    }
    nativeJob := multicall.Job{
        Calls: nativeCalls,
        CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
            return multicall.ProcessNativeBalanceResult(result)
        },
    }

    // Create ERC20 balance job
    tokenCalls := make([]multicall3.IMulticall3Call, 0, len(accounts)*len(tokens))
    for _, account := range accounts {
        for _, token := range tokens {
            tokenCalls = append(tokenCalls, multicall.BuildERC20BalanceCall(account, token))
        }
    }
    tokenJob := multicall.Job{
        Calls: tokenCalls,
        CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
            return multicall.ProcessERC20BalanceResult(result)
        },
    }

    // Execute jobs synchronously
    jobs := []multicall.Job{nativeJob, tokenJob}
    results := multicall.RunSync(ctx, jobs, nil, caller, 100)

    // Process native balance results
    for i, callResult := range results[0].Results {
        if callResult.Err != nil {
            fmt.Printf("Error processing native balance for account %d: %v\n", i, callResult.Err)
            continue
        }
        balance := callResult.Value.(*big.Int)
        fmt.Printf("Account %s native balance: %s wei\n", accounts[i].Hex(), balance.String())
    }

    // Process token balance results
    callIndex := 0
    for _, account := range accounts {
        for _, token := range tokens {
            callResult := results[1].Results[callIndex]
            if callResult.Err != nil {
                fmt.Printf("Error processing token balance: %v\n", callResult.Err)
                callIndex++
                continue
            }
            balance := callResult.Value.(*big.Int)
            fmt.Printf("Account %s token %s balance: %s\n", account.Hex(), token.Hex(), balance.String())
            callIndex++
        }
    }

    // Alternative: Execute jobs asynchronously
    // resultsCh := multicall.RunAsync(ctx, jobs, nil, caller, 100)
    // for result := range resultsCh {
    //     jobIdx := result.JobIdx
    //     jobResult := result.JobResult
    //     // Process results as they become available
    // }
}
```

### 4.2 Web‑Based Balance Fetcher

The `examples/balance-fetcher-web` folder contains a complete web application that demonstrates how to use the balance fetcher. Key aspects include:
- **Features** – The web UI allows users to specify custom chains (chain ID and RPC URL), enter multiple Ethereum addresses and an optional block number, then fetch balances across chains. It supports batch fetching for native tokens, automatic fallback to standard RPC, and displays balances in both ETH and wei. The example pre‑populates common chains such as Ethereum, Optimism, Arbitrum and Polygon.
- **Usage** – Running `go run .` in the example directory starts an HTTP server on `localhost:8080`. Users can configure chains, input addresses and click Fetch Balances. The backend sends a `POST /fetch` request containing a JSON payload with chains, addresses and block number.
- **Project Structure** – The example is organised into `main.go` (entry point), `types.go` (data structures), `rpc_client.go` (custom RPC client), `utils.go`, `templates.go` (HTML/JS templates), and `handlers.go` (HTTP handlers).
- **Security Considerations** – The example warns that it is for demonstration only. Production deployments should secure RPC endpoints, implement authentication, validate user input and add rate‑limiting.

### 4.3 Ethereum Client Example

The `examples/ethclient-usage` folder shows how to use the Ethereum client across multiple networks. It exercises a wide range of RPC methods and demonstrates multi‑endpoint support.

- **Features** – The example tests connectivity and functionality across multiple RPC endpoints, retrieves network and blockchain data, account balances and nonces, contract code, filters events, retrieves transaction details, checks network status, and estimates gas. It highlights the chain‑agnostic benefits of the custom **eth.go** methods, which make no assumptions about transaction types or chain‑specific fields.

- **Usage** – Users specify one or more RPC endpoints via the **ETH_RPC_ENDPOINTS** environment variable and run **go run main.go**. The program iterates through each endpoint, prints client and network information, queries blocks and transactions, and demonstrates various API calls. Example output shows block and transaction details, balances, gas prices and event logs.

- **Configuration** – The example includes defaults for Ethereum Mainnet, Optimism, Arbitrum and Sepolia but can be configured to use Infura, Alchemy or local nodes by setting `ETH_RPC_ENDPOINTS` ENV variable.

- **Code Structure** – The example is split into `main.go`, which loops over endpoints, and helper functions such as `testRPC()` that call various methods and handle errors.

### 4.4 Multi-Standard Fetcher Example

The `examples/multistandardfetcher-example` folder demonstrates how to use the multistandardfetcher package to fetch balances across all token standards (Native ETH, ERC20, ERC721, ERC1155) for a specific address using Multicall3 batched calls.

- **Features** – The example fetches native ETH balance, queries ERC20 token balances for popular tokens (USDC, DAI, USDT, WBTC, LINK, UNI, MATIC, SHIB), checks ERC721 NFT balances for well-known collections (BAYC, MAYC, CryptoPunks, Azuki, Moonbirds, Doodles), and retrieves ERC1155 collectible balances from popular contracts. It displays results in a formatted report with token symbols and readable balances.

- **Usage** – Users set the `RPC_URL` environment variable and run the example. The program automatically queries vitalik.eth's balances across all token standards and displays a comprehensive report showing native ETH, ERC20 tokens, ERC721 NFTs, and ERC1155 collectibles with non-zero balances.

- **Multi-Standard Support** – Demonstrates the unified interface for fetching balances across different token standards in a single operation, leveraging the underlying multicall package for efficient batching.

- **Output Format** – The example displays a clean, formatted report with sections for each token type, showing token symbols, balances, and summary statistics. It includes proper error handling and graceful degradation when calls fail.

- **Integration** – The example demonstrates the seamless integration between the `multistandardfetcher` and `multicall` packages, showing how to efficiently fetch balances across multiple token standards with minimal RPC calls.

### 4.5 Event Filter Example

The `examples/eventfilter-example` folder demonstrates how to use the event filter and event log parser packages to detect and display transfer events for specific accounts with concurrent processing.

- **Features** – The example provides a command-line interface with flexible options for filtering transfer events. It supports multi-token filtering (ERC20, ERC721, and ERC1155), direction-based filtering (send, receive, or both), and comprehensive transfer details extraction. The example uses the new `FilterTransfers` function for concurrent processing of multiple filter queries, enhanced formatting with shortened addresses and scientific notation for large numbers, raw event metadata including event signatures and log properties, and debug information showing contract keys and unpacked types.

- **Usage** – Users can specify an account address, block range, and optional RPC endpoint. The example supports filtering by direction and displays detailed information about each transfer event found. Command-line options include `-account` (required), `-start` and `-end` block numbers (required), `-rpc` for custom endpoints, and `-direction` for filtering.

- **Concurrent Processing** – The example leverages the `FilterTransfers` function which automatically handles concurrent execution of multiple filter queries, improving performance when scanning large block ranges or multiple token types.

- **Output Format** – The example displays transfers grouped by token type with comprehensive details extracted from the `Unpacked` field. It shows block numbers, transaction hashes, addresses, amounts, token IDs, contract addresses, log indices, event signatures, and other metadata. Raw event data is also displayed for debugging purposes.

- **Integration** – The example demonstrates the seamless integration between the `eventfilter` and `eventlog` packages, showing how to filter events and parse them into type-safe Go structs for application use with improved performance through concurrent processing.

### 4.6 Gas Comparison Example

The `examples/gas-comparison` folder demonstrates gas fee estimation across multiple networks and compares different implementations with comprehensive analysis tools.

- **Features** – The example provides a multi-network gas fee comparison tool that tests gas estimation accuracy across Ethereum, Arbitrum, Optimism, Base, Polygon, Linea, BSC, and Status Network Sepolia. It compares three implementations: the new `GetTxSuggestions` API, a legacy estimator, and Infura's Gas API. The tool displays comprehensive analysis including priority fees, max fees, base fees, wait times, and network congestion metrics.

- **Usage** – Users can run with local mock data (`-fake` flag) for testing without network access, or with real networks by providing an Infura API key (`-infura-api-key YOUR_KEY`). The tool automatically configures chain-specific parameters including block times and estimation strategies for each network. Example output shows detailed comparisons in wei with percentage differences between implementations.

- **Chain-Specific Parameters** – The example demonstrates proper configuration for different chain classes:
  - **Ethereum Mainnet**: 12s block time, 10 blocks for estimation, L1 congestion analysis
  - **Arbitrum One**: 0.25s block time, 50 blocks for estimation, ArbStack optimizations
  - **Optimism/Base**: 2s block time, 50 blocks for estimation, OPStack fixed multipliers
  - **Polygon**: 2.25s block time, 50 blocks for estimation, L1 congestion analysis
  - **Linea/Status Network**: 2s block time, 50 blocks for estimation, LineaStack with `linea_estimateGas`

- **Test Transaction** – Uses a simple 0-valued ETH transfer from Vitalik's address (`0xd8da6bf26964af9d7eed9e03e53415d37aa96045`) to the zero address for consistent testing across networks. This minimal transaction allows focus on fee estimation accuracy without contract complexity.

- **Mock Data Support** – Includes pre-captured network data for offline testing. The `data/generator` subdirectory provides tools to regenerate test data by fetching fresh blockchain data including latest blocks with full transactions, 1024 blocks of fee history, current gas prices, and Infura fee suggestions.

- **Output Format** – Displays comprehensive comparison results showing fee suggestions for three priority levels, time estimates with min/max ranges in seconds, network congestion scores (L1 only), and percentage differences between implementations. Results help validate estimation accuracy and identify optimization opportunities.

- **Data Generator** – The `data/generator` tool captures real blockchain data for testing. Users run `go run main.go -rpc YOUR_RPC_URL` to fetch data from any EVM chain. The tool automatically detects chain ID and generates chain-specific Go code with embedded test data for reproducible offline testing.

### 4.7 Accounts Example

The `examples/accounts` folder demonstrates how to use the extended keystore and mnemonic packages with an interactive web interface for testing keystore functionality.

- **Features** – The web application provides a comprehensive testing environment for both extended keystore and standard keystore implementations. It includes mnemonic phrase generation, account creation from mnemonics, child account derivation using BIP44 paths, import/export of keys in various formats, message signing, account unlocking/locking, and account deletion. The interface is split into two sections to facilitate testing import/export functionality between the two keystore types.

- **Usage** – Running `go run .` in the example directory starts an HTTP server on `localhost:8081`. The web interface provides dropdown selectors for account addresses, displays account information including file paths and keystore file contents, and shows comprehensive error messages for all operations.

- **Keystore Management** – Demonstrates full account lifecycle management:
  - Generate random mnemonic phrases (12, 15, 18, 21, or 24 words) in a separate top section
  - Create accounts from mnemonic phrases with optional passphrase encryption
  - Derive child accounts from parent accounts using BIP44 derivation paths
  - Import/export keys in extended keystore format or standard private key format
  - Sign messages with accounts (with or without passphrase)
  - Unlock accounts with configurable timeout or lock them
  - Delete accounts with passphrase confirmation
  - View account information including file paths and keystore file contents

- **Integration Testing** – The two-column layout allows side-by-side testing of extended keystore and standard keystore, making it easy to test import/export functionality between the two implementations. Users can export a key from one keystore type and import it into the other.

- **Code Structure** – The example is organized into `main.go` (server and API handlers), `templates.go` (HTML/JavaScript templates), and `go.mod` (dependency management). It uses gorilla/mux for routing and Go's html/template package for rendering.

### 4.8 C Application Example

The `examples/c-app` folder demonstrates how to use the Go Wallet SDK from C applications using the shared library.

- **Features** – The example shows how to create an Ethereum client, retrieve the chain ID, fetch account balances, make raw JSON-RPC calls, and fetch multi-standard token balances (Native ETH, ERC20, ERC721, ERC1155) from a C application. It demonstrates proper memory management by freeing all C strings returned by the SDK functions. The example uses the generated C header file (`libgowalletsdk.h`) and links against the shared library. The multi-standard balance fetcher example shows how to construct the fetch configuration JSON with examples for all token types including ERC1155 collectibles in the `"contractAddress:tokenID"` format.

- **Usage** – Users first build the shared library from the repository root using `make shared-library`, then build and run the C example using the provided Makefile. The example connects to a public Ethereum RPC endpoint and queries Vitalik's address for demonstration purposes.

- **Build Process** – The example includes a Makefile that handles linking against the shared library, setting appropriate rpath on macOS for library loading, and copying the library next to the executable for convenience. The build process demonstrates proper integration of the Go-compiled shared library with C applications.

- **Memory Management** – The example demonstrates critical memory management practices: all string values returned by GoWSK functions must be freed using `GoWSK_FreeCString` to prevent memory leaks. Error messages passed through `errOut` parameters must also be freed if they are not NULL.

- **Code Structure** – The example consists of `main.c` (the C application), `Makefile` (build configuration), and `README.md` (usage instructions). It shows a minimal but complete integration of the SDK's C API.

### 4.9 Token Builder Example

The `examples/token-builder` folder demonstrates how to use the token builder package for incremental token collection building:

- **Features** – The example shows incremental building starting with empty state, automatic native token generation for supported chains (ETH, BNB, etc.), automatic deduplication using chain ID and address combinations, raw token list processing with various parsers, and advanced builder patterns including validation and error handling.
- **Usage** – Running `go run .` demonstrates basic builder usage, incremental building patterns, raw token list processing, token deduplication, and advanced builder patterns with comprehensive error handling examples.
- **Key Concepts** – Demonstrates the Builder pattern with stateful construction, token deduplication using unique keys, native token support for multiple chains, and performance characteristics including time complexity and memory usage.
- **Integration** – Shows how the builder integrates with parsers to handle different token list formats and provides a foundation for building token collections in blockchain applications.


### 4.10 Token Fetcher Example

The `examples/token-fetcher` folder demonstrates HTTP-based token list fetching with production-ready features:

- **Features** – Shows single token list fetching, concurrent fetching of multiple token lists using goroutines, HTTP ETag caching for bandwidth efficiency, list-of-token-lists fetching for discovery patterns, and robust error handling for network failures and invalid responses.
- **Usage** – Running `go run .` demonstrates single fetch operations, concurrent fetching with performance improvements, ETag-based caching with 304 Not Modified responses, and list-of-token-lists processing for dynamic token list discovery.
- **Performance** – Includes benchmarks showing typical performance metrics, memory usage patterns, and optimization strategies for production deployments.
- **Integration** – Demonstrates integration with token manager and background refresh services for automated token list management.

### 4.11 Token Manager Example

The `examples/token-manager` folder demonstrates comprehensive token management across multiple blockchain networks:

- **Features** – Shows complete token management with multi-source integration (native, remote, local, custom tokens), thread-safe concurrent access, rich query capabilities by chain/address/list ID, custom token support, and automatic refresh management with configurable intervals.
- **Usage** – Running `go run .` demonstrates manager configuration with multiple token sources, HTTP fetcher setup, storage backend implementations, token operations including search and filtering, and custom token management.
- **Production Considerations** – Includes examples of database-backed content stores, file-based custom token stores, dynamic auto-refresh management, and monitoring/observability patterns.
- **Integration** – Shows wallet service integration patterns and demonstrates how the manager centralizes token operations for wallet applications.

### 4.12 Token Parser Example

The `examples/token-parser` folder demonstrates parsing different token list formats from various sources:

- **Features** – Shows multiple parser types including Standard (Uniswap-style), Status-specific with chain grouping, CoinGecko API with platform mappings, and list-of-token-lists metadata parsing, with comprehensive input validation and error handling.
- **Usage** – Running `go run .` demonstrates parser selection strategies, chain filtering, error handling for various scenarios, and format comparison showing different token list structures and their use cases.
- **Performance** – Includes performance characteristics comparison between parsers, memory usage patterns, and processing speed benchmarks for different formats.
- **Integration** – Shows integration with token manager and token fetcher, batch processing patterns, and parser selection strategies for different data sources.

### 4.13 Transaction Generator Example

The `examples/txgenerator-example` folder demonstrates a web-based interface for generating unsigned Ethereum transactions:

- **Features** – The web UI allows users to select transaction types (ETH transfers, ERC20, ERC721, ERC1155 operations), choose fee types (Legacy or EIP-1559), fill in transaction parameters, and generate unsigned transactions in JSON format. Supports all major token standards with comprehensive parameter validation.
- **Usage** – Running `go run .` in the example directory starts an HTTP server on `localhost:8080`. Users can select transaction types from a dropdown, configure gas parameters, fill in transaction-specific fields, and click "Generate Transaction" to receive the unsigned transaction as JSON.
- **Project Structure** – The example is organized into `main.go` (entry point and server setup), `handlers.go` (transaction generation logic), and `templates.go` (HTML templates and frontend JavaScript).
- **Transaction Types** – Supports native ETH transfers, ERC20 transfers and approvals, ERC721 transfers (transferFrom and safeTransferFrom), approvals, and operator management, as well as ERC1155 single and batch transfers and operator management.
- **Response Format** – Generated transactions are returned as JSON with all standard transaction fields including type, nonce, gas parameters, to, value, data, chainID, and signature fields (v, r, s) set to zero for unsigned transactions.

## 5. Testing & Development

### 5.1 Fetching  SDK

Developers can fetch the SDK by running:

```bash
go get github.com/status-im/go-wallet-sdk
```

### 5.2 Running Tests

All packages are fully tested. To run the tests do:

```bash
go test ./...
```

This executes unit tests for the balance fetcher and Ethereum client. The balance fetcher includes a `mock` package to simulate RPC responses. The repository also includes continuous integration workflows (`.github/workflows`) and static analysis configurations (`.golangci.yml`).

### 5.3 Building the C Library

The SDK includes build support for creating C libraries (shared or static) that expose core functionality to non-Go applications. The C bindings are implemented in the `clib` package, which provides C-compatible functions for interacting with the Ethereum client.

#### 5.3.1 Package Structure

The `clib` package consists of:
- `c.go` - Memory management utilities for C strings (`GoWSK_FreeCString`)
- `error.go` - Error handling utilities for C bindings
- `ethclient.go` - C bindings for Ethereum client functionality
- `balance_multistandardfetcher.go` - C bindings for multi-standard balance fetching (Native ETH, ERC20, ERC721, ERC1155). Includes internal string conversion utilities for CollectibleID types used in JSON serialization.
- `accounts_extkeystore.go` - C bindings for extended keystore operations (HD wallet support with BIP32 extended keys)
- `accounts_keystore.go` - C bindings for standard keystore operations (go-ethereum compatible keystore)
- `accounts_keys.go` - C bindings for key derivation and conversion utilities (extended keys, ECDSA keys, public keys, addresses)
- `accounts_mnemonic.go` - C bindings for BIP39 mnemonic phrase generation and utilities
- `main.go` - Entry point for building the library

#### 5.3.2 Building the Library

The SDK supports building both shared and static libraries:

**Shared Library:**
```bash
make shared-library
```

This creates:
- `build/libgowalletsdk.dylib` (macOS) or `build/libgowalletsdk.so` (Linux)
- `build/libgowalletsdk.h` (C header file with exported function declarations)

The build process:
1. Checks for Go installation
2. Compiles the `clib` package with `-buildmode=c-shared`
3. Generates the shared library and header file in the `build/` directory

**Static Library:**
```bash
make static-library
```

This creates:
- `build/libgowalletsdk.a` (static library)
- `build/libgowalletsdk.h` (C header file with exported function declarations)

The build process:
1. Checks for Go installation
2. Compiles the `clib` package with `-buildmode=c-archive`
3. Generates the static library and header file in the `build/` directory

#### 5.3.3 C API Functions

The C library (shared and static) exports the following functions:

**Memory Management:**
- `void GoWSK_FreeCString(char* s)` - Frees C strings returned by GoWSK functions to prevent memory leaks. Must be called for all string return values.

**Ethereum Client:**
- `uintptr_t GoWSK_ethclient_NewClient(char* rpcURL, char** errOut)` - Creates a new Ethereum client instance. Returns a handle (uintptr_t) on success, or 0 on error. If `errOut` is provided and an error occurs, it will contain an error message (must be freed with `GoWSK_FreeCString`).
- `void GoWSK_ethclient_CloseClient(uintptr_t handle)` - Closes and cleans up an Ethereum client instance. The handle becomes invalid after this call.
- `char* GoWSK_ethclient_ChainID(uintptr_t handle, char** errOut)` - Returns the chain ID as a string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`. If `errOut` is provided and an error occurs, it will contain an error message (must be freed with `GoWSK_FreeCString`).
- `char* GoWSK_ethclient_GetBalance(uintptr_t handle, char* address, char** errOut)` - Returns the balance of an address in wei as a string. The address should be a hex string (e.g., "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"). Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`. If `errOut` is provided and an error occurs, it will contain an error message (must be freed with `GoWSK_FreeCString`).
- `char* GoWSK_ethclient_RPCCall(uintptr_t handle, char* method, char* params, char** errOut)` - Executes a raw JSON-RPC call. The `method` parameter should be the RPC method name (e.g., `"eth_getBalance"`). The `params` parameter should be a JSON array string with the method parameters (e.g., `"[\"0x...\",\"latest\"]"`). Returns the JSON-RPC response as a string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`. If `errOut` is provided and an error occurs, it will contain an error message (must be freed with `GoWSK_FreeCString`).

**Multi-Standard Balance Fetcher:**
- `char* GoWSK_balance_multistandardfetcher_FetchBalances(uintptr_t ethClientHandle, unsigned long chainID, unsigned long batchSize, char* fetchConfigJSON, uintptr_t* cancelHandleOut, char** errOut)` - Fetches balances across multiple token standards (Native ETH, ERC20, ERC721, ERC1155) using Multicall3 batched calls. The `fetchConfigJSON` parameter should be a JSON string with the configuration (see format below). The `cancelHandleOut` parameter is an output parameter that receives a cancel handle if provided (can be NULL if cancellation is not needed). This handle can be used to cancel the operation. Returns a JSON string with the results. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`. If `errOut` is provided and an error occurs, it will contain an error message (must be freed with `GoWSK_FreeCString`).
- `void GoWSK_balance_multistandardfetcher_CancelFetchBalances(uintptr_t cancelHandle)` - Cancels an ongoing fetch operation. The `cancelHandle` should be the value obtained from the `cancelHandleOut` parameter of `GoWSK_balance_multistandardfetcher_FetchBalances`. This will stop all goroutines associated with the fetch operation. Safe to call multiple times.
- `void GoWSK_balance_multistandardfetcher_FreeCancelHandle(uintptr_t cancelHandle)` - Frees the cancel handle and associated resources. **Must be called to free the cancel handle in all cases, including if `GoWSK_balance_multistandardfetcher_FetchBalances` returns NULL due to an error.** The handle is created before the fetch starts, so it must be freed whether the fetch operation completes successfully, is cancelled, or fails with an error, to prevent memory leaks.

The fetch configuration JSON format:
```json
{
  "native": ["0x...", "0x..."],
  "erc20": {
    "0xAccount...": ["0xToken1...", "0xToken2..."]
  },
  "erc721": {
    "0xAccount...": ["0xNFTContract1...", "0xNFTContract2..."]
  },
  "erc1155": {
    "0xAccount...": ["0xContract1:tokenID1", "0xContract2:tokenID2"]
  }
}
```

Note: ERC1155 collectibles use the format `"contractAddress:tokenID"` where tokenID is a decimal string. Both contract address and token ID must be non-empty. The function returns results as a JSON array of result objects with fields: `resultType`, `account`, `result` (for native), `results` (for ERC20/ERC721/ERC1155), `err`, `atBlockNumber`, and `atBlockHash`.

**Mnemonic Utilities:**
- `char* GoWSK_accounts_mnemonic_CreateRandomMnemonic(int length, char** errOut)` - Creates a random BIP39 mnemonic phrase with the specified word count (12, 15, 18, 21, or 24). Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_mnemonic_CreateRandomMnemonicWithDefaultLength(char** errOut)` - Creates a random BIP39 mnemonic phrase with the default length (12 words). Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `uint32_t GoWSK_accounts_mnemonic_LengthToEntropyStrength(int length, char** errOut)` - Converts a mnemonic word count to its corresponding entropy strength in bits. Returns 0 on error.

**Key Derivation and Conversion:**
- `char* GoWSK_accounts_keys_CreateExtKeyFromMnemonic(char* phrase, char* passphrase, char** errOut)` - Creates a BIP32 extended key from a BIP39 mnemonic phrase. The `passphrase` parameter is optional (can be NULL) and is used for BIP39 seed extension. Returns the extended key as a base58-encoded string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keys_DeriveExtKey(char* extKeyStr, char* pathStr, char** errOut)` - Derives a child extended key from a parent extended key using a BIP32 derivation path (e.g., "m/44'/60'/0'/0/0"). Returns the derived extended key as a base58-encoded string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keys_ExtKeyToECDSA(char* extKeyStr, char** errOut)` - Converts a BIP32 extended key to an ECDSA private key. Returns the private key as a hex-encoded string (without "0x" prefix). Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keys_ECDSAToPublicKey(char* privateKeyHex, char** errOut)` - Derives the public key from an ECDSA private key. The `privateKeyHex` should be a hex-encoded string (with or without "0x" prefix). Returns the public key as a hex-encoded string (without "0x" prefix). Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keys_PublicKeyToAddress(char* publicKeyHex, char** errOut)` - Derives an Ethereum address from a public key. The `publicKeyHex` should be a hex-encoded string (with or without "0x" prefix). Returns the address as a hex-encoded string (with "0x" prefix). Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.

**Extended Keystore:**
- `uintptr_t GoWSK_accounts_extkeystore_NewKeyStore(char* keydir, int scryptN, int scryptP, char** errOut)` - Creates a new extended keystore instance. The `keydir` parameter specifies the directory where encrypted keys will be stored. `scryptN` and `scryptP` are scrypt parameters for key encryption (use 262144 and 1 for standard, or 4096 and 6 for light). Returns a handle on success, or 0 on error.
- `void GoWSK_accounts_extkeystore_CloseKeyStore(uintptr_t handle)` - Closes and cleans up an extended keystore instance.
- `char* GoWSK_accounts_extkeystore_Accounts(uintptr_t handle, char** errOut)` - Returns a JSON array of all accounts in the keystore. Each account object contains `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_NewAccount(uintptr_t handle, char* passphrase, char** errOut)` - Creates a new account in the extended keystore. The `passphrase` parameter is optional (can be NULL). Returns a JSON object with `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_Import(uintptr_t handle, char* keyJSON, char* oldPassphrase, char* newPassphrase, char** errOut)` - Imports an account from a JSON-encoded key file (standard keystore format). Returns a JSON object with `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_ImportExtendedKey(uintptr_t handle, char* extKeyStr, char* passphrase, char** errOut)` - Imports an extended key into the keystore. The `extKeyStr` should be a base58-encoded extended key string. Returns a JSON object with `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_ExportExt(uintptr_t handle, char* address, char* passphrase, char* newPassphrase, char** errOut)` - Exports an account's extended key. Returns a JSON-encoded key file. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_ExportPriv(uintptr_t handle, char* address, char* passphrase, char* newPassphrase, char** errOut)` - Exports an account's private key in standard keystore format. Returns a JSON-encoded key file. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `void GoWSK_accounts_extkeystore_Delete(uintptr_t handle, char* address, char* passphrase, char** errOut)` - Deletes an account from the keystore. Requires the account's passphrase for confirmation.
- `int GoWSK_accounts_extkeystore_HasAddress(uintptr_t handle, char* address, char** errOut)` - Checks if the keystore contains an account with the given address. Returns 1 if found, 0 otherwise.
- `void GoWSK_accounts_extkeystore_Unlock(uintptr_t handle, char* address, char* passphrase, char** errOut)` - Unlocks an account indefinitely. The account remains unlocked until explicitly locked.
- `void GoWSK_accounts_extkeystore_Lock(uintptr_t handle, char* address, char** errOut)` - Locks an account immediately.
- `void GoWSK_accounts_extkeystore_TimedUnlock(uintptr_t handle, char* address, char* passphrase, unsigned long timeout, char** errOut)` - Unlocks an account for a specified duration (in seconds). The account will automatically lock after the timeout expires.
- `void GoWSK_accounts_extkeystore_Update(uintptr_t handle, char* address, char* oldPassphrase, char* newPassphrase, char** errOut)` - Changes the passphrase for an account.
- `char* GoWSK_accounts_extkeystore_SignHash(uintptr_t handle, char* address, char* hash, char** errOut)` - Signs a hash with an unlocked account. The `hash` should be a hex-encoded string (with or without "0x" prefix). Returns the signature as a hex-encoded string (without "0x" prefix). Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_SignHashWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* hash, char** errOut)` - Signs a hash with an account using a passphrase (does not require the account to be unlocked). Returns the signature as a hex-encoded string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_SignTx(uintptr_t handle, char* address, char* txJSON, char* chainIDHex, char** errOut)` - Signs a transaction with an unlocked account. The `txJSON` should be a JSON-encoded transaction object. The `chainIDHex` parameter should be a hex-encoded chain ID (with or without "0x" prefix). Returns the signed transaction as a hex-encoded RLP string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_SignTxWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* txJSON, char* chainIDHex, char** errOut)` - Signs a transaction with an account using a passphrase and a specified chain ID. The `chainIDHex` parameter should be a hex-encoded string representing the chain ID (with or without "0x" prefix) for EIP-155 transaction signing. Returns the signed transaction as a hex-encoded RLP string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_Derive(uintptr_t handle, char* address, char* pathStr, char* pin, char** errOut)` - Derives a child account from a parent account and optionally pins it to the keystore. The `pin` parameter determines if the derived account should be saved (non-NULL) or ephemeral (NULL). Returns a JSON object with `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_DeriveWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* pathStr, char* pin, char** errOut)` - Derives a child account using a passphrase. Returns a JSON object with `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_extkeystore_Find(uintptr_t handle, char* address, char* url, char** errOut)` - Finds an account by address and URL. Returns a JSON object with `address` and `url` fields, or NULL if not found. The returned string must be freed with `GoWSK_FreeCString`.

**Transaction Generator:**
- `char* GoWSK_txgenerator_TransferETH(char* paramsJSON, char** errOut)` - Generates an unsigned ETH transfer transaction. The `paramsJSON` should be a JSON string with transaction parameters (nonce, gasLimit, chainID, gasPrice or maxFeePerGas/maxPriorityFeePerGas, to, value). Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_TransferERC20(char* paramsJSON, char** errOut)` - Generates an unsigned ERC20 token transfer transaction. The `paramsJSON` should include tokenAddress, to, and amount fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_ApproveERC20(char* paramsJSON, char** errOut)` - Generates an unsigned ERC20 approval transaction. The `paramsJSON` should include tokenAddress, spender, and amount fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_TransferFromERC721(char* paramsJSON, char** errOut)` - Generates an unsigned ERC721 transferFrom transaction. The `paramsJSON` should include tokenAddress, from, to, and tokenID fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_SafeTransferFromERC721(char* paramsJSON, char** errOut)` - Generates an unsigned ERC721 safeTransferFrom transaction. The `paramsJSON` should include tokenAddress, from, to, and tokenID fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_ApproveERC721(char* paramsJSON, char** errOut)` - Generates an unsigned ERC721 approval transaction. The `paramsJSON` should include tokenAddress, to, and tokenID fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_SetApprovalForAllERC721(char* paramsJSON, char** errOut)` - Generates an unsigned ERC721 setApprovalForAll transaction. The `paramsJSON` should include tokenAddress, operator, and approved fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_TransferERC1155(char* paramsJSON, char** errOut)` - Generates an unsigned ERC1155 single token transfer transaction. The `paramsJSON` should include tokenAddress, from, to, tokenID, and value fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_BatchTransferERC1155(char* paramsJSON, char** errOut)` - Generates an unsigned ERC1155 batch transfer transaction. The `paramsJSON` should include tokenAddress, from, to, tokenIDs (array), and values (array) fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_txgenerator_SetApprovalForAllERC1155(char* paramsJSON, char** errOut)` - Generates an unsigned ERC1155 setApprovalForAll transaction. The `paramsJSON` should include tokenAddress, operator, and approved fields. Returns the transaction as a JSON string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.

**Standard Keystore:**
- `uintptr_t GoWSK_accounts_keystore_NewKeyStore(char* keydir, int scryptN, int scryptP, char** errOut)` - Creates a new standard keystore instance (go-ethereum compatible). Returns a handle on success, or 0 on error.
- `void GoWSK_accounts_keystore_CloseKeyStore(uintptr_t handle)` - Closes and cleans up a standard keystore instance.
- `char* GoWSK_accounts_keystore_Accounts(uintptr_t handle, char** errOut)` - Returns a JSON array of all accounts in the keystore. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keystore_NewAccount(uintptr_t handle, char* passphrase, char** errOut)` - Creates a new account in the standard keystore. Returns a JSON object with `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keystore_Import(uintptr_t handle, char* keyJSON, char* passphrase, char* newPassphrase, char** errOut)` - Imports an account from a JSON-encoded key file. Returns a JSON object with `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keystore_Export(uintptr_t handle, char* address, char* passphrase, char* newPassphrase, char** errOut)` - Exports an account's private key. Returns a JSON-encoded key file. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `void GoWSK_accounts_keystore_Delete(uintptr_t handle, char* address, char* passphrase, char** errOut)` - Deletes an account from the keystore.
- `int GoWSK_accounts_keystore_HasAddress(uintptr_t handle, char* address, char** errOut)` - Checks if the keystore contains an account with the given address. Returns 1 if found, 0 otherwise.
- `void GoWSK_accounts_keystore_Unlock(uintptr_t handle, char* address, char* passphrase, char** errOut)` - Unlocks an account indefinitely.
- `void GoWSK_accounts_keystore_Lock(uintptr_t handle, char* address, char** errOut)` - Locks an account immediately.
- `void GoWSK_accounts_keystore_TimedUnlock(uintptr_t handle, char* address, char* passphrase, unsigned long timeout, char** errOut)` - Unlocks an account for a specified duration (in seconds). The account will automatically lock after the timeout expires.
- `void GoWSK_accounts_keystore_Update(uintptr_t handle, char* address, char* oldPassphrase, char* newPassphrase, char** errOut)` - Changes the passphrase for an account.
- `char* GoWSK_accounts_keystore_SignHash(uintptr_t handle, char* address, char* hash, char** errOut)` - Signs a hash with an unlocked account. Returns the signature as a hex-encoded string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keystore_SignHashWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* hash, char** errOut)` - Signs a hash with an account using a passphrase. Returns the signature as a hex-encoded string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keystore_ImportECDSA(uintptr_t handle, char* privateKeyHex, char* passphrase, char** errOut)` - Imports a private key (ECDSA) into the keystore. The `privateKeyHex` should be a hex-encoded string. Returns a JSON object with `address` and `url` fields. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keystore_SignTx(uintptr_t handle, char* address, char* txJSON, char* chainIDHex, char** errOut)` - Signs a transaction with an unlocked account. The `chainIDHex` parameter should be a hex-encoded string representing the chain ID. Returns the signed transaction as a hex-encoded RLP string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keystore_SignTxWithPassphrase(uintptr_t handle, char* address, char* passphrase, char* txJSON, char* chainIDHex, char** errOut)` - Signs a transaction with an account using a passphrase and chain ID. Returns the signed transaction as a hex-encoded RLP string. Returns NULL on error. The returned string must be freed with `GoWSK_FreeCString`.
- `char* GoWSK_accounts_keystore_Find(uintptr_t handle, char* address, char* url, char** errOut)` - Finds an account by address and URL. Returns a JSON object with `address` and `url` fields, or NULL if not found. The returned string must be freed with `GoWSK_FreeCString`.

#### 5.3.4 Usage Example

```c
#include "libgowalletsdk.h"
#include <stdio.h>
#include <stdlib.h>

int main() {
    char* err = NULL;

    // Create client
    uintptr_t handle = GoWSK_ethclient_NewClient("https://ethereum-rpc.publicnode.com", &err);
    if (handle == 0) {
        fprintf(stderr, "Failed to create client: %s\n", err ? err : "unknown error");
        if (err) GoWSK_FreeCString(err);
        return 1;
    }

    // Get chain ID
    char* chainID = GoWSK_ethclient_ChainID(handle, &err);
    if (chainID == NULL) {
        fprintf(stderr, "ChainID error: %s\n", err ? err : "unknown error");
        if (err) GoWSK_FreeCString(err);
        GoWSK_ethclient_CloseClient(handle);
        return 1;
    }
    printf("ChainID: %s\n", chainID);
    GoWSK_FreeCString(chainID);

    // Get balance
    char* balance = GoWSK_ethclient_GetBalance(handle, "0x0000000000000000000000000000000000000000", &err);
    if (balance == NULL) {
        fprintf(stderr, "GetBalance error: %s\n", err ? err : "unknown error");
        if (err) GoWSK_FreeCString(err);
        GoWSK_ethclient_CloseClient(handle);
        return 1;
    }
    printf("Balance (wei): %s\n", balance);
    GoWSK_FreeCString(balance);

    // Make a raw RPC call
    char* method = "eth_getBalance";
    char* params = "[\"0x0000000000000000000000000000000000000000\",\"latest\"]";
    char* rpcResponse = GoWSK_ethclient_RPCCall(handle, method, params, &err);
    if (rpcResponse == NULL) {
        fprintf(stderr, "RPCCall error: %s\n", err ? err : "unknown error");
        if (err) GoWSK_FreeCString(err);
        GoWSK_ethclient_CloseClient(handle);
        return 1;
    }
    printf("RPC Call Response: %s\n", rpcResponse);
    GoWSK_FreeCString(rpcResponse);

    // Clean up
    GoWSK_ethclient_CloseClient(handle);
    return 0;
}
```

#### 5.3.5 Memory Management

All string values returned by GoWSK functions are allocated in C memory and must be freed using `GoWSK_FreeCString` to prevent memory leaks. Error messages passed through `errOut` parameters must also be freed if they are not NULL.

**Important:** Always check return values for NULL before using them, and always free returned strings and error messages to prevent memory leaks.

## 6. Limitations & Future Improvements

-
