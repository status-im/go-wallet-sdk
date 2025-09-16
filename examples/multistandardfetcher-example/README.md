# Multi-Standard Fetcher Example

This example demonstrates how to use the `multistandardfetcher` package to fetch balances across all token standards (Native ETH, ERC20, ERC721, ERC1155) for a specific address using Multicall3 batched calls.

## What It Does

- Fetches native ETH balance for vitalik.eth
- Queries ERC20 token balances for popular tokens (USDC, DAI, USDT, WBTC, LINK, UNI, MATIC, SHIB)
- Checks ERC721 NFT balances for well-known collections (BAYC, MAYC, CryptoPunks, Azuki, Moonbirds, POAP)
- Retrieves ERC1155 collectible balances from popular contracts (OpenSea, Rarible)
- Displays results in a formatted report with token symbols and readable balances

## Quick Start

```bash
# Set your RPC endpoint
export RPC_URL="https://eth-mainnet.g.alchemy.com/v2/YOUR_API_KEY"

# Run the example
go run main.go
```

## Features

- **Multi-Standard Support**: Queries all four token standards in a single operation
- **Popular Tokens**: Includes well-known ERC20 tokens, NFT collections, and ERC1155 collectibles
- **Formatted Output**: Clean, readable report with token symbols and proper formatting
- **Error Handling**: Graceful handling of failed calls and network errors
- **Efficient Batching**: Uses Multicall3 to minimize RPC calls

## Environment Variables

- `RPC_URL`: Your Ethereum RPC endpoint (required)
  - Alchemy: `https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY`
  - Infura: `https://mainnet.infura.io/v3/YOUR_KEY`
  - QuickNode: `https://your-node.quiknode.pro/YOUR_KEY/`

## Output Example

```
Using Multicall3 contract at: 0xcA11bde05977b3631167028862bE2a173976CA11
Fetching balances for 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045 across all token standards...
Checking 8 ERC20 tokens, 6 ERC721 NFTs, and 4 ERC1155 collectibles
✅ Native ETH balance: 1000000000000000000 wei (block 19543210)
✅ ERC20 balances fetched (block 19543210)
✅ ERC721 balances fetched (block 19543210)
✅ ERC1155 balances fetched (block 19543210)

================================================================================
BALANCE REPORT FOR 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
================================================================================

💰 NATIVE ETH BALANCE
----------------------------------------
ETH: 1.000000 ETH (1000000000000000000 wei)

🪙 ERC20 TOKEN BALANCES
----------------------------------------
USDC: 1500000000
DAI: 1000000000000000000000
Found 2 tokens with non-zero balances

🖼️  ERC721 NFT BALANCES
----------------------------------------
BAYC: 1 NFTs
POAP: 3 NFTs
Found 2 NFT collections with non-zero balances

🎨 ERC1155 COLLECTIBLE BALANCES
----------------------------------------
0x495f9472... (token 1): 5
0xd07dc426... (token 24775): 2
Found 2 collectibles with non-zero balances

📊 SUMMARY
----------------------------------------
Total result sets processed: 4
Native ETH: 1000000000000000000
ERC20 tokens with balance: 2
ERC721 NFT collections with balance: 2
ERC1155 collectibles with balance: 2
```

## Supported Tokens

### ERC20 Tokens
- USDC (USD Coin)
- DAI (Dai Stablecoin)
- USDT (Tether USD)
- WBTC (Wrapped Bitcoin)
- LINK (Chainlink)
- UNI (Uniswap)
- MATIC (Polygon)
- SHIB (Shiba Inu)

### ERC721 NFT Collections
- BAYC (Bored Ape Yacht Club)
- MAYC (Mutant Ape Yacht Club)
- CryptoPunks
- Azuki
- Moonbirds
- POAP

### ERC1155 Collectibles
- OpenSea Shared Storefront tokens
- Rarible collectibles

## Customization

You can easily modify the example to:

1. **Change the target address**: Update the `walletAddress` variable
2. **Add more tokens**: Extend the `erc20Tokens`, `erc721NFTs`, or `erc1155Collectibles` slices
3. **Query different chains**: Change the `chainID` variable
4. **Adjust batch size**: Modify the batch size parameter in `FetchBalances`

## Code Structure

- `main.go`: Complete example implementation
- `go.mod`: Go module configuration
- `README.md`: This documentation

## Dependencies

- `github.com/status-im/go-wallet-sdk/pkg/balance/multistandardfetcher`
- `github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3`
- `github.com/ethereum/go-ethereum`

## Performance

- **Batch Size**: Uses 100 calls per batch (adjustable)
- **Concurrent Processing**: Results are processed asynchronously
- **RPC Efficiency**: Minimizes API calls through Multicall3 batching
- **Memory Usage**: Streams results through channels to minimize memory usage
