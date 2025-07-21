# Gas Package

A comprehensive gas estimation and fee suggestion package for Ethereum and L2 networks.

## Features

- **Multi-chain support**: Ethereum L1, Arbitrum Stack, Optimism Stack, Linea Stack
- **Smart fee estimation**: Priority fees, base fees, and max fees with inclusion time estimates

## Quick Start

```go
import "github.com/status-im/go-wallet-sdk/pkg/gas"

// Get fee suggestions for a transaction
config := gas.DefaultConfig()
config.ChainClass = gas.ChainClassL1

callMsg := &ethereum.CallMsg{
    To:   &common.Address{},
    Data: []byte{},
    Value: big.NewInt(0),
}

suggestions, err := gas.GetTxSuggestions(ctx, ethClient, config, callMsg)
if err != nil {
    return err
}

// Access fee suggestions
lowFee := suggestions.FeeSuggestions.Low.MaxFeePerGas
lowTime := suggestions.FeeSuggestions.Low.MinTimeUntilInclusion
```

## Chain Classes

- **L1**: Ethereum mainnet, Polygon, BSC
- **ArbStack**: Arbitrum One, Arbitrum Nova
- **OPStack**: Optimism, Base, OP Sepolia
- **LineaStack**: Linea mainnet, Linea testnet

## Configuration

```go
config := gas.SuggestionsConfig{
    ChainClass:               gas.ChainClassL1,
    NetworkBlockTime:         12.0,        // seconds
    GasPriceEstimationBlocks: 10,          // blocks
    BaseFeeMultiplier:        1.025,       // 2.5% buffer
    LowRewardPercentile:      10,          // %
    MediumRewardPercentile:   50,          // %
    HighRewardPercentile:     90,          // %
}
```

## Output

Returns `TxSuggestions` with:
- **FeeSuggestions**: Low/Medium/High priority and max fees
- **GasLimit**: Estimated gas limit for the transaction
- **Time estimates**: Min/Max time until inclusion (in seconds)
- **Network congestion**: 0-1 scale for L1 chains
