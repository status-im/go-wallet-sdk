# Gas Comparison Tool

A multi-network gas fee comparison tool that compares our gas estimation implementation against legacy estimators and Infura's Gas API.

## What it demonstrates

- **Compares implementations**: Our new `GetTxSuggestions` vs old estimator vs Infura API
- **Multi-network support**: Ethereum, Arbitrum, Optimism, Base, Polygon, Linea, BSC, Status Network
- **Comprehensive analysis**: Priority fees, max fees, base fees, wait times, network congestion
- **Real vs local data**: Test with live networks or use local mock data

## Run

```bash
# Test with local mock data
./gas-comparison -fake

# Test with real networks (requires Infura API key)
./gas-comparison -infura-api-key YOUR_API_KEY
```

## What You'll See

```
🔸 LOW PRIORITY FEES
   Current Implementation:               200000 wei
   Old Implementation:                   330120 wei
   Infura:                              1000000 wei
   Current vs Old:                      -130120 wei (-39.4%)
   Current vs Infura:                   -800000 wei (-80.0%)

🔸 LOW WAIT TIME
   Wait Time (Current):    12.0-72.0 seconds
   Wait Time (Old):        125 seconds
   Wait Time (Infura):     12.0-48.0 seconds
```

## Test Transaction

Uses a simple 0-valued ETH transfer from Vitalik's address:
- **From**: `0xd8da6bf26964af9d7eed9e03e53415d37aa96045`
- **To**: Zero address
- **Value**: 0 ETH
- **Data**: Empty

## Networks Tested

- **Ethereum Mainnet** (L1) - 12s block time
- **Arbitrum One** (ArbStack) - 0.25s block time
- **Optimism** (OPStack) - 2s block time
- **Optimism Sepolia** (OPStack) - 2s block time
- **Base** (OPStack) - 2s block time
- **Polygon** (L1) - 2.25s block time
- **Linea** (LineaStack) - 2s block time
- **BNB Smart Chain** (L1) - 0.75s block time
- **Status Network Sepolia** (LineaStack) - 2s block time

## Use Cases

- **Development**: Test gas estimation accuracy across networks
- **Comparison**: Evaluate different fee strategies (current vs legacy vs Infura)
- **Monitoring**: Track gas fee trends and network conditions
- **Validation**: Ensure our implementation matches industry standards

## Architecture

The tool uses:
- **Chain-specific parameters**: Each network has optimized block counts and configurations
- **DefaultConfig**: Automatically configures estimation based on chain class
- **Separate inclusions**: Time estimates are now in dedicated inclusion structs
- **Mock data support**: Test locally with pre-captured network data

## Data Generator

Generate fresh test data for any network:

```bash
cd data/generator
go run main.go -rpc https://mainnet.infura.io/v3/YOUR_API_KEY
```

This captures:
- Latest block with full transactions
- Fee history (1024 blocks)
- Current gas prices
- Infura fee suggestions
