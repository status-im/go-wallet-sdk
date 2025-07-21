package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"gas-comparison/data"
	data_arbitrum "gas-comparison/data/arbitrum"
	data_base "gas-comparison/data/base"
	data_ethereum "gas-comparison/data/ethereum"
	data_optimism "gas-comparison/data/optimism"
	data_polygon "gas-comparison/data/polygon"
	"gas-comparison/old"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

// NetworkInfo represents information about a network to test
type NetworkInfo struct {
	Name              string
	ChainID           int
	RPC               string
	SuggestionsConfig gas.SuggestionsConfig
	LocalData         *data.GasData
}

type GasDataClient interface {
	gas.EthClient
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*ethclient.BlockWithFullTxs, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	GetGasSuggestions(ctx context.Context, networkID int) (*infura.GasResponse, error)
	Close()
}

// mustGetGasData is a helper function to handle the error from GetGasData functions
func mustGetGasData(getDataFunc func() (*data.GasData, error)) *data.GasData {
	gasData, err := getDataFunc()
	if err != nil {
		fmt.Printf("Error loading gas data: %v\n", err)
		os.Exit(1)
	}
	return gasData
}

func main() {
	// Define command line flags
	var (
		infuraToken = flag.String("infura-api-key", "", "Infura API key for gas suggestions (required for network mode)")
		fake        = flag.Bool("fake", false, "Use local data if set. Otherwise fetch data from the network")
		help        = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate required arguments
	if !*fake && *infuraToken == "" {
		fmt.Fprintf(os.Stderr, "Error: -infura-api-key flag is required in network mode\n\n")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("🔥 Multi-Network Gas Fee Comparison Tool")
	fmt.Println("Comparing our current implementation, old implementation, and Infura's Gas API across multiple networks")
	fmt.Println(strings.Repeat("=", 80))

	ethereumConfig := gas.DefaultConfig()
	ethereumConfig.ChainClass = gas.ChainClassL1
	ethereumConfig.NetworkBlockTime = 12
	ethereumConfig.GasPriceEstimationBlocks = 10

	arbitrumConfig := gas.DefaultConfig()
	arbitrumConfig.ChainClass = gas.ChainClassArbStack
	arbitrumConfig.NetworkBlockTime = 0.25
	arbitrumConfig.GasPriceEstimationBlocks = 50

	optimismConfig := gas.DefaultConfig()
	optimismConfig.ChainClass = gas.ChainClassOPStack
	optimismConfig.NetworkBlockTime = 2.0
	optimismConfig.GasPriceEstimationBlocks = 50

	baseConfig := gas.DefaultConfig()
	baseConfig.ChainClass = gas.ChainClassOPStack
	baseConfig.NetworkBlockTime = 2.0
	baseConfig.GasPriceEstimationBlocks = 50

	polygonConfig := gas.DefaultConfig()
	polygonConfig.ChainClass = gas.ChainClassL1
	polygonConfig.NetworkBlockTime = 2.25
	polygonConfig.GasPriceEstimationBlocks = 50

	bscConfig := gas.DefaultConfig()
	bscConfig.ChainClass = gas.ChainClassL1
	bscConfig.NetworkBlockTime = 0.75
	bscConfig.GasPriceEstimationBlocks = 50

	lineaConfig := gas.DefaultConfig()
	lineaConfig.ChainClass = gas.ChainClassLineaStack
	lineaConfig.NetworkBlockTime = 2.00
	lineaConfig.GasPriceEstimationBlocks = 50

	statusNetworkSepoliaConfig := gas.DefaultConfig()
	statusNetworkSepoliaConfig.ChainClass = gas.ChainClassLineaStack
	statusNetworkSepoliaConfig.NetworkBlockTime = 2.00
	statusNetworkSepoliaConfig.GasPriceEstimationBlocks = 50

	// Define networks to test
	networks := []NetworkInfo{
		{
			Name:              "Ethereum Mainnet",
			ChainID:           infura.Ethereum,
			RPC:               "https://mainnet.infura.io/v3/" + *infuraToken,
			SuggestionsConfig: ethereumConfig,
			LocalData:         mustGetGasData(data_ethereum.GetGasData),
		},
		{
			Name:              "Arbitrum One",
			ChainID:           infura.ArbitrumOne,
			RPC:               "https://arbitrum-mainnet.infura.io/v3/" + *infuraToken,
			SuggestionsConfig: arbitrumConfig,
			LocalData:         mustGetGasData(data_arbitrum.GetGasData),
		},
		{
			Name:              "Optimism",
			ChainID:           infura.Optimism,
			RPC:               "https://optimism-mainnet.infura.io/v3/" + *infuraToken,
			SuggestionsConfig: optimismConfig,
			LocalData:         mustGetGasData(data_optimism.GetGasData),
		},
		{
			Name:              "Base",
			ChainID:           infura.Base,
			RPC:               "https://base-mainnet.infura.io/v3/" + *infuraToken,
			SuggestionsConfig: baseConfig,
			LocalData:         mustGetGasData(data_base.GetGasData),
		},
		{
			Name:              "Polygon",
			ChainID:           infura.Polygon,
			RPC:               "https://polygon-mainnet.infura.io/v3/" + *infuraToken,
			SuggestionsConfig: polygonConfig,
			LocalData:         mustGetGasData(data_polygon.GetGasData),
		},
		{
			Name:              "BNB Smart Chain",
			ChainID:           infura.BNB,
			RPC:               "https://bsc-mainnet.infura.io/v3/" + *infuraToken,
			SuggestionsConfig: bscConfig,
		},
		{
			Name:              "Linea",
			ChainID:           infura.Linea,
			RPC:               "https://linea-mainnet.infura.io/v3/" + *infuraToken,
			SuggestionsConfig: lineaConfig,
		},
		{
			Name:              "Status Network Sepolia",
			ChainID:           1660990954,
			RPC:               "https://public.sepolia.rpc.status.network",
			SuggestionsConfig: statusNetworkSepoliaConfig,
		},
	}

	// Test each network
	for i, network := range networks {
		fmt.Printf("\n%s %s (%d)\n", getNetworkEmoji(network.ChainID), network.Name, network.ChainID)
		fmt.Println(strings.Repeat("-", 60))

		var client GasDataClient
		if *fake {
			if network.LocalData == nil {
				fmt.Printf("❌ No local data found for %s\n", network.Name)
				continue
			}
			client = data.NewFakeClient(network.LocalData)
		} else {
			var err error
			client, err = NewRealClient(network, *infuraToken)
			if err != nil {
				fmt.Printf("❌ Error creating RPC client: %v\n", err)
				continue
			}
		}

		err := compareNetwork(network, client)
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

func displayComparison(current *gas.FeeSuggestions, oldFees *old.SuggestedFees, infura *infura.GasResponse, gasPrice *big.Int, gasTipCap *big.Int) {
	fmt.Println("\n📋 COMPARISON RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	oldFeesValid := oldFees != nil && oldFees.EIP1559Enabled
	infuraFeesValid := infura != nil

	// Use wei values directly instead of converting to gwei
	currentLowPriority := current.Low.MaxPriorityFeePerGas
	currentLowMaxFee := current.Low.MaxFeePerGas
	currentMediumPriority := current.Medium.MaxPriorityFeePerGas
	currentMediumMaxFee := current.Medium.MaxFeePerGas
	currentHighPriority := current.High.MaxPriorityFeePerGas
	currentHighMaxFee := current.High.MaxFeePerGas
	currentBaseFee := current.EstimatedBaseFee

	oldLowPriority := big.NewInt(0)
	oldLowMaxFee := big.NewInt(0)
	oldMediumPriority := big.NewInt(0)
	oldMediumMaxFee := big.NewInt(0)
	oldHighPriority := big.NewInt(0)
	oldHighMaxFee := big.NewInt(0)
	oldBaseFee := big.NewInt(0)
	if oldFeesValid {
		oldLowPriority = oldFees.MaxFeesLevels.LowPriority.ToInt()
		oldLowMaxFee = oldFees.MaxFeesLevels.Low.ToInt()
		oldMediumPriority = oldFees.MaxFeesLevels.MediumPriority.ToInt()
		oldMediumMaxFee = oldFees.MaxFeesLevels.Medium.ToInt()
		oldHighPriority = oldFees.MaxFeesLevels.HighPriority.ToInt()
		oldHighMaxFee = oldFees.MaxFeesLevels.High.ToInt()
		oldBaseFee = oldFees.CurrentBaseFee
	}

	infuraLowPriority := big.NewInt(0)
	infuraLowMaxFee := big.NewInt(0)
	infuraMediumPriority := big.NewInt(0)
	infuraMediumMaxFee := big.NewInt(0)
	infuraHighPriority := big.NewInt(0)
	infuraHighMaxFee := big.NewInt(0)
	infuraBaseFee := big.NewInt(0)
	if infuraFeesValid {
		infuraLowPriority = stringToWei(infura.Low.SuggestedMaxPriorityFeePerGas)
		infuraLowMaxFee = stringToWei(infura.Low.SuggestedMaxFeePerGas)
		infuraMediumPriority = stringToWei(infura.Medium.SuggestedMaxPriorityFeePerGas)
		infuraMediumMaxFee = stringToWei(infura.Medium.SuggestedMaxFeePerGas)
		infuraHighPriority = stringToWei(infura.High.SuggestedMaxPriorityFeePerGas)
		infuraHighMaxFee = stringToWei(infura.High.SuggestedMaxFeePerGas)
		infuraBaseFee = stringToWei(infura.EstimatedBaseFee)
	}

	fmt.Printf("\n")
	fmt.Printf("📋 NODE SUGGESTIONS\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("   Gas Price:              %20s wei\n", gasPrice.String())
	fmt.Printf("   Gas Tip Cap:            %20s wei\n", gasTipCap.String())
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("\n")

	fmt.Printf("🔸 BASE FEE\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentBaseFee.String())
	if oldFeesValid {
		fmt.Printf("   Old Implementation:     %20s wei\n", oldBaseFee.String())
		fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentBaseFee, oldBaseFee).String(),
			percentDiffWei(currentBaseFee, oldBaseFee))
	}
	if infuraFeesValid {
		fmt.Printf("   Infura:                 %20s wei\n", infuraBaseFee.String())
		fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentBaseFee, infuraBaseFee).String(),
			percentDiffWei(currentBaseFee, infuraBaseFee))
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 LOW PRIORITY FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentLowPriority.String())
	if oldFeesValid {
		fmt.Printf("   Old Implementation:     %20s wei\n", oldLowPriority.String())
		fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentLowPriority, oldLowPriority).String(),
			percentDiffWei(currentLowPriority, oldLowPriority))
	}
	if infuraFeesValid {
		fmt.Printf("   Infura:                 %20s wei\n", infuraLowPriority.String())
		fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentLowPriority, infuraLowPriority).String(),
			percentDiffWei(currentLowPriority, infuraLowPriority))
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 MEDIUM PRIORITY FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentMediumPriority.String())
	if oldFeesValid {
		fmt.Printf("   Old Implementation:     %20s wei\n", oldMediumPriority.String())
		fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentMediumPriority, oldMediumPriority).String(),
			percentDiffWei(currentMediumPriority, oldMediumPriority))
	}
	if infuraFeesValid {
		fmt.Printf("   Infura:                 %20s wei\n", infuraMediumPriority.String())
		fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentMediumPriority, infuraMediumPriority).String(),
			percentDiffWei(currentMediumPriority, infuraMediumPriority))
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 HIGH PRIORITY FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentHighPriority.String())
	if oldFeesValid {
		fmt.Printf("   Old Implementation:     %20s wei\n", oldHighPriority.String())
		fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentHighPriority, oldHighPriority).String(),
			percentDiffWei(currentHighPriority, oldHighPriority))
	}
	if infuraFeesValid {
		fmt.Printf("   Infura:                 %20s wei\n", infuraHighPriority.String())
		fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentHighPriority, infuraHighPriority).String(),
			percentDiffWei(currentHighPriority, infuraHighPriority))
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 LOW MAX FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentLowMaxFee.String())
	if oldFeesValid {
		fmt.Printf("   Old Implementation:     %20s wei\n", oldLowMaxFee.String())
		fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentLowMaxFee, oldLowMaxFee).String(),
			percentDiffWei(currentLowMaxFee, oldLowMaxFee))
	}
	if infuraFeesValid {
		fmt.Printf("   Infura:                 %20s wei\n", infuraLowMaxFee.String())
		fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentLowMaxFee, infuraLowMaxFee).String(),
			percentDiffWei(currentLowMaxFee, infuraLowMaxFee))
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 MEDIUM MAX FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentMediumMaxFee.String())
	if oldFeesValid {
		fmt.Printf("   Old Implementation:     %20s wei\n", oldMediumMaxFee.String())
		fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentMediumMaxFee, oldMediumMaxFee).String(),
			percentDiffWei(currentMediumMaxFee, oldMediumMaxFee))
	}
	if infuraFeesValid {
		fmt.Printf("   Infura:                 %20s wei\n", infuraMediumMaxFee.String())
		fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentMediumMaxFee, infuraMediumMaxFee).String(),
			percentDiffWei(currentMediumMaxFee, infuraMediumMaxFee))
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 HIGH MAX FEES\n")
	fmt.Printf("   Current Implementation: %20s wei\n", currentHighMaxFee.String())
	if oldFeesValid {
		fmt.Printf("   Old Implementation:     %20s wei\n", oldHighMaxFee.String())
		fmt.Printf("   Current vs Old:         %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentHighMaxFee, oldHighMaxFee).String(),
			percentDiffWei(currentHighMaxFee, oldHighMaxFee))
	}
	if infuraFeesValid {
		fmt.Printf("   Infura:                 %20s wei\n", infuraHighMaxFee.String())
		fmt.Printf("   Current vs Infura:      %+20s wei (%+.1f%%)\n",
			new(big.Int).Sub(currentHighMaxFee, infuraHighMaxFee).String(),
			percentDiffWei(currentHighMaxFee, infuraHighMaxFee))
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 LOW WAIT TIME\n")
	fmt.Printf("   Wait Time (Current):    %.1f-%.1f seconds\n",
		current.Low.MinTimeUntilInclusion, current.Low.MaxTimeUntilInclusion)
	if oldFeesValid {
		fmt.Printf("   Wait Time (Old):        %d seconds\n", oldFees.MaxFeesLevels.LowEstimatedTime)
	}
	if infuraFeesValid {
		fmt.Printf("   Wait Time (Infura):     %.1f-%.1f seconds\n",
			float64(infura.Low.MinWaitTimeEstimate)/1000, float64(infura.Low.MaxWaitTimeEstimate)/1000)
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 MEDIUM WAIT TIME\n")
	fmt.Printf("   Wait Time (Current):    %.1f-%.1f seconds\n",
		current.Medium.MinTimeUntilInclusion, current.Medium.MaxTimeUntilInclusion)
	if oldFeesValid {
		fmt.Printf("   Wait Time (Old):        %d seconds\n", oldFees.MaxFeesLevels.MediumEstimatedTime)
	}
	if infuraFeesValid {
		fmt.Printf("   Wait Time (Infura):     %.1f-%.1f seconds\n",
			float64(infura.Medium.MinWaitTimeEstimate)/1000, float64(infura.Medium.MaxWaitTimeEstimate)/1000)
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 HIGH WAIT TIME\n")
	fmt.Printf("   Wait Time (Current):    %.1f-%.1f seconds\n",
		current.High.MinTimeUntilInclusion, current.High.MaxTimeUntilInclusion)
	if oldFeesValid {
		fmt.Printf("   Wait Time (Old):        %d seconds\n", oldFees.MaxFeesLevels.HighEstimatedTime)
	}
	if infuraFeesValid {
		fmt.Printf("   Wait Time (Infura):     %.1f-%.1f seconds\n",
			float64(infura.High.MinWaitTimeEstimate)/1000, float64(infura.High.MaxWaitTimeEstimate)/1000)
	}
	fmt.Printf("\n")

	fmt.Printf("🔸 NETWORK CONGESTION\n")
	fmt.Printf("   Current Implementation: %.3f\n", current.NetworkCongestion)
	if infuraFeesValid {
		fmt.Printf("   Infura:                 %.3f\n", infura.NetworkCongestion)
		fmt.Printf("   Difference:             %+.3f\n", current.NetworkCongestion-infura.NetworkCongestion)
	}
	fmt.Printf("\n")
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
func compareNetwork(network NetworkInfo, client GasDataClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas tip cap: %w", err)
	}

	// Create a simple 0-valued ETH transfer call message for testing
	callMsg := &ethereum.CallMsg{
		From:  common.HexToAddress("0xd8da6bf26964af9d7eed9e03e53415d37aa96045"), // vitalik.eth
		To:    &common.Address{},                                                 // Zero address for ETH transfer
		Data:  []byte{},                                                          // No data for simple ETH transfer
		Value: big.NewInt(0),                                                     // 0 ETH value
	}

	// Get our suggestions using GetTxSuggestions
	fmt.Printf("📊 Fetching our gas suggestions for %s...\n", network.Name)
	txSuggestions, err := gas.GetTxSuggestions(ctx, client, network.SuggestionsConfig, callMsg)
	if err != nil {
		return fmt.Errorf("failed to get our suggestions: %w", err)
	}

	// Get old estimator suggestions
	fmt.Printf("📊 Fetching old estimator suggestions for %s...\n", network.Name)
	oldFeeManager := old.NewFeeManager(client)
	oldSuggestions, _, _, err := oldFeeManager.SuggestedFees(ctx, uint64(network.ChainID), old.ZeroAddress())
	if err != nil {
		fmt.Printf("⚠️  Old estimator not available for %s: %v\n", network.Name, err)
		oldSuggestions = nil
	}

	// Get Infura's suggestions for this network
	fmt.Printf("📊 Fetching Infura's gas suggestions for %s...\n", network.Name)
	infuraSuggestions, err := client.GetGasSuggestions(ctx, network.ChainID)
	if err != nil {
		fmt.Printf("⚠️  Infura API not available for %s: %v\n", network.Name, err)
	}

	if infuraSuggestions == nil && oldSuggestions == nil {
		fmt.Printf("📊 Showing only our implementation results:\n")
		displayOurSuggestions(txSuggestions.FeeSuggestions, gasPrice, gasTipCap)
	} else {
		// Display comparison
		fmt.Printf("📊 Showing our implementation vs old implementation:\n")
		displayComparison(txSuggestions.FeeSuggestions, oldSuggestions, infuraSuggestions, gasPrice, gasTipCap)
	}

	return nil
}

// displayOurSuggestions displays only our suggestions when Infura is not available
func displayOurSuggestions(suggestions *gas.FeeSuggestions, gasPrice *big.Int, gasTipCap *big.Int) {
	ourLow := suggestions.Low.MaxPriorityFeePerGas
	ourMedium := suggestions.Medium.MaxPriorityFeePerGas
	ourHigh := suggestions.High.MaxPriorityFeePerGas
	ourBaseFee := suggestions.EstimatedBaseFee

	// Use wei values directly for max fees
	ourLowMaxFee := suggestions.Low.MaxFeePerGas
	ourMediumMaxFee := suggestions.Medium.MaxFeePerGas
	ourHighMaxFee := suggestions.High.MaxFeePerGas

	fmt.Printf("📋 NODE SUGGESTIONS\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Gas Price: %20s wei\n", gasPrice.String())
	fmt.Printf("Gas Tip Cap: %20s wei\n", gasTipCap.String())
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════\n")

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
	fmt.Printf("Low            │ %8.0f │ %8.0f │ %7.1f │ %s\n",
		suggestions.Low.MinTimeUntilInclusion, suggestions.Low.MaxTimeUntilInclusion,
		(suggestions.Low.MinTimeUntilInclusion+suggestions.Low.MaxTimeUntilInclusion)/2,
		getTimeCategory(int(suggestions.Low.MinTimeUntilInclusion), int(suggestions.Low.MaxTimeUntilInclusion)))
	fmt.Printf("Medium         │ %8.0f │ %8.0f │ %7.1f │ %s\n",
		suggestions.Medium.MinTimeUntilInclusion, suggestions.Medium.MaxTimeUntilInclusion,
		(suggestions.Medium.MinTimeUntilInclusion+suggestions.Medium.MaxTimeUntilInclusion)/2,
		getTimeCategory(int(suggestions.Medium.MinTimeUntilInclusion), int(suggestions.Medium.MaxTimeUntilInclusion)))
	fmt.Printf("High           │ %8.0f │ %8.0f │ %7.1f │ %s\n",
		suggestions.High.MinTimeUntilInclusion, suggestions.High.MaxTimeUntilInclusion,
		(suggestions.High.MinTimeUntilInclusion+suggestions.High.MaxTimeUntilInclusion)/2,
		getTimeCategory(int(suggestions.High.MinTimeUntilInclusion), int(suggestions.High.MaxTimeUntilInclusion)))
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
