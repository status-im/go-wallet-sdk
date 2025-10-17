package data

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

type FakeClient struct {
	gasData *GasData
}

func NewFakeClient(gasData *GasData) *FakeClient {
	return &FakeClient{
		gasData: gasData,
	}
}

func (c *FakeClient) Close() {
	// No-op
}

func (c *FakeClient) ChainID(ctx context.Context) (*big.Int, error) {
	if c.gasData == nil || c.gasData.LatestBlock == nil {
		return nil, fmt.Errorf("no gas data available")
	}
	for _, tx := range c.gasData.LatestBlock.Transactions {
		if tx.ChainID != nil {
			return big.NewInt(tx.ChainID.Int64()), nil
		}
	}
	return nil, fmt.Errorf("no chain ID found in transactions")
}

func (c *FakeClient) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
	// Validate gas data availability
	if c.gasData == nil || c.gasData.FeeHistory == nil {
		return nil, fmt.Errorf("no fee history data available")
	}

	// Validate blockCount
	if blockCount == 0 {
		return nil, fmt.Errorf("blockCount must be greater than 0")
	}

	availableBlocks := uint64(len(c.gasData.FeeHistory.GasUsedRatio))
	if blockCount > availableBlocks {
		return nil, fmt.Errorf("blockCount %d exceeds available blocks %d", blockCount, availableBlocks)
	}

	// Validate rewardPercentiles
	for _, percentile := range rewardPercentiles {
		if percentile < 0 || percentile > 100 {
			return nil, fmt.Errorf("rewardPercentiles must be between 0 and 100")
		}
	}

	if lastBlock != nil && lastBlock.Cmp(c.gasData.LatestBlock.Number) != 0 {
		return nil, fmt.Errorf("requested lastBlock %s is not available", lastBlock.String())
	}

	// Calculate the starting index for the requested blocks
	startIdx := uint64(0)
	effectiveBlockCount := availableBlocks
	if blockCount <= availableBlocks {
		startIdx = availableBlocks - blockCount
		effectiveBlockCount = blockCount
	}

	// Calculate oldest block number
	latestBlockNumber := c.gasData.LatestBlock.Number.Uint64()
	oldestBlockNumber := latestBlockNumber - effectiveBlockCount + 1

	ret := &ethereum.FeeHistory{
		OldestBlock:  big.NewInt(int64(oldestBlockNumber)),
		BaseFee:      make([]*big.Int, effectiveBlockCount+1), // +1 for next block base fee
		GasUsedRatio: make([]float64, effectiveBlockCount),
		Reward:       make([][]*big.Int, effectiveBlockCount),
	}

	// Copy base fees (including the extra one for next block)
	for i := uint64(0); i <= effectiveBlockCount; i++ {
		// BaseFeePerGas includes estimation for the next block, so we must start one block before
		if startIdx-1+i < uint64(len(c.gasData.FeeHistory.BaseFee)) {
			ret.BaseFee[i] = big.NewInt(0).Set(c.gasData.FeeHistory.BaseFee[startIdx-1+i])
		} else {
			ret.BaseFee[i] = big.NewInt(0)
		}
	}

	// Copy gas used ratios
	for i := uint64(0); i < effectiveBlockCount; i++ {
		if startIdx+i < uint64(len(c.gasData.FeeHistory.GasUsedRatio)) {
			ret.GasUsedRatio[i] = c.gasData.FeeHistory.GasUsedRatio[startIdx+i]
		} else {
			ret.GasUsedRatio[i] = 0.0
		}
	}

	// Filter rewards based on requested percentiles
	for i := uint64(0); i < effectiveBlockCount; i++ {
		if startIdx+i < uint64(len(c.gasData.FeeHistory.Reward)) {
			sourceRewards := c.gasData.FeeHistory.Reward[startIdx+i]
			ret.Reward[i] = make([]*big.Int, len(rewardPercentiles))

			// Map requested percentiles to available reward data
			for j, requestedPercentile := range rewardPercentiles {
				// Find the closest available percentile index
				// Assuming the source data has percentiles at regular intervals
				percentileIdx := int(requestedPercentile / 5) // Assuming 5% intervals (0, 5, 10, ..., 100)
				ret.Reward[i][j] = big.NewInt(0)
				if percentileIdx < len(sourceRewards) {
					ret.Reward[i][j].Set(sourceRewards[percentileIdx])
				}
			}
		} else {
			// Create empty reward data for missing blocks
			ret.Reward[i] = make([]*big.Int, len(rewardPercentiles))
			for j := range rewardPercentiles {
				ret.Reward[i][j] = big.NewInt(0)
			}
		}
	}

	return ret, nil
}

func (c *FakeClient) BlockNumber(ctx context.Context) (uint64, error) {
	if c.gasData == nil || c.gasData.LatestBlock == nil {
		return 0, fmt.Errorf("no gas data available")
	}
	return c.gasData.LatestBlock.Number.Uint64(), nil
}

func (c *FakeClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethclient.BlockWithFullTxs, error) {
	if c.gasData == nil || c.gasData.LatestBlock == nil {
		return nil, fmt.Errorf("no gas data available")
	}
	if number == nil {
		// Return latest block
		return c.gasData.LatestBlock, nil
	}
	// For simplicity, always return the latest block regardless of the requested number
	return c.gasData.LatestBlock, nil
}

func (c *FakeClient) GetGasSuggestions(ctx context.Context, networkID int) (*infura.GasResponse, error) {
	dataChainID, err := c.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	if networkID != int(dataChainID.Int64()) {
		return nil, fmt.Errorf("networkID %d does not match latest block networkID %d", networkID, dataChainID.Int64())
	}
	return c.gasData.InfuraSuggestedFees, nil
}

func (c *FakeClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return c.gasData.GasPrice, nil
}

func (c *FakeClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return c.gasData.MaxPriorityFeePerGas, nil
}

func (c *FakeClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	// Return a default gas limit for simple ETH transfers
	if msg.To != nil && len(msg.Data) == 0 && (msg.Value == nil || msg.Value.Cmp(big.NewInt(0)) == 0) {
		return 21000, nil // Standard ETH transfer gas limit
	}
	// For other transactions, return a reasonable default
	return 100000, nil
}

func (c *FakeClient) LineaEstimateGas(ctx context.Context, msg ethereum.CallMsg) (*ethclient.LineaEstimateGasResult, error) {
	return nil, fmt.Errorf("LineaEstimateGas not implemented")
}
