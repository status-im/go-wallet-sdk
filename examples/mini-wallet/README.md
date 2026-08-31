# Mini Wallet Example

A full-featured web-based cryptocurrency wallet demonstrating the capabilities of the Go Wallet SDK. This example showcases a complete wallet implementation with account management, multi-chain balance fetching, token transfers, and transaction preview.

## What it demonstrates

- 🔐 **Account Management**: Create and manage Ethereum accounts with keystore encryption
- 🌐 **Multi-Chain Support**: Support for multiple EVM-compatible chains (Ethereum, Arbitrum, Optimism, and their testnets)
- 💰 **Balance Fetching**: Efficiently fetch native and ERC20 token balances using Multicall3
- 📊 **Token Lists**: Fetch and combine token lists from multiple sources (Status, Uniswap, CoinGecko, Aave)
- 🔍 **Token Exploration**: Browse combined token lists with statistics and chain-specific filtering
- 📤 **Token Transfers**: Send native tokens and ERC20 tokens with transaction preview
- 🌍 **ENS Resolution**: Resolve ENS names to addresses (always uses Ethereum mainnet)
- 👁️ **Transaction Preview**: Preview transaction details including gas price (in Gwei), estimated gas, and total cost before signing
- 🔗 **Chain Management**: Enable/disable chains and switch between accounts via sidebar
- 📝 **RPC Logging**: Comprehensive logging of all RPC calls for debugging

## Run

1. Navigate to the example directory:
   ```bash
   cd examples/mini-wallet
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the server:
   ```bash
   go run .
   ```

4. Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

## Features

### Account Management

- **Create Account**: Generate a new Ethereum account with a password-protected keystore
- **Select Account**: Switch between multiple accounts without requiring a password
- **Account List**: View all accounts in the keystore in the sidebar

### Multi-Chain Balance Fetching

- **Native Tokens**: View native token balances (ETH, ARB, OP, etc.) across all enabled chains
- **ERC20 Tokens**: Automatically fetch ERC20 token balances using Multicall3 for efficiency
- **Token Icons**: Display token icons from token lists
- **Formatted Balances**: Show balances in human-readable format with proper decimal places
- **Token Addresses**: Clickable token addresses that copy to clipboard

### Token Lists

- **Multiple Sources**: Fetches token lists from:
  - Status Market
  - Uniswap
  - CoinGecko (Ethereum, Optimism, Arbitrum)
  - Aave
- **Token Builder**: Combines all token lists into a unified, deduplicated collection
- **Native Tokens**: Automatically includes native token information for all chains
- **Token Exploration**: Browse token statistics and filter by chain

### Token Transfers

- **Native Transfers**: Send native tokens (ETH, ARB, OP, etc.)
- **ERC20 Transfers**: Send ERC20 tokens with automatic gas estimation
- **ENS Support**: Send to ENS names (e.g., `vitalik.eth`) with automatic resolution
- **Transaction Preview**: Review transaction details before signing:
  - From/To addresses
  - Amount and token symbol
  - Chain information
  - Gas limit and gas price (in Gwei)
  - Total cost (in Gwei)
  - ENS resolution details (if applicable)
- **Password Protection**: Password required only when signing transactions

### Chain Management

- **Enable/Disable Chains**: Toggle chains on/off via sidebar switches
- **Chain Selection**: Only enabled chains appear in transaction forms
- **Default State**: Ethereum Mainnet is disabled by default (can be enabled via sidebar)

### Supported Chains

- **Ethereum Mainnet** (Chain ID: 1) - Disabled by default
- **Arbitrum One** (Chain ID: 42161)
- **Optimism** (Chain ID: 10)
- **Ethereum Sepolia** (Chain ID: 11155111)
- **Arbitrum Sepolia** (Chain ID: 421614)
- **Optimism Sepolia** (Chain ID: 11155420)
- **Status Sepolia** (Chain ID: 1660990954)

## Usage

### Initial Setup

1. **Initialize Keystore**: Click "Initialize Keystore" to create a new keystore directory
2. **Create Account**: Create your first account with a password
3. **Select Account**: Click on an account in the sidebar to view its balances

### Viewing Balances

1. Select an account from the sidebar
2. Click "Refresh All Balances" to fetch balances across all enabled chains
3. Balances are displayed sorted by chain name, then token symbol
4. Click on a token address to copy it to clipboard

### Sending Tokens

1. **Select Account**: Choose the account to send from
2. **Choose Chain**: Select the chain from the dropdown (only enabled chains shown)
3. **Enter Recipient**: Enter an Ethereum address or ENS name (e.g., `vitalik.eth`)
4. **Select Token**: Choose native token or an ERC20 token
5. **Enter Amount**: Enter the amount in decimal format (e.g., `1.5`)
6. **Preview Transaction**: Click "Preview Transaction" to review details
7. **Confirm & Send**: Enter your password and click "Confirm & Send"

### Exploring Token Lists

1. Click "🔍 Explore Token List" in the sidebar
2. View general statistics about the combined token list
3. Select a chain to see all tokens for that chain
4. View token details including symbol, name, decimals, and logo

### Managing Chains

1. Use the toggle switches in the sidebar to enable/disable chains
2. Disabled chains won't appear in balance fetching or transaction forms
3. Changes take effect immediately

## Project Structure

```
mini-wallet/
├── main.go              # Application entry point and HTTP handlers
├── templates/
│   └── index.html       # Frontend HTML, CSS, and JavaScript
├── go.mod               # Go module dependencies
├── go.sum               # Dependency checksums
└── README.md            # This file
```

## API Endpoints

- `GET /` - Web interface
- `POST /init-keystore` - Initialize keystore
- `POST /create-account` - Create a new account
- `POST /unlock` - Select an account (no password required)
- `GET /accounts` - Get all accounts
- `GET /chains` - Get all chains with enabled status
- `POST /chains/toggle` - Toggle chain enabled status
- `GET /balances?address=<address>` - Get balances for an address
- `POST /transfer/preview` - Preview transaction details
- `POST /transfer` - Send a transaction
- `GET /tokens/stats` - Get token list statistics
- `GET /tokens/chain?chainId=<id>` - Get tokens for a specific chain

## Key SDK Packages Used

- **`pkg/ethclient`**: Multi-chain RPC client with logging wrapper
- **`pkg/balance/multistandardfetcher`**: Efficient balance fetching using Multicall3
- **`pkg/contracts/multicall3`**: Multicall3 contract interface
- **`pkg/tokens/builder`**: Combine token lists from multiple sources
- **`pkg/tokens/fetcher`**: Fetch token lists from remote URLs
- **`pkg/tokens/parsers`**: Parse different token list formats
- **`pkg/txgenerator`**: Generate unsigned transactions
- **`pkg/ens`**: ENS name resolution
- **`pkg/accounts/keystore`**: Account management and signing

## RPC Logging

All RPC calls are logged to the console with:
- Chain name and ID
- RPC method name
- Number of arguments
- Execution time
- Success/error status

Example log output:
```
[RPC] Ethereum (1) -> eth_chainId with 0 args
[RPC] Ethereum (1) -> eth_chainId SUCCESS after 125ms
[RPC] Optimism (10) -> eth_call to=0xcA11bde05977b3631167028862bE2a173976CA11 block=latest
[RPC] Optimism (10) -> eth_call SUCCESS after 45ms (result length: 64)
```

