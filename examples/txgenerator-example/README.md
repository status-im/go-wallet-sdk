# Transaction Generator Web Example

A web-based interface for generating unsigned Ethereum transactions using the `pkg/txgenerator` package. Supports multiple transaction types including ETH transfers, ERC20, ERC721, and ERC1155 operations.

## What it demonstrates

- 🌐 Web interface for easy transaction generation
- 🔷 Support for multiple transaction types (ETH, ERC20, ERC721, ERC1155)
- ⛽ Support for both Legacy and EIP-1559 fee types
- 📝 Dynamic form fields based on selected transaction type
- 📄 Returns transactions in JSON format
- ✅ Comprehensive parameter validation
- 🔒 Unsigned transactions ready for signing

## Run

### Prerequisites

- Go 1.23.0 or later
- No external dependencies required (no RPC endpoints needed)

### Running the Example

```bash
cd examples/txgenerator-example
go mod tidy
go run .
```

Access: http://localhost:8080

The server will start and log the access URL. You can also build and run the binary:

```bash
go build -o txgenerator-example
./txgenerator-example
```

## Supported Transaction Types

### Native ETH
- **Transfer ETH**: Simple native token transfer

### ERC20 Tokens
- **Transfer ERC20**: Transfer ERC20 tokens
- **Approve ERC20**: Approve ERC20 token spending

### ERC721 Tokens (NFTs)
- **Transfer ERC721 (transferFrom)**: Basic NFT transfer
- **Transfer ERC721 (safeTransferFrom)**: Safe NFT transfer with recipient check
- **Approve ERC721**: Approve a specific NFT
- **Set Approval For All ERC721**: Approve/revoke operator for all NFTs

### ERC1155 Tokens
- **Transfer ERC1155**: Single token transfer
- **Batch Transfer ERC1155**: Batch transfer multiple tokens
- **Set Approval For All ERC1155**: Approve/revoke operator for all tokens

## Using the UI

1. **Select Transaction Type**: Choose from the dropdown menu
2. **Choose Fee Type**: 
   - **Legacy**: Uses GasPrice (for older networks)
   - **EIP-1559**: Uses MaxFeePerGas and MaxPriorityFeePerGas (recommended for modern networks)
3. **Fill Common Parameters**:
   - Nonce: Transaction nonce
   - Gas Limit: Maximum gas to use
   - Chain ID: Network chain ID (1 for Ethereum mainnet)
   - Gas Price (Legacy) or MaxFeePerGas + MaxPriorityFeePerGas (EIP-1559)
4. **Fill Transaction-Specific Parameters**: Fields will appear based on selected transaction type
5. **Generate Transaction**: Click "Generate Transaction" to get the JSON output

## Example Parameters

### Transfer ETH
```json
{
  "txType": "transferETH",
  "useEIP1559": false,
  "nonce": "0",
  "gasLimit": "21000",
  "chainID": "1",
  "gasPrice": "20000000000",
  "params": {
    "to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "value": "1000000000000000000"
  }
}
```
- To: `0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb`
- Value: `1000000000000000000` (1 ETH in wei)

### Transfer ERC20
```json
{
  "txType": "transferERC20",
  "useEIP1559": true,
  "nonce": "0",
  "gasLimit": "65000",
  "chainID": "1",
  "maxFeePerGas": "30000000000",
  "maxPriorityFeePerGas": "2000000000",
  "params": {
    "tokenAddress": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
    "to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "amount": "1000000"
  }
}
```
- Token Address: `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48` (USDC)
- To: `0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb`
- Amount: `1000000` (1 USDC with 6 decimals)

### Transfer ERC721
```json
{
  "txType": "safeTransferFromERC721",
  "useEIP1559": true,
  "nonce": "0",
  "gasLimit": "100000",
  "chainID": "1",
  "maxFeePerGas": "30000000000",
  "maxPriorityFeePerGas": "2000000000",
  "params": {
    "tokenAddress": "0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D",
    "from": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "to": "0x8ba1f109551bd432803012645ac136ddd64dba72",
    "tokenID": "1234"
  }
}
```
- Token Address: `0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D` (Bored Ape Yacht Club)
- From: `0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb`
- To: `0x8ba1f109551bd432803012645ac136ddd64dba72`
- Token ID: `1234`

### Batch Transfer ERC1155
```json
{
  "txType": "batchTransferERC1155",
  "useEIP1559": true,
  "nonce": "0",
  "gasLimit": "200000",
  "chainID": "1",
  "maxFeePerGas": "30000000000",
  "maxPriorityFeePerGas": "2000000000",
  "params": {
    "tokenAddress": "0x...",
    "from": "0x...",
    "to": "0x...",
    "tokenIDs": "1,2,3",
    "values": "10,20,30"
  }
}
```
- Token Address: `0x...`
- From: `0x...`
- To: `0x...`
- Token IDs: `1,2,3` (comma-separated)
- Values: `10,20,30` (comma-separated, corresponding to token IDs)

## Response Format

The generated transaction is returned in JSON format with the following fields:

```json
{
  "type": "0x0" or "0x2",
  "nonce": "0",
  "gasPrice": "...",
  "maxFeePerGas": "...",
  "maxPriorityFeePerGas": "...",
  "gasLimit": "21000",
  "to": "0x...",
  "value": "...",
  "data": "0x...",
  "chainID": "1",
  "v": "0x...",
  "r": "0x...",
  "s": "0x...",
  "hash": "0x...",
  "raw": "0x..."
}
```

## Code Structure

The example is organized into several files:

- `main.go` - Application entry point, server setup, and HTTP routing
- `handlers.go` - Transaction generation logic and request handling
- `templates.go` - HTML templates and frontend JavaScript for the web interface
- `go.mod` - Go module definition and dependencies

### Key Functions

- `handleHome()` - Serves the web interface HTML
- `handleGenerateTransaction()` - Handles POST requests to generate transactions
- `GenerateTransaction()` - Core transaction generation logic based on request type
- `TransactionToJSON()` - Converts transaction to JSON format for response

## API

- `GET /` - Web interface
- `POST /generate` - Generate transaction

### Request Format

```json
{
  "txType": "transferETH",
  "useEIP1559": false,
  "nonce": "0",
  "gasLimit": "21000",
  "chainID": "1",
  "gasPrice": "20000000000",
  "maxFeePerGas": "",
  "maxPriorityFeePerGas": "",
  "params": {
    "to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "value": "1000000000000000000"
  }
}
```

## Fee Types

### Legacy Transactions (Type 0)
- Uses `GasPrice` field
- Compatible with older networks
- Example: `gasPrice: "20000000000"` (20 gwei)

### EIP-1559 Transactions (Type 2)
- Uses `MaxFeePerGas` and `MaxPriorityFeePerGas` fields
- Recommended for modern networks (Ethereum after London fork)
- Example: 
  - `maxFeePerGas: "30000000000"` (30 gwei)
  - `maxPriorityFeePerGas: "2000000000"` (2 gwei)

## Dependencies

- `github.com/ethereum/go-ethereum` - Core Ethereum types and transaction structures
- `github.com/status-im/go-wallet-sdk/pkg/txgenerator` - Transaction generation package
- `github.com/gorilla/mux` - HTTP router and URL matcher

## Error Handling

The example includes comprehensive error handling for:
- Invalid transaction parameters (zero addresses, negative amounts)
- Missing required fields
- Invalid numeric values
- Unsupported transaction types
- JSON encoding/decoding errors

All errors are returned as JSON responses with descriptive error messages.

## Troubleshooting

- **"undefined: handleHome" error**: Use `go run .` instead of `go run main.go`
- **Invalid Addresses**: Ensure addresses are valid Ethereum addresses (0x-prefixed, 40 hex characters)
- **Missing Parameters**: All required fields must be filled based on the selected transaction type
- **Invalid Amounts**: Ensure numeric values are valid (no decimals for wei/token units)
- **Port Already in Use**: Change the port in `main.go` if 8080 is already in use
- **Transaction Generation Fails**: Check that all required parameters are provided and valid

