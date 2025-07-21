# Gas Comparison Tool

A multi-network gas fee comparison tool that compares our gas estimation implementation against legacy estimators and Infura's Gas API.

## What It Does

- **Compares implementations**: Our new `GetTxSuggestions` vs old estimator vs Infura API
- **Multi-network support**: Ethereum, Arbitrum, Optimism, Base, Polygon, Linea, BSC, Status Network
- **Comprehensive analysis**: Priority fees, max fees, base fees, wait times, network congestion
- **Real vs local data**: Test with live networks or use local mock data

## Quick Start

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
   Wait Time (Current):    72.0--12.0 seconds
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

- **Ethereum Mainnet** (L1)
- **Arbitrum One** (ArbStack)
- **Optimism** (OPStack)
- **Base** (OPStack)
- **Polygon** (L1)
- **Linea** (LineaStack)
- **BNB Smart Chain** (L1)
- **Status Network Sepolia** (LineaStack)

## Use Cases

- **Development**: Test gas estimation accuracy across networks
- **Comparison**: Evaluate different fee strategies
- **Monitoring**: Track gas fee trends and network conditions
- **Validation**: Ensure our implementation matches industry standards
