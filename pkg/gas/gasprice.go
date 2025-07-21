package gas

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

// Estimator provides gas estimation functionality
type Estimator struct {
	ethClient EthClient
	config    EstimatorConfig
}

// NewEstimator creates a new gas estimator
func NewEstimator(ethClient EthClient) (*Estimator, error) {
	config := DefaultConfig()

	return &Estimator{
		ethClient: ethClient,
		config:    config,
	}, nil
}

// NewEstimatorWithConfig creates a new gas estimator with custom config
func NewEstimatorWithConfig(ethClient EthClient, config EstimatorConfig) (*Estimator, error) {
	return &Estimator{
		ethClient: ethClient,
		config:    config,
	}, nil
}

// DefaultConfig returns default configuration
func DefaultConfig() EstimatorConfig { // Defaults for Ethereum Mainnet
	return EstimatorConfig{
		LegacyBlocks:     5,  // 5 for Ethereum, 25 for L2s
		PercentileBlocks: 10, // 10 for Ethereum, 50 for L2s
		LowPercentile:    10,
		//MediumPercentile:  50,
		MediumPercentile:  45,
		HighPercentile:    90,
		BaseFeeMultiplier: 1.025, // 2.5% buffer for base fee
	}
}

// GetFeeSuggestions returns gas fee suggestions compatible with Infura's API
func (e *Estimator) GetFeeSuggestions(ctx context.Context) (*FeeSuggestions, error) {
	// Get latest block to check if EIP-1559 is supported
	latestBlock, err := e.getLatestBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	var suggestions *FeeSuggestions

	if latestBlock.BaseFeePerGas != nil {
		// EIP-1559 network - try eth_feeHistory first
		suggestions, err = e.getEIP1559Suggestions(ctx, latestBlock)
	} else {
		// Legacy network
		suggestions, err = e.getLegacySuggestions(ctx)
	}

	if err != nil {
		return nil, err
	}

	return suggestions, nil
}

// getEIP1559SuggestionsLegacy is the fallback method when eth_feeHistory is not available
func (e *Estimator) getEIP1559SuggestionsLegacy(ctx context.Context, latestBlock *ethclient.BlockWithFullTxs) (*FeeSuggestions, error) {
	// Get historical blocks for analysis
	blocks, err := e.getHistoricalBlocks(ctx, latestBlock.Number, e.config.LegacyBlocks)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical blocks: %w", err)
	}

	// Calculate priority fee percentiles
	priorityFees := e.extractPriorityFees(blocks)
	if len(priorityFees) == 0 {
		return nil, fmt.Errorf("no priority fee data available")
	}

	sort.Slice(priorityFees, func(i, j int) bool {
		return priorityFees[i].Cmp(priorityFees[j]) < 0
	})

	lowPriorityFee := e.getPercentile(priorityFees, e.config.LowPercentile)
	mediumPriorityFee := e.getPercentile(priorityFees, e.config.MediumPercentile)
	highPriorityFee := e.getPercentile(priorityFees, e.config.HighPercentile)

	// Calculate next block's estimated base fee
	estimatedBaseFee := e.calculateNextBaseFee(latestBlock)

	// Apply multiplier for buffer
	bufferedBaseFee := new(big.Int).Mul(estimatedBaseFee, big.NewInt(int64(e.config.BaseFeeMultiplier*1000)))
	bufferedBaseFee.Div(bufferedBaseFee, big.NewInt(1000))

	// Calculate max fee per gas (base fee + priority fee)
	lowMaxFee := new(big.Int).Add(bufferedBaseFee, lowPriorityFee)
	mediumMaxFee := new(big.Int).Add(bufferedBaseFee, mediumPriorityFee)
	highMaxFee := new(big.Int).Add(bufferedBaseFee, highPriorityFee)

	// Calculate network congestion
	congestion := e.calculateNetworkCongestion(blocks)

	// Get accurate time estimates for each priority level
	lowMinTime, lowMaxTime := e.getTimeEstimatesForFee(ctx, lowPriorityFee, lowMaxFee)
	mediumMinTime, mediumMaxTime := e.getTimeEstimatesForFee(ctx, mediumPriorityFee, mediumMaxFee)
	highMinTime, highMaxTime := e.getTimeEstimatesForFee(ctx, highPriorityFee, highMaxFee)

	return &FeeSuggestions{
		Low: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(lowPriorityFee),
			SuggestedMaxFeePerGas:         new(big.Int).Set(lowMaxFee),
			MinWaitTimeEstimate:           lowMinTime,
			MaxWaitTimeEstimate:           lowMaxTime,
		},
		Medium: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(mediumPriorityFee),
			SuggestedMaxFeePerGas:         new(big.Int).Set(mediumMaxFee),
			MinWaitTimeEstimate:           mediumMinTime,
			MaxWaitTimeEstimate:           mediumMaxTime,
		},
		High: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(highPriorityFee),
			SuggestedMaxFeePerGas:         new(big.Int).Set(highMaxFee),
			MinWaitTimeEstimate:           highMinTime,
			MaxWaitTimeEstimate:           highMaxTime,
		},
		EstimatedBaseFee:  new(big.Int).Set(estimatedBaseFee),
		NetworkCongestion: congestion,
	}, nil
}

// getLegacySuggestions calculates legacy gas price suggestions
func (e *Estimator) getLegacySuggestions(ctx context.Context) (*FeeSuggestions, error) {
	// Get current gas price
	gasPrice, err := e.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// For legacy networks, use gas price with different multipliers
	lowGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(90))
	lowGasPrice.Div(lowGasPrice, big.NewInt(100)) // 90% of current

	mediumGasPrice := gasPrice // 100% of current

	highGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(110))
	highGasPrice.Div(highGasPrice, big.NewInt(100)) // 110% of current

	// Get accurate time estimates for each priority level (using gas price as both priority and max fee for legacy)
	lowMinTime, lowMaxTime := e.getTimeEstimatesForFee(ctx, lowGasPrice, lowGasPrice)
	mediumMinTime, mediumMaxTime := e.getTimeEstimatesForFee(ctx, mediumGasPrice, mediumGasPrice)
	highMinTime, highMaxTime := e.getTimeEstimatesForFee(ctx, highGasPrice, highGasPrice)

	return &FeeSuggestions{
		Low: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(lowGasPrice),
			SuggestedMaxFeePerGas:         new(big.Int).Set(lowGasPrice),
			MinWaitTimeEstimate:           lowMinTime,
			MaxWaitTimeEstimate:           lowMaxTime,
		},
		Medium: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(mediumGasPrice),
			SuggestedMaxFeePerGas:         new(big.Int).Set(mediumGasPrice),
			MinWaitTimeEstimate:           mediumMinTime,
			MaxWaitTimeEstimate:           mediumMaxTime,
		},
		High: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(highGasPrice),
			SuggestedMaxFeePerGas:         new(big.Int).Set(highGasPrice),
			MinWaitTimeEstimate:           highMinTime,
			MaxWaitTimeEstimate:           highMaxTime,
		},
		EstimatedBaseFee:  big.NewInt(0), // No base fee in legacy
		NetworkCongestion: 0.5,           // Default moderate congestion
	}, nil
}

// getLatestBlock retrieves the latest block information
func (e *Estimator) getLatestBlock(ctx context.Context) (*ethclient.BlockWithFullTxs, error) {
	block, err := e.ethClient.BlockByNumber(ctx, nil) // nil means latest block
	if err != nil {
		return nil, err
	}

	return block, nil
}

// getHistoricalBlocks retrieves historical blocks for analysis
func (e *Estimator) getHistoricalBlocks(ctx context.Context, latestBlockNum *big.Int, count int) ([]*ethclient.BlockWithFullTxs, error) {
	blocks := make([]*ethclient.BlockWithFullTxs, 0, count)

	for i := 0; i < count; i++ {
		blockNum := new(big.Int).Sub(latestBlockNum, big.NewInt(int64(i)))
		if blockNum.Sign() < 0 {
			break
		}

		block, err := e.ethClient.BlockByNumber(ctx, blockNum)
		if err != nil {
			continue // Skip failed blocks
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

// extractPriorityFees extracts priority fees from blocks
func (e *Estimator) extractPriorityFees(blocks []*ethclient.BlockWithFullTxs) []*big.Int {
	var fees []*big.Int

	for _, block := range blocks {
		for _, tx := range block.Transactions {
			var priorityFee *big.Int

			if tx.Type != nil && *tx.Type == 2 && tx.MaxPriorityFeePerGas != nil {
				// EIP-1559 transaction
				priorityFee = tx.MaxPriorityFeePerGas
			} else if tx.GasPrice != nil && block.BaseFeePerGas != nil {
				// Legacy transaction on EIP-1559 network
				priorityFee = new(big.Int).Sub(tx.GasPrice, block.BaseFeePerGas)
				if priorityFee.Sign() < 0 {
					priorityFee = big.NewInt(0)
				}
			}

			if priorityFee != nil && priorityFee.Sign() > 0 {
				fees = append(fees, new(big.Int).Set(priorityFee))
			}
		}
	}

	return fees
}

// getPercentile calculates the percentile value from sorted data
func (e *Estimator) getPercentile(sortedData []*big.Int, percentile float64) *big.Int {
	if len(sortedData) == 0 {
		return big.NewInt(0)
	}

	index := (len(sortedData) * int(math.Ceil(percentile))) / 100
	if index >= len(sortedData) {
		index = len(sortedData) - 1
	}

	return new(big.Int).Set(sortedData[index])
}

// calculateNextBaseFee calculates the next block's base fee using EIP-1559 formula
func (e *Estimator) calculateNextBaseFee(block *ethclient.BlockWithFullTxs) *big.Int {
	if block.BaseFeePerGas == nil {
		return big.NewInt(0)
	}

	baseFee := new(big.Int).Set(block.BaseFeePerGas)
	gasUsed := big.NewInt(int64(block.GasUsed))
	gasLimit := big.NewInt(int64(block.GasLimit))

	// Target gas is half of the gas limit
	targetGas := new(big.Int).Div(gasLimit, big.NewInt(2))

	if gasUsed.Cmp(targetGas) == 0 {
		// Gas used equals target, base fee stays the same
		return baseFee
	}

	// Calculate the change in base fee
	var delta *big.Int
	if gasUsed.Cmp(targetGas) > 0 {
		// Gas used > target, increase base fee
		numerator := new(big.Int).Mul(baseFee, new(big.Int).Sub(gasUsed, targetGas))
		denominator := new(big.Int).Mul(targetGas, big.NewInt(8))
		delta = new(big.Int).Div(numerator, denominator)
		baseFee.Add(baseFee, delta)
	} else {
		// Gas used < target, decrease base fee
		numerator := new(big.Int).Mul(baseFee, new(big.Int).Sub(targetGas, gasUsed))
		denominator := new(big.Int).Mul(targetGas, big.NewInt(8))
		delta = new(big.Int).Div(numerator, denominator)
		baseFee.Sub(baseFee, delta)
	}

	// Ensure base fee doesn't go below 0
	if baseFee.Sign() < 0 {
		baseFee.SetInt64(0)
	}

	return baseFee
}

// calculateNetworkCongestion calculates network congestion from blocks
func (e *Estimator) calculateNetworkCongestion(blocks []*ethclient.BlockWithFullTxs) float64 {
	if len(blocks) == 0 {
		return 0.5 // Default moderate congestion
	}

	totalRatio := 0.0
	for _, block := range blocks {
		gasUsed := float64(block.GasUsed)
		gasLimit := float64(block.GasLimit)
		if gasLimit > 0 {
			totalRatio += gasUsed / gasLimit
		}
	}

	avgRatio := totalRatio / float64(len(blocks))

	// Normalize to 0-1 scale, considering that 50% utilization is the EIP-1559 target
	congestion := avgRatio * 2
	if congestion > 1.0 {
		congestion = 1.0
	}

	return congestion
}

// Close closes the estimator (no-op for EthClient interface)
func (e *Estimator) Close() {
	// EthClient interface doesn't have a Close method
	// The underlying client should be closed by the caller
}

// getFeeHistory retrieves fee history using eth_feeHistory RPC method
// TODO: Implement using the FeeHistory method from EthClient interface
func (e *Estimator) getFeeHistory(ctx context.Context, blockCount int) (*ethclient.FeeHistory, error) {
	return e.ethClient.FeeHistory(ctx, uint64(blockCount), nil, []float64{e.config.LowPercentile, e.config.MediumPercentile, e.config.HighPercentile})
}

// calculatePriorityFeeFromHistory calculates priority fee for a given percentile from fee history
func (e *Estimator) calculatePriorityFeeFromHistory(feeHistory *ethclient.FeeHistory, targetPercentile float64) *big.Int {
	if len(feeHistory.Reward) == 0 {
		return big.NewInt(1000000000) // 1 gwei fallback
	}

	// Find which percentile index matches our target
	percentiles := []float64{e.config.LowPercentile, e.config.MediumPercentile, e.config.HighPercentile}
	percentileIndex := -1
	for i, p := range percentiles {
		if math.Abs(p-targetPercentile) < 1 {
			percentileIndex = i
			break
		}
	}

	if percentileIndex == -1 {
		// Fallback: use medium percentile
		fmt.Printf("No percentile index found for %f, using medium percentile\n", targetPercentile)
		percentileIndex = 1
	}

	// Collect all priority fees for this percentile across all blocks
	var priorityFees []*big.Int
	for _, blockRewards := range feeHistory.Reward {
		if percentileIndex < len(blockRewards) && blockRewards[percentileIndex] != nil {
			priorityFees = append(priorityFees, new(big.Int).Set(blockRewards[percentileIndex]))
		}
	}

	if len(priorityFees) == 0 {
		fmt.Printf("No priority fees found for %f, using 1 gwei fallback\n", targetPercentile)
		return big.NewInt(1000000000) // 1 gwei fallback
	}

	// Calculate median of the collected priority fees for stability
	sort.Slice(priorityFees, func(i, j int) bool {
		return priorityFees[i].Cmp(priorityFees[j]) < 0
	})

	medianIndex := len(priorityFees) / 2
	medianFee := priorityFees[medianIndex]

	return medianFee
}

// calculateNextBaseFeeFromHistory calculates next block's base fee using fee history
func (e *Estimator) calculateNextBaseFeeFromHistory(feeHistory *ethclient.FeeHistory, currentBaseFee *big.Int) *big.Int {
	if len(feeHistory.GasUsedRatio) == 0 || currentBaseFee == nil {
		return currentBaseFee
	}

	// Use the most recent gas used ratio
	latestGasUsedRatio := feeHistory.GasUsedRatio[len(feeHistory.GasUsedRatio)-1]

	// EIP-1559 base fee calculation
	// If gas used > 50% of limit, increase base fee
	// If gas used < 50% of limit, decrease base fee
	targetUtilization := 0.5

	baseFee := new(big.Int).Set(currentBaseFee)

	if latestGasUsedRatio > targetUtilization {
		// Increase base fee
		delta := new(big.Int).Mul(baseFee, big.NewInt(int64((latestGasUsedRatio-targetUtilization)*1000)))
		delta.Div(delta, big.NewInt(int64(targetUtilization*8*1000))) // Divide by 8 as per EIP-1559
		baseFee.Add(baseFee, delta)
	} else if latestGasUsedRatio < targetUtilization {
		// Decrease base fee
		delta := new(big.Int).Mul(baseFee, big.NewInt(int64((targetUtilization-latestGasUsedRatio)*1000)))
		delta.Div(delta, big.NewInt(int64(targetUtilization*8*1000))) // Divide by 8 as per EIP-1559
		baseFee.Sub(baseFee, delta)
	}

	// Ensure base fee doesn't go below 0
	if baseFee.Sign() < 0 {
		baseFee.SetInt64(0)
	}

	return baseFee
}

// calculateNetworkCongestionFromHistory calculates network congestion from fee history
func (e *Estimator) calculateNetworkCongestionFromHistory(feeHistory *ethclient.FeeHistory) float64 {
	if len(feeHistory.GasUsedRatio) == 0 {
		return 0.5 // Default moderate congestion
	}

	// Calculate average gas used ratio
	totalRatio := 0.0
	for _, ratio := range feeHistory.GasUsedRatio {
		totalRatio += ratio
	}
	avgRatio := totalRatio / float64(len(feeHistory.GasUsedRatio))

	// Normalize to 0-1 scale, considering that 50% utilization is the EIP-1559 target
	congestion := avgRatio * 2
	if congestion > 1.0 {
		congestion = 1.0
	}

	return congestion
}

// getEIP1559Suggestions calculates EIP-1559 fee suggestions using eth_feeHistory with fallback
func (e *Estimator) getEIP1559Suggestions(ctx context.Context, latestBlock *ethclient.BlockWithFullTxs) (*FeeSuggestions, error) {
	// Try to use eth_feeHistory for more accurate data
	feeHistory, err := e.getFeeHistory(ctx, e.config.PercentileBlocks)
	if err != nil {
		// Fallback to the legacy method if eth_feeHistory is not available
		fmt.Printf("eth_feeHistory not available, falling back to legacy method: %v\n", err)
		return e.getEIP1559SuggestionsLegacy(ctx, latestBlock)
	}

	// Extract priority fees from fee history
	lowPriorityFee := e.calculatePriorityFeeFromHistory(feeHistory, e.config.LowPercentile)
	mediumPriorityFee := e.calculatePriorityFeeFromHistory(feeHistory, e.config.MediumPercentile)
	highPriorityFee := e.calculatePriorityFeeFromHistory(feeHistory, e.config.HighPercentile)

	// Get the latest base fee from fee history
	var latestBaseFee *big.Int
	if len(feeHistory.BaseFeePerGas) > 0 {
		latestBaseFee = feeHistory.BaseFeePerGas[len(feeHistory.BaseFeePerGas)-1]
	} else {
		latestBaseFee = latestBlock.BaseFeePerGas
	}

	// Calculate next block's estimated base fee using the most recent gas usage
	estimatedBaseFee := e.calculateNextBaseFeeFromHistory(feeHistory, latestBaseFee)

	// Apply multiplier for buffer
	bufferedBaseFee := new(big.Int).Mul(estimatedBaseFee, big.NewInt(int64(e.config.BaseFeeMultiplier*1000)))
	bufferedBaseFee.Div(bufferedBaseFee, big.NewInt(1000))

	// Calculate max fee per gas (base fee + priority fee)
	lowMaxFee := new(big.Int).Add(bufferedBaseFee, lowPriorityFee)
	mediumMaxFee := new(big.Int).Add(bufferedBaseFee, mediumPriorityFee)
	highMaxFee := new(big.Int).Add(bufferedBaseFee, highPriorityFee)

	// Calculate network congestion from fee history
	congestion := e.calculateNetworkCongestionFromHistory(feeHistory)

	// Get accurate time estimates for each priority level
	lowMinTime, lowMaxTime := e.getTimeEstimatesForFee(ctx, lowPriorityFee, lowMaxFee)
	mediumMinTime, mediumMaxTime := e.getTimeEstimatesForFee(ctx, mediumPriorityFee, mediumMaxFee)
	highMinTime, highMaxTime := e.getTimeEstimatesForFee(ctx, highPriorityFee, highMaxFee)

	return &FeeSuggestions{
		Low: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(lowPriorityFee),
			SuggestedMaxFeePerGas:         new(big.Int).Set(lowMaxFee),
			MinWaitTimeEstimate:           lowMinTime,
			MaxWaitTimeEstimate:           lowMaxTime,
		},
		Medium: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(mediumPriorityFee),
			SuggestedMaxFeePerGas:         new(big.Int).Set(mediumMaxFee),
			MinWaitTimeEstimate:           mediumMinTime,
			MaxWaitTimeEstimate:           mediumMaxTime,
		},
		High: FeeSuggestion{
			SuggestedMaxPriorityFeePerGas: new(big.Int).Set(highPriorityFee),
			SuggestedMaxFeePerGas:         new(big.Int).Set(highMaxFee),
			MinWaitTimeEstimate:           highMinTime,
			MaxWaitTimeEstimate:           highMaxTime,
		},
		EstimatedBaseFee:  new(big.Int).Set(estimatedBaseFee),
		NetworkCongestion: congestion,
	}, nil
}
