# Balance Fetcher Web Example

A web-based interface for fetching native token balances across multiple Ethereum chains using the `pkg/balance/fetcher` package.

## Run

```bash
cd examples/balance-fetcher-web
go run .
```

Access: http://localhost:8080

## What it demonstrates

- 🌐 Web interface for easy interaction
- 🔗 Support for any EVM-compatible chain
- 📊 Batch balance fetching for multiple addresses
- 📦 Automatic fallback between Multicall3 and standard RPC calls
- 💰 Display balances in both ETH and Wei
- ⚡ Prepopulated with popular chains (Ethereum, Optimism, Arbitrum, Polygon)

## Using the UI

1. **Configure Chains**: Add custom chains with ChainID and RPC URL
2. **Enter Addresses**: Add Ethereum addresses (one per line)
3. **Block Number** (optional): Specify block number or leave empty for latest
4. **Fetch Balances**: Click "Fetch Balances"

### Example Addresses

```
0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6
0x1234567890123456789012345678901234567890
```

### Supported Chains

Any EVM-compatible chain. Popular chains are prepopulated:
- **Ethereum Mainnet** (Chain ID: 1)
- **Optimism Mainnet** (Chain ID: 10)
- **Arbitrum One** (Chain ID: 42161)
- **Polygon** (Chain ID: 137)

## Project Structure

```
balance-fetcher-web/
├── main.go          # Application entry point
├── types.go         # Data structures
├── rpc_client.go    # Custom RPC client implementation
├── utils.go         # Utility functions
├── templates.go     # HTML templates and frontend JavaScript
├── handlers.go      # HTTP request handlers
└── README.md        # This file
```

## API

- `GET /` - Web interface
- `POST /fetch` - Fetch balances

### Request Format

```json
{
  "chains": [
    {
      "chainId": 1,
      "rpcUrl": "https://ethereum-rpc.publicnode.com"
    }
  ],
  "addresses": [
    "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"
  ],
  "blockNum": "18000000"
}
```

## Troubleshooting

- **"undefined: handleHome" error**: Use `go run .` instead of `go run main.go`
- **Connection Errors**: Check RPC endpoints are accessible
- **Invalid Addresses**: Ensure addresses are valid Ethereum addresses (0x-prefixed, 40 hex characters)

## Security Notes

This is an example application. For production use:
- Secure RPC endpoints
- Implement authentication
- Validate user inputs
- Consider rate limits 