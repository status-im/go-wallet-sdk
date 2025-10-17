package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
)

func main() {
	// Define command line flags
	var (
		rpcURL = flag.String("rpc", "", "RPC URL for the blockchain network (required)")
		help   = flag.Bool("help", false, "Show help message")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Gas Data Generator - Fetches blockchain gas data and generates Go code\n\n")
		fmt.Fprintf(os.Stderr, "The chain ID is automatically detected from the RPC endpoint.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # BSC Mainnet\n")
		fmt.Fprintf(os.Stderr, "  %s -rpc https://bsc-mainnet.infura.io/v3/YOUR_API_KEY\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Ethereum Mainnet\n")
		fmt.Fprintf(os.Stderr, "  %s -rpc https://mainnet.infura.io/v3/YOUR_API_KEY\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Polygon Mainnet\n")
		fmt.Fprintf(os.Stderr, "  %s -rpc https://polygon-mainnet.infura.io/v3/YOUR_API_KEY\n\n", os.Args[0])
	}

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate required arguments
	if *rpcURL == "" {
		fmt.Fprintf(os.Stderr, "Error: -rpc flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Fetching gas data from network...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create RPC client
	rpcClient, err := rpc.Dial(*rpcURL)
	if err != nil {
		fmt.Printf("Error connecting to RPC: %v\n", err)
		os.Exit(1)
	}
	defer rpcClient.Close()

	// Create ethclient
	ethClient := ethclient.NewClient(rpcClient)

	// Create Infura client
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	infuraClient := infura.NewClient(httpClient)
	defer infuraClient.Close()

	// 1. Get chain ID
	fmt.Println("Detecting chain ID...")
	chainID, err := getChainID(ctx, ethClient)
	if err != nil {
		fmt.Printf("Error getting chain ID: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Detected chain ID: %d\n", chainID.Int64())

	// 2. Get latest block
	fmt.Println("Fetching latest block...")
	latestBlock, err := getLatestBlock(ctx, ethClient)
	if err != nil {
		fmt.Printf("Error fetching latest block: %v\n", err)
		os.Exit(1)
	}

	// 3. Get current gas price
	fmt.Println("Fetching current gas price...")
	gasPrice, err := ethClient.SuggestGasPrice(ctx)
	if err != nil {
		fmt.Printf("Error fetching gas price: %v\n", err)
		os.Exit(1)
	}

	// 4. Get max fee per gas
	fmt.Println("Fetching max priority fee per gas...")
	maxPriorityFeePerGas, err := ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		fmt.Printf("Error fetching max priority fee per gas: %v\n", err)
		os.Exit(1)
	}

	// 5. Get Infura suggested fees
	fmt.Println("Fetching Infura suggested fees...")
	infuraFees, err := getInfuraSuggestedFees(ctx, infuraClient, int(chainID.Int64()))
	if err != nil {
		fmt.Printf("Error fetching Infura suggested fees: %v\n", err)
		os.Exit(1)
	}

	// 6. Get fee history
	fmt.Println("Fetching fee history...")
	feeHistory, err := getFeeHistory(ctx, ethClient, latestBlock.Number)
	if err != nil {
		fmt.Printf("Error fetching fee history: %v\n", err)
		os.Exit(1)
	}

	// Generate the Go file with the data
	fmt.Println("Generating chain-specific data file...")
	err = generateDataFile(latestBlock, gasPrice, maxPriorityFeePerGas, feeHistory, infuraFees, int(chainID.Int64()))
	if err != nil {
		fmt.Printf("Error generating data file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully generated data file for chain %d!\n", chainID.Int64())
}

// getChainID fetches the chain ID using ethclient
func getChainID(ctx context.Context, client *ethclient.Client) (*big.Int, error) {
	return client.ChainID(ctx)
}

// getLatestBlock fetches the latest block using ethclient
func getLatestBlock(ctx context.Context, client *ethclient.Client) (*ethclient.BlockWithFullTxs, error) {
	return client.EthGetBlockByNumberWithFullTxs(ctx, nil)
}

// getFeeHistory fetches fee history using ethclient
func getFeeHistory(ctx context.Context, client *ethclient.Client, lastBlock *big.Int) (*ethereum.FeeHistory, error) {
	percentiles := []float64{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100}
	blockCount := uint64(1024) // 0x400 in decimal

	return client.FeeHistory(ctx, blockCount, lastBlock, percentiles)
}

// getInfuraSuggestedFees fetches suggested fees from Infura Gas API using the infura client
func getInfuraSuggestedFees(ctx context.Context, client *infura.Client, chainID int) (*infura.GasResponse, error) {
	return client.GetGasSuggestions(ctx, chainID)
}

// generateDataFile creates the data.go file with the fetched data
func generateDataFile(
	block *ethclient.BlockWithFullTxs,
	gasPrice *big.Int,
	maxPriorityFeePerGas *big.Int,
	feeHistory *ethereum.FeeHistory,
	infuraFees *infura.GasResponse,
	chainID int) error {
	// Marshal the data to JSON strings for embedding
	blockJSON, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}

	gasPriceJSON := []byte(hexutil.EncodeBig(gasPrice))
	maxPriorityFeePerGasJSON := []byte(hexutil.EncodeBig(maxPriorityFeePerGas))

	feeHistoryJSON, err := json.MarshalIndent(feeHistory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal fee history: %w", err)
	}

	infuraFeesJSON, err := json.MarshalIndent(infuraFees, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal infura fees: %w", err)
	}

	// Generate package name and network name
	packageName := getPackageName(chainID)
	networkName := getNetworkName(chainID)

	// Generate the Go file content
	content := fmt.Sprintf(`package %s

import (
	"encoding/json"
	"gas-comparison/data"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/gas/infura"
	"math/big"
)

// GetGasData returns the gas data fetched from %s (Chain ID: %d)
func GetGasData() (*data.GasData, error) {
	// Parse the embedded JSON data
	var block ethclient.BlockWithFullTxs
	err := json.Unmarshal([]byte(latestBlockJSON), &block)
	if err != nil {
		return nil, err
	}

	var gasPrice *big.Int
	gasPrice, err = hexutil.DecodeBig(gasPriceJSON)
	if err != nil {
		return nil, err
	}

	var maxPriorityFeePerGas *big.Int
	maxPriorityFeePerGas, err = hexutil.DecodeBig(maxPriorityFeePerGasJSON)
	if err != nil {
		return nil, err
	}

	var feeHistory ethereum.FeeHistory
	err = json.Unmarshal([]byte(feeHistoryJSON), &feeHistory)
	if err != nil {
		return nil, err
	}

	var infuraFees infura.GasResponse
	err = json.Unmarshal([]byte(infuraFeesJSON), &infuraFees)
	if err != nil {
		return nil, err
	}

	return &data.GasData{
		LatestBlock:          &block,
		GasPrice:             gasPrice,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		FeeHistory:           &feeHistory,
		InfuraSuggestedFees:  &infuraFees,
	}, nil
}

// Embedded JSON data
const latestBlockJSON = `+"`%s`"+`

const gasPriceJSON = `+"`%s`"+`

const maxPriorityFeePerGasJSON = `+"`%s`"+`

const feeHistoryJSON = `+"`%s`"+`

const infuraFeesJSON = `+"`%s`"+`
`, packageName, networkName, chainID, string(blockJSON), string(gasPriceJSON), string(maxPriorityFeePerGasJSON), string(feeHistoryJSON), string(infuraFeesJSON))

	// Create directory for the chain-specific package (one level up from generator)
	dirPath := filepath.Join("..", packageName)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	// Write the file to the chain-specific directory
	filePath := filepath.Join(dirPath, "data.go")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	fmt.Printf("Generated file: %s\n", filePath)
	return nil
}

// getPackageName returns a valid Go package name for the given chain ID
func getPackageName(chainID int) string {
	switch chainID {
	case 1:
		return "ethereum"
	case 56:
		return "bsc"
	case 137:
		return "polygon"
	case 42161:
		return "arbitrum"
	case 10:
		return "optimism"
	case 8453:
		return "base"
	case 43114:
		return "avalanche"
	case 250:
		return "fantom"
	case 100:
		return "gnosis"
	case 25:
		return "cronos"
	case 11155420:
		return "optimismsepolia"
	case 421614:
		return "arbitrumsepolia"
	case 84532:
		return "basesepolia"
	case 11155111:
		return "sepolia"
	default:
		return fmt.Sprintf("chain%d", chainID)
	}
}

// getNetworkName returns a human-readable network name for the given chain ID
func getNetworkName(chainID int) string {
	switch chainID {
	case 1:
		return "Ethereum Mainnet"
	case 56:
		return "BSC Mainnet"
	case 137:
		return "Polygon Mainnet"
	case 42161:
		return "Arbitrum One"
	case 10:
		return "Optimism Mainnet"
	case 8453:
		return "Base Mainnet"
	case 43114:
		return "Avalanche C-Chain"
	case 250:
		return "Fantom Opera"
	case 100:
		return "Gnosis Chain"
	case 25:
		return "Cronos Mainnet"
	case 11155420:
		return "Optimism Sepolia"
	case 421614:
		return "Arbitrum Sepolia"
	case 84532:
		return "Base Sepolia"
	case 11155111:
		return "Ethereum Sepolia"
	default:
		return fmt.Sprintf("Chain %d", chainID)
	}
}
