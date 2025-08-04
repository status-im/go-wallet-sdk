package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"gas-comparison/internal/old"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

// NetworkInfo represents information about a network to test
type NetworkInfo struct {
	Name         string
	ChainID      int
	RPC          string
	AvgBlockTime float64
}

func main() {
	fmt.Println("🔥 Multi-Network Gas Fee Comparison Tool")
	fmt.Println("Comparing our current implementation, old implementation, and Infura's Gas API across multiple networks")
	fmt.Println(strings.Repeat("=", 80))

	infuraToken := os.Getenv("INFURA_TOKEN")
	if infuraToken == "" {
		fmt.Println("INFURA_TOKEN is not set")
		os.Exit(1)
	}

	// Define networks to test
	networks := []NetworkInfo{
		{
			Name:         "Ethereum Mainnet",
			ChainID:      infura.Ethereum,
			RPC:          "https://mainnet.infura.io/v3/" + infuraToken,
			AvgBlockTime: 12,
		},
		{
			Name:         "Arbitrum One",
			ChainID:      infura.ArbitrumOne,
			RPC:          "https://arbitrum-mainnet.infura.io/v3/" + infuraToken,
			AvgBlockTime: 0.25,
		},
		{
			Name:         "Optimism",
			ChainID:      infura.Optimism,
			RPC:          "https://optimism-mainnet.infura.io/v3/" + infuraToken,
			AvgBlockTime: 2.0,
		},
		{
			Name:         "Polygon",
			ChainID:      infura.Polygon,
			RPC:          "https://polygon-mainnet.infura.io/v3/" + infuraToken,
			AvgBlockTime: 2.25,
		},
		{
			Name:         "Base",
			ChainID:      infura.Base,
			RPC:          "https://base-mainnet.infura.io/v3/" + infuraToken,
			AvgBlockTime: 2,
		},
		{
			Name:         "Status Network Sepolia",
			ChainID:      1660990954,
			RPC:          "https://public.sepolia.rpc.status.network",
			AvgBlockTime: 2,
		},
	}

	// Test each network
	for i, network := range networks {
		fmt.Printf("\n%s %s (%d)\n", getNetworkEmoji(network.ChainID), network.Name, network.ChainID)
		fmt.Println(strings.Repeat("-", 60))

		err := compareNetwork(network)
		if err != nil {
			fmt.Printf("❌ Error testing %s: %v\n", network.Name, err)
			continue
		}

		// Add spacing between networks (except for the last one)
		if i < len(networks)-1 {
			fmt.Println()
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("✅ Multi-network comparison complete!")
}

func displayComparison(current *gas.FeeSuggestions, oldFees *old.SuggestedFees, infura *infura.GasResponse) {
	fmt.Println("\n📋 COMPARISON RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	// Use wei values directly instead of converting to gwei
	currentLowPriority := current.Low.SuggestedMaxPriorityFeePerGas
	oldLowPriority := oldFees.MaxFeesLevels.LowPriority.ToInt()
	infuraLowPriority := stringToWei(infura.Low.SuggestedMaxPriorityFeePerGas)

	currentMediumPriority := current.Medium.SuggestedMaxPriorityFeePerGas
	oldMediumPriority := oldFees.MaxFeesLevels.MediumPriority.ToInt()
	infuraMediumPriority := stringToWei(infura.Medium.SuggestedMaxPriorityFeePerGas)

	currentHighPriority := current.High.SuggestedMaxPriorityFeePerGas
	oldHighPriority := oldFees.MaxFeesLevels.HighPriority.ToInt()
	infuraHighPriority := stringToWei(infura.High.SuggestedMaxPriorityFeePerGas)

	currentBaseFee := current.EstimatedBaseFee
	oldBaseFee := oldFees.CurrentBaseFee
	infuraBaseFee := stringToWei(infura.EstimatedBaseFee)

	fmt.Printf("🔸 LOW PRIORITY FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentLowPriority.String())
	fmt.Printf("   Old Implementation:     %20s wei\n", oldLowPriority.String())
	fmt.Printf("   Infura:                 %20s wei\n", infuraLowPriority.String())
	fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
		new(big.Int).Sub(currentLowPriority, oldLowPriority).String(),
		percentDiffWei(currentLowPriority, oldLowPriority))
	fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
		new(big.Int).Sub(currentLowPriority, infuraLowPriority).String(),
		percentDiffWei(currentLowPriority, infuraLowPriority))
	fmt.Printf("   Wait Time (Current):    %d-%d seconds\n",
		current.Low.MinWaitTimeEstimate, current.Low.MaxWaitTimeEstimate)
	fmt.Printf("   Wait Time (Old):        %d seconds\n", oldFees.MaxFeesLevels.LowEstimatedTime)
	fmt.Printf("   Wait Time (Infura):     %.1f-%.1f seconds\n\n",
		float64(infura.Low.MinWaitTimeEstimate)/1000, float64(infura.Low.MaxWaitTimeEstimate)/1000)

	fmt.Printf("🔸 MEDIUM PRIORITY FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentMediumPriority.String())
	fmt.Printf("   Old Implementation:     %20s wei\n", oldMediumPriority.String())
	fmt.Printf("   Infura:                 %20s wei\n", infuraMediumPriority.String())
	fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
		new(big.Int).Sub(currentMediumPriority, oldMediumPriority).String(),
		percentDiffWei(currentMediumPriority, oldMediumPriority))
	fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
		new(big.Int).Sub(currentMediumPriority, infuraMediumPriority).String(),
		percentDiffWei(currentMediumPriority, infuraMediumPriority))
	fmt.Printf("   Wait Time (Current):    %d-%d seconds\n",
		current.Medium.MinWaitTimeEstimate, current.Medium.MaxWaitTimeEstimate)
	fmt.Printf("   Wait Time (Old):        %d seconds\n", oldFees.MaxFeesLevels.MediumEstimatedTime)
	fmt.Printf("   Wait Time (Infura):     %.1f-%.1f seconds\n\n",
		float64(infura.Medium.MinWaitTimeEstimate)/1000, float64(infura.Medium.MaxWaitTimeEstimate)/1000)

	fmt.Printf("🔸 HIGH PRIORITY FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentHighPriority.String())
	fmt.Printf("   Old Implementation:     %20s wei\n", oldHighPriority.String())
	fmt.Printf("   Infura:                 %20s wei\n", infuraHighPriority.String())
	fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
		new(big.Int).Sub(currentHighPriority, oldHighPriority).String(),
		percentDiffWei(currentHighPriority, oldHighPriority))
	fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
		new(big.Int).Sub(currentHighPriority, infuraHighPriority).String(),
		percentDiffWei(currentHighPriority, infuraHighPriority))
	fmt.Printf("   Wait Time (Current):    %d-%d seconds\n",
		current.High.MinWaitTimeEstimate, current.High.MaxWaitTimeEstimate)
	fmt.Printf("   Wait Time (Old):        %d seconds\n", oldFees.MaxFeesLevels.HighEstimatedTime)
	fmt.Printf("   Wait Time (Infura):     %.1f-%.1f seconds\n\n",
		float64(infura.High.MinWaitTimeEstimate)/1000, float64(infura.High.MaxWaitTimeEstimate)/1000)

	fmt.Printf("🔸 BASE FEE\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentBaseFee.String())
	fmt.Printf("   Old Implementation:     %20s wei\n", oldBaseFee.String())
	fmt.Printf("   Infura:                 %20s wei\n", infuraBaseFee.String())
	fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
		new(big.Int).Sub(currentBaseFee, oldBaseFee).String(),
		percentDiffWei(currentBaseFee, oldBaseFee))
	fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n\n",
		new(big.Int).Sub(currentBaseFee, infuraBaseFee).String(),
		percentDiffWei(currentBaseFee, infuraBaseFee))

	fmt.Printf("🔸 NETWORK CONGESTION\n")
	fmt.Printf("   Current Implementation: %.3f\n", current.NetworkCongestion)
	fmt.Printf("   Infura:                 %.3f\n", infura.NetworkCongestion)
	fmt.Printf("   Difference:             %+.3f\n", current.NetworkCongestion-infura.NetworkCongestion)
}

func hexToWei(hexStr string) *big.Int {
	if hexStr == "" || hexStr == "0x0" {
		return big.NewInt(0)
	}

	wei, err := hexutil.DecodeBig(hexStr)
	if err != nil {
		return big.NewInt(0)
	}

	return wei
}

func stringToWei(str string) *big.Int {
	if str == "" {
		return big.NewInt(0)
	}

	// Try parsing as decimal first (Infura format)
	if val, err := strconv.ParseFloat(str, 64); err == nil {
		// Convert from gwei to wei if it's a small number (likely gwei)
		if val < 1000 {
			val = val * 1e9 // Convert gwei to wei
		}
		return big.NewInt(int64(val))
	}

	// Fall back to hex parsing
	return hexToWei(str)
}

func percentDiffWei(a, b *big.Int) float64 {
	if b.Sign() == 0 {
		return 0
	}

	// Convert to float64 for percentage calculation
	aFloat := new(big.Float).SetInt(a)
	bFloat := new(big.Float).SetInt(b)

	diff := new(big.Float).Sub(aFloat, bFloat)
	percent := new(big.Float).Quo(diff, bFloat)
	percent.Mul(percent, big.NewFloat(100))

	result, _ := percent.Float64()
	return result
}

// compareNetwork compares gas fees for a specific network
func compareNetwork(network NetworkInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	rpcClient, err := rpc.Dial(network.RPC)
	if err != nil {
		return fmt.Errorf("failed to dial RPC: %w", err)
	}
	ethClient := ethclient.NewClient(rpcClient)
	if err != nil {
		return fmt.Errorf("failed to create eth client: %w", err)
	}

	// Initialize our estimator for this network
	estimatorConfig := gas.DefaultConfig()
	estimatorConfig.NetworkBlockTime = network.AvgBlockTime
	estimator, err := gas.NewEstimatorWithConfig(ethClient, estimatorConfig)
	if err != nil {
		return fmt.Errorf("failed to create estimator: %w", err)
	}
	defer estimator.Close()

	// Get our suggestions
	fmt.Printf("📊 Fetching our gas suggestions for %s...\n", network.Name)
	ourSuggestions, err := estimator.GetFeeSuggestions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get our suggestions: %w", err)
	}

	// Get old estimator suggestions
	fmt.Printf("📊 Fetching old estimator suggestions for %s...\n", network.Name)
	oldFeeManager := old.NewFeeManager(ethClient)
	oldSuggestions, _, _, err := oldFeeManager.SuggestedFees(ctx, uint64(network.ChainID), old.ZeroAddress())
	if err != nil {
		fmt.Printf("⚠️  Old estimator not available for %s: %v\n", network.Name, err)
		oldSuggestions = nil
	}

	// Get Infura's suggestions for this network
	fmt.Printf("📊 Fetching Infura's gas suggestions for %s...\n", network.Name)
	infuraSuggestions, err := getInfuraGasSuggestionsForNetwork(network.ChainID)
	if err != nil {
		fmt.Printf("⚠️  Infura API not available for %s: %v\n", network.Name, err)
		if oldSuggestions != nil {
			fmt.Printf("📊 Showing our implementation vs old implementation:\n")
			displayOldVsNewComparison(ourSuggestions, oldSuggestions)
		} else {
			fmt.Printf("📊 Showing only our implementation results:\n")
			displayOurSuggestions(ourSuggestions)
		}
		return nil
	}

	// Display comparison
	displayComparison(ourSuggestions, oldSuggestions, infuraSuggestions)

	return nil
}

// getInfuraGasSuggestionsForNetwork gets Infura suggestions for a specific network
func getInfuraGasSuggestionsForNetwork(chainID int) (*infura.GasResponse, error) {
	infuraToken := os.Getenv("INFURA_TOKEN")
	if infuraToken == "" {
		return nil, fmt.Errorf("INFURA_TOKEN is not set")
	}

	client := infura.NewClient(infuraToken)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.GetGasSuggestions(ctx, chainID)
}

// displayOldVsNewComparison displays comparison between old and new implementations
func displayOldVsNewComparison(newSuggestions *gas.FeeSuggestions, oldSuggestions *old.SuggestedFees) {
	// Use wei values directly instead of converting to gwei
	newLowPriority := newSuggestions.Low.SuggestedMaxPriorityFeePerGas
	oldLowPriority := oldSuggestions.MaxFeesLevels.LowPriority.ToInt()

	newMediumPriority := newSuggestions.Medium.SuggestedMaxPriorityFeePerGas
	oldMediumPriority := oldSuggestions.MaxFeesLevels.MediumPriority.ToInt()

	newHighPriority := newSuggestions.High.SuggestedMaxPriorityFeePerGas
	oldHighPriority := oldSuggestions.MaxFeesLevels.HighPriority.ToInt()

	newBaseFee := newSuggestions.EstimatedBaseFee
	oldBaseFee := oldSuggestions.CurrentBaseFee

	// Use wei values directly for max fees
	newLowMaxFee := newSuggestions.Low.SuggestedMaxFeePerGas
	oldLowMaxFee := oldSuggestions.MaxFeesLevels.Low.ToInt()

	newMediumMaxFee := newSuggestions.Medium.SuggestedMaxFeePerGas
	oldMediumMaxFee := oldSuggestions.MaxFeesLevels.Medium.ToInt()

	newHighMaxFee := newSuggestions.High.SuggestedMaxFeePerGas
	oldHighMaxFee := oldSuggestions.MaxFeesLevels.High.ToInt()

	fmt.Printf("📋 OLD vs NEW IMPLEMENTATION COMPARISON\n")

	// Priority Fee Comparison Table
	fmt.Printf("\n🔸 PRIORITY FEE COMPARISON (wei)\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Level  │ New Implementation │ Old Implementation │ Difference │ Percentage │ Status\n")
	fmt.Printf("───────┼────────────────────┼────────────────────┼────────────┼────────────┼────────\n")
	fmt.Printf("Low    │ %20s │ %20s │ %+12s │ %+9.1f%% │ %s\n",
		newLowPriority.String(), oldLowPriority.String(),
		new(big.Int).Sub(newLowPriority, oldLowPriority).String(),
		percentDiffWei(newLowPriority, oldLowPriority), getStatusEmoji(percentDiffWei(newLowPriority, oldLowPriority)))
	fmt.Printf("Medium │ %20s │ %20s │ %+12s │ %+9.1f%% │ %s\n",
		newMediumPriority.String(), oldMediumPriority.String(),
		new(big.Int).Sub(newMediumPriority, oldMediumPriority).String(),
		percentDiffWei(newMediumPriority, oldMediumPriority), getStatusEmoji(percentDiffWei(newMediumPriority, oldMediumPriority)))
	fmt.Printf("High   │ %20s │ %20s │ %+12s │ %+9.1f%% │ %s\n",
		newHighPriority.String(), oldHighPriority.String(),
		new(big.Int).Sub(newHighPriority, oldHighPriority).String(),
		percentDiffWei(newHighPriority, oldHighPriority), getStatusEmoji(percentDiffWei(newHighPriority, oldHighPriority)))
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")

	// Max Fee Comparison Table
	fmt.Printf("\n🔸 MAX FEE COMPARISON (wei)\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Level  │ New Implementation │ Old Implementation │ Difference │ Percentage │ Status\n")
	fmt.Printf("───────┼────────────────────┼────────────────────┼────────────┼────────────┼────────\n")
	fmt.Printf("Low    │ %20s │ %20s │ %+12s │ %+9.1f%% │ %s\n",
		newLowMaxFee.String(), oldLowMaxFee.String(),
		new(big.Int).Sub(newLowMaxFee, oldLowMaxFee).String(),
		percentDiffWei(newLowMaxFee, oldLowMaxFee), getStatusEmoji(percentDiffWei(newLowMaxFee, oldLowMaxFee)))
	fmt.Printf("Medium │ %20s │ %20s │ %+12s │ %+9.1f%% │ %s\n",
		newMediumMaxFee.String(), oldMediumMaxFee.String(),
		new(big.Int).Sub(newMediumMaxFee, oldMediumMaxFee).String(),
		percentDiffWei(newMediumMaxFee, oldMediumMaxFee), getStatusEmoji(percentDiffWei(newMediumMaxFee, oldMediumMaxFee)))
	fmt.Printf("High   │ %20s │ %20s │ %+12s │ %+9.1f%% │ %s\n",
		newHighMaxFee.String(), oldHighMaxFee.String(),
		new(big.Int).Sub(newHighMaxFee, oldHighMaxFee).String(),
		percentDiffWei(newHighMaxFee, oldHighMaxFee), getStatusEmoji(percentDiffWei(newHighMaxFee, oldHighMaxFee)))
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")

	// Wait Time Comparison Table
	fmt.Printf("\n🔸 WAIT TIME COMPARISON (seconds)\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Level  │ New Implementation │ Old Implementation │ Time Difference │ Estimation Method\n")
	fmt.Printf("       │ Min    │ Max       │ Min      │ Max         │ Min    │ Max    │ Comparison\n")
	fmt.Printf("───────┼────────┼───────────┼──────────┼─────────────┼────────┼────────┼────────────\n")
	fmt.Printf("Low    │ %6d │ %9d │ %8d │ %11d │ %+6.0f │ %+6.0f │ %s\n",
		newSuggestions.Low.MinWaitTimeEstimate, newSuggestions.Low.MaxWaitTimeEstimate,
		oldSuggestions.MaxFeesLevels.LowEstimatedTime, oldSuggestions.MaxFeesLevels.LowEstimatedTime,
		float64(newSuggestions.Low.MinWaitTimeEstimate)-float64(oldSuggestions.MaxFeesLevels.LowEstimatedTime),
		float64(newSuggestions.Low.MaxWaitTimeEstimate)-float64(oldSuggestions.MaxFeesLevels.LowEstimatedTime),
		getTimeComparisonStatus(newSuggestions.Low.MinWaitTimeEstimate, newSuggestions.Low.MaxWaitTimeEstimate, int(oldSuggestions.MaxFeesLevels.LowEstimatedTime), int(oldSuggestions.MaxFeesLevels.LowEstimatedTime)))
	fmt.Printf("Medium │ %6d │ %9d │ %8d │ %11d │ %+6.0f │ %+6.0f │ %s\n",
		newSuggestions.Medium.MinWaitTimeEstimate, newSuggestions.Medium.MaxWaitTimeEstimate,
		oldSuggestions.MaxFeesLevels.MediumEstimatedTime, oldSuggestions.MaxFeesLevels.MediumEstimatedTime,
		float64(newSuggestions.Medium.MinWaitTimeEstimate)-float64(oldSuggestions.MaxFeesLevels.MediumEstimatedTime),
		float64(newSuggestions.Medium.MaxWaitTimeEstimate)-float64(oldSuggestions.MaxFeesLevels.MediumEstimatedTime),
		getTimeComparisonStatus(newSuggestions.Medium.MinWaitTimeEstimate, newSuggestions.Medium.MaxWaitTimeEstimate, int(oldSuggestions.MaxFeesLevels.MediumEstimatedTime), int(oldSuggestions.MaxFeesLevels.MediumEstimatedTime)))
	fmt.Printf("High   │ %6d │ %9d │ %8d │ %11d │ %+6.0f │ %+6.0f │ %s\n",
		newSuggestions.High.MinWaitTimeEstimate, newSuggestions.High.MaxWaitTimeEstimate,
		oldSuggestions.MaxFeesLevels.HighEstimatedTime, oldSuggestions.MaxFeesLevels.HighEstimatedTime,
		float64(newSuggestions.High.MinWaitTimeEstimate)-float64(oldSuggestions.MaxFeesLevels.HighEstimatedTime),
		float64(newSuggestions.High.MaxWaitTimeEstimate)-float64(oldSuggestions.MaxFeesLevels.HighEstimatedTime),
		getTimeComparisonStatus(newSuggestions.High.MinWaitTimeEstimate, newSuggestions.High.MaxWaitTimeEstimate, int(oldSuggestions.MaxFeesLevels.HighEstimatedTime), int(oldSuggestions.MaxFeesLevels.HighEstimatedTime)))
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════\n")

	fmt.Printf("\n🔸 Base Fee: %20s wei (new) vs %20s wei (old) - %+.1f%% diff\n",
		newBaseFee.String(), oldBaseFee.String(), percentDiffWei(newBaseFee, oldBaseFee))

	fmt.Printf("🔸 Network Congestion: %.1f%% (new implementation)\n",
		newSuggestions.NetworkCongestion*100)
}

// displayOurSuggestions displays only our suggestions when Infura is not available
func displayOurSuggestions(suggestions *gas.FeeSuggestions) {
	ourLow := suggestions.Low.SuggestedMaxPriorityFeePerGas
	ourMedium := suggestions.Medium.SuggestedMaxPriorityFeePerGas
	ourHigh := suggestions.High.SuggestedMaxPriorityFeePerGas
	ourBaseFee := suggestions.EstimatedBaseFee

	// Use wei values directly for max fees
	ourLowMaxFee := suggestions.Low.SuggestedMaxFeePerGas
	ourMediumMaxFee := suggestions.Medium.SuggestedMaxFeePerGas
	ourHighMaxFee := suggestions.High.SuggestedMaxFeePerGas

	fmt.Printf("📋 OUR IMPLEMENTATION RESULTS\n")

	// Priority Fee Table
	fmt.Printf("\n🔸 PRIORITY FEES (wei)\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Priority Level │ Fee (wei)           │ Competitiveness │ Fee Category\n")
	fmt.Printf("───────────────┼─────────────────────┼─────────────────┼──────────────────\n")
	fmt.Printf("Low            │ %20s │ %15s │ %s\n", ourLow.String(), getFeeCompetitivenessWei(ourLow), getFeeCategoryEmojiWei(ourLow))
	fmt.Printf("Medium         │ %20s │ %15s │ %s\n", ourMedium.String(), getFeeCompetitivenessWei(ourMedium), getFeeCategoryEmojiWei(ourMedium))
	fmt.Printf("High           │ %20s │ %15s │ %s\n", ourHigh.String(), getFeeCompetitivenessWei(ourHigh), getFeeCategoryEmojiWei(ourHigh))
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")

	// Max Fee Table
	fmt.Printf("\n🔸 MAX FEES (wei)\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Priority Level │ Max Fee             │ Priority Fee    │ Base Fee      │ Fee Breakdown\n")
	fmt.Printf("───────────────┼─────────────────────┼─────────────────┼───────────────┼──────────────────\n")
	fmt.Printf("Low            │ %20s │ %15s │ %13s │ %s + %s\n",
		ourLowMaxFee.String(), ourLow.String(), ourBaseFee.String(), ourLow.String(), new(big.Int).Sub(ourLowMaxFee, ourLow).String())
	fmt.Printf("Medium         │ %20s │ %15s │ %13s │ %s + %s\n",
		ourMediumMaxFee.String(), ourMedium.String(), ourBaseFee.String(), ourMedium.String(), new(big.Int).Sub(ourMediumMaxFee, ourMedium).String())
	fmt.Printf("High           │ %20s │ %15s │ %13s │ %s + %s\n",
		ourHighMaxFee.String(), ourHigh.String(), ourBaseFee.String(), ourHigh.String(), new(big.Int).Sub(ourHighMaxFee, ourHigh).String())
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")

	// Wait Time Table
	fmt.Printf("\n🔸 WAIT TIME ESTIMATES (seconds)\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Priority Level │ Min Time │ Max Time │ Average │ Time Category\n")
	fmt.Printf("───────────────┼──────────┼──────────┼─────────┼──────────────────\n")
	fmt.Printf("Low            │ %8d │ %8d │ %7.0f │ %s\n",
		suggestions.Low.MinWaitTimeEstimate, suggestions.Low.MaxWaitTimeEstimate,
		float64(suggestions.Low.MinWaitTimeEstimate+suggestions.Low.MaxWaitTimeEstimate)/2,
		getTimeCategory(suggestions.Low.MinWaitTimeEstimate, suggestions.Low.MaxWaitTimeEstimate))
	fmt.Printf("Medium         │ %8d │ %8d │ %7.0f │ %s\n",
		suggestions.Medium.MinWaitTimeEstimate, suggestions.Medium.MaxWaitTimeEstimate,
		float64(suggestions.Medium.MinWaitTimeEstimate+suggestions.Medium.MaxWaitTimeEstimate)/2,
		getTimeCategory(suggestions.Medium.MinWaitTimeEstimate, suggestions.Medium.MaxWaitTimeEstimate))
	fmt.Printf("High           │ %8d │ %8d │ %7.0f │ %s\n",
		suggestions.High.MinWaitTimeEstimate, suggestions.High.MaxWaitTimeEstimate,
		float64(suggestions.High.MinWaitTimeEstimate+suggestions.High.MaxWaitTimeEstimate)/2,
		getTimeCategory(suggestions.High.MinWaitTimeEstimate, suggestions.High.MaxWaitTimeEstimate))
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")

	fmt.Printf("\n🔸 Base Fee: %20s wei\n", ourBaseFee.String())
	fmt.Printf("🔸 Network Congestion: %.1f%%\n", suggestions.NetworkCongestion*100)
}

// Update helper functions to work with wei values
func getFeeCompetitivenessWei(feeWei *big.Int) string {
	// Convert wei to gwei for comparison (1 gwei = 1e9 wei)
	feeGwei := new(big.Float).Quo(new(big.Float).SetInt(feeWei), big.NewFloat(1e9))
	feeFloat, _ := feeGwei.Float64()

	if feeFloat >= 5.0 {
		return "Very High"
	} else if feeFloat >= 3.0 {
		return "High"
	} else if feeFloat >= 1.5 {
		return "Medium"
	} else if feeFloat >= 0.5 {
		return "Low"
	} else {
		return "Very Low"
	}
}

func getFeeCategoryEmojiWei(feeWei *big.Int) string {
	// Convert wei to gwei for comparison (1 gwei = 1e9 wei)
	feeGwei := new(big.Float).Quo(new(big.Float).SetInt(feeWei), big.NewFloat(1e9))
	feeFloat, _ := feeGwei.Float64()

	if feeFloat >= 5.0 {
		return "🚀 Premium"
	} else if feeFloat >= 3.0 {
		return "⚡ Fast"
	} else if feeFloat >= 1.5 {
		return "🟢 Standard"
	} else if feeFloat >= 0.5 {
		return "🟡 Economy"
	} else {
		return "🟠 Slow"
	}
}

// getNetworkEmoji returns an emoji for the network
func getNetworkEmoji(chainID int) string {
	switch chainID {
	case infura.Ethereum:
		return "🔷" // Ethereum blue diamond
	case infura.ArbitrumOne:
		return "🔵" // Arbitrum blue circle
	case infura.Optimism:
		return "🔴" // Optimism red circle
	case infura.Polygon:
		return "🟣" // Polygon purple circle
	case infura.Base:
		return "🔵" // Base blue circle
	default:
		return "⚪" // Default white circle
	}
}

// getStatusEmoji returns a status emoji based on percentage difference
func getStatusEmoji(percentDiff float64) string {
	absDiff := percentDiff
	if absDiff < 0 {
		absDiff = -absDiff
	}

	if absDiff <= 5 {
		return "✅ Excellent"
	} else if absDiff <= 15 {
		return "🟢 Good"
	} else if absDiff <= 30 {
		return "🟡 Fair"
	} else if absDiff <= 50 {
		return "🟠 Poor"
	} else {
		return "🔴 Very Different"
	}
}

// getTimeComparisonStatus returns a comparison status for wait times
func getTimeComparisonStatus(ourMin, ourMax, infuraMin, infuraMax int) string {
	infuraMinSec := float64(infuraMin) / 1000
	infuraMaxSec := float64(infuraMax) / 1000

	// Check if our ranges overlap with Infura's ranges
	ourMinFloat := float64(ourMin)
	ourMaxFloat := float64(ourMax)

	// Calculate overlap
	overlapMin := ourMinFloat
	if infuraMinSec > overlapMin {
		overlapMin = infuraMinSec
	}

	overlapMax := ourMaxFloat
	if infuraMaxSec < overlapMax {
		overlapMax = infuraMaxSec
	}

	if overlapMax > overlapMin {
		// There's overlap
		overlapPercent := (overlapMax - overlapMin) / ((ourMaxFloat - ourMinFloat + infuraMaxSec - infuraMinSec) / 2) * 100
		if overlapPercent > 80 {
			return "✅ Very Similar"
		} else if overlapPercent > 50 {
			return "🟢 Similar"
		} else {
			return "🟡 Some Overlap"
		}
	} else {
		// No overlap
		if ourMaxFloat < infuraMinSec {
			return "🔵 Ours Faster"
		} else {
			return "🔴 Ours Slower"
		}
	}
}

// getTimeCategory returns a category description for wait times
func getTimeCategory(minTime, maxTime int) string {
	avgTime := (minTime + maxTime) / 2

	if avgTime <= 15 {
		return "⚡ Instant"
	} else if avgTime <= 60 {
		return "🟢 Fast"
	} else if avgTime <= 180 {
		return "🟡 Moderate"
	} else if avgTime <= 300 {
		return "🟠 Slow"
	} else {
		return "🔴 Very Slow"
	}
}
