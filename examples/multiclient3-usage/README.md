# Multicall3 Example: Batch ERC20 Balance Queries

Fetch hundreds of ERC20 token balances in a single RPC call using Multicall3.

## Run

```bash
# Set your RPC endpoint
export RPC_URL="https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY"

# Run the example
go run main.go
```

## What it demonstrates

- Loads 1000+ tokens from CoinGecko's `all.json`
- Queries all ERC20 balances + ETH balance + block number in one call
- Shows non-zero balances in human-readable format
- Uses vitalik.eth's address by default (easily changeable)

## Key Benefits

- **1 RPC call** instead of 1000+ individual calls
- **Faster execution** with parallel processing
- **Lower costs** and reduced rate limiting
- **Better reliability** with atomic execution

## Environment Variables

- `RPC_URL`: Your Ethereum RPC endpoint (required)
  - Alchemy: `https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY`
  - Infura: `https://mainnet.infura.io/v3/YOUR_KEY`
  - QuickNode: `https://your-node.quiknode.pro/YOUR_KEY/`

## Output Example

```
=== Multicall3 Results ===
Block Number: 19500000
✅ ETH (Ethereum): 0.000000000000000001

=== ERC20 Token Balances for 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 ===

✅ USDC (USD Coin): 150.50
✅ WETH (Wrapped Ether): 2.5
✅ DAI (Dai Stablecoin): 1000.00

=== Summary ===
Total tokens queried: 1000+
Successful calls: 950+
Tokens with non-zero balance: 3
Success rate: 95.00%
```

## Supported Chains

Automatically detects Multicall3 deployment on:
- Ethereum, Polygon, Arbitrum, Optimism, BSC, and 100+ more
