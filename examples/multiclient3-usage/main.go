package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	"github.com/status-im/go-wallet-sdk/pkg/multicall"
)

// Token represents a token from the all.json file
type Token struct {
	ChainID  int64  `json:"chainId"`
	Address  string `json:"address"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	LogoURI  string `json:"logoURI"`
}

// TokenList represents the structure of all.json
type TokenList struct {
	Name      string   `json:"name"`
	LogoURI   string   `json:"logoURI"`
	Keywords  []string `json:"keywords"`
	Timestamp string   `json:"timestamp"`
	Tokens    []Token  `json:"tokens"`
}

// BalanceResult represents the result of a balance query
type BalanceResult struct {
	Token   Token
	Balance *big.Int
	Success bool
}

func main() {
	// Configuration
	chainID := int64(1)                                           // Ethereum mainnet
	walletAddress := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045" // vitalik.eth

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		log.Fatalf("RPC_URL environment variable is not set")
	}

	// Load token list
	tokens, err := loadTokenList("all.json")
	if err != nil {
		log.Fatalf("Failed to load token list: %v", err)
	}

	// Filter tokens for the specified chain
	var chainTokens []Token
	for _, token := range tokens {
		if token.ChainID == chainID {
			chainTokens = append(chainTokens, token)
		}
	}

	if len(chainTokens) == 0 {
		log.Printf("No tokens found for chain ID %d", chainID)
		return
	}

	log.Printf("Found %d tokens for chain ID %d", len(chainTokens), chainID)

	// Connect to Ethereum client
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}
	defer client.Close()

	// Get multicall3 contract address
	multicallAddr, exists := multicall3.GetMulticall3Address(chainID)
	if !exists {
		log.Fatalf("Multicall3 not supported on chain ID %d", chainID)
	}

	log.Printf("Using Multicall3 contract at: %s", multicallAddr.Hex())

	// Create multicall3 contract instance for the caller interface
	multicallContract, err := multicall3.NewMulticall3(multicallAddr, client)
	if err != nil {
		log.Fatalf("Failed to create multicall3 contract instance: %v", err)
	}

	// Execute multicall for batch balance reads
	ctx := context.Background()

	// Prepare calls using the new multicall package
	calls := prepareMulticallCalls(multicallAddr, walletAddress, chainTokens)

	log.Printf("Prepared %d balance calls for multicall", len(calls))

	// Execute multicall using the new RunSync function
	results := multicall.RunSync(ctx, [][]multicall3.IMulticall3Call{calls}, nil, multicallContract, 100)

	// Process results
	blockNumber := results[0].BlockNumber
	nativeBalance, balanceResults := processResults(chainTokens, results[0])

	// Display results
	displayResults(blockNumber, nativeBalance, balanceResults, walletAddress)
}

func loadTokenList(filename string) ([]Token, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var tokenList TokenList
	if err := json.Unmarshal(data, &tokenList); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return tokenList.Tokens, nil
}

func prepareMulticallCalls(contractAddress common.Address, walletAddress string, tokens []Token) []multicall3.IMulticall3Call {
	var calls []multicall3.IMulticall3Call

	// Add native ETH balance call
	calls = append(calls, multicall.BuildNativeBalanceCall(common.HexToAddress(walletAddress), contractAddress))

	// Add ERC20 balance calls for each token
	for _, token := range tokens {
		calls = append(calls, multicall.BuildERC20BalanceCall(common.HexToAddress(walletAddress), common.HexToAddress(token.Address)))
	}

	return calls
}

func processResults(tokens []Token, jobResult multicall.JobResult) (*big.Int, []BalanceResult) {
	var balanceResults []BalanceResult

	// Check for errors
	if jobResult.Err != nil {
		log.Fatalf("Multicall execution failed: %v", jobResult.Err)
	}

	results := jobResult.Results
	if len(results) == 0 {
		log.Fatalf("No results returned from multicall")
	}

	// First result is the native ETH balance
	nativeBalance, err := multicall.ProcessNativeBalanceResult(results[0])
	if err != nil {
		log.Printf("Failed to process native balance: %v", err)
		nativeBalance = big.NewInt(0)
	}

	// Process ERC20 token balances
	for i, result := range results[1:] {
		if i >= len(tokens) {
			break
		}

		token := tokens[i]
		balanceResult := BalanceResult{
			Token:   token,
			Success: result.Success,
		}

		if result.Success {
			balance, err := multicall.ProcessERC20BalanceResult(result)
			if err != nil {
				log.Printf("Failed to process balance for token %s: %v", token.Symbol, err)
				balanceResult.Balance = big.NewInt(0)
			} else {
				balanceResult.Balance = balance
			}
		} else {
			balanceResult.Balance = big.NewInt(0)
		}

		balanceResults = append(balanceResults, balanceResult)
	}

	return nativeBalance, balanceResults
}

func displayResults(blockNumber *big.Int, nativeBalance *big.Int, results []BalanceResult, walletAddress string) {
	fmt.Printf("\n=== Multicall3 Results ===\n")
	fmt.Printf("Block Number: %d\n", blockNumber)

	// Convert balance to human readable format
	decimals := 18
	balanceFloat := new(big.Float).SetInt(nativeBalance)
	divisor := new(big.Float).SetFloat64(10.0)
	for i := 0; i < decimals; i++ {
		balanceFloat.Quo(balanceFloat, divisor)
	}
	fmt.Printf("\n✅ %s (%s): %s\n",
		"ETH",
		"Ethereum",
		balanceFloat.Text('f', decimals))

	fmt.Printf("\n=== ERC20 Token Balances for %s ===\n\n", walletAddress)

	var totalTokens int
	var successfulCalls int
	var nonZeroBalances int

	for _, result := range results {
		totalTokens++

		if result.Success {
			successfulCalls++
		}

		if result.Balance.Cmp(big.NewInt(0)) > 0 {
			nonZeroBalances++

			// Convert balance to human readable format
			decimals := result.Token.Decimals
			balanceFloat := new(big.Float).SetInt(result.Balance)
			divisor := new(big.Float).SetFloat64(10.0)
			for i := 0; i < decimals; i++ {
				balanceFloat.Quo(balanceFloat, divisor)
			}

			fmt.Printf("✅ %s (%s): %s\n",
				result.Token.Symbol,
				result.Token.Name,
				balanceFloat.Text('f', decimals))
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total tokens queried: %d\n", totalTokens)
	fmt.Printf("Successful calls: %d\n", successfulCalls)
	fmt.Printf("Tokens with non-zero balance: %d\n", nonZeroBalances)
	fmt.Printf("Success rate: %.2f%%\n", float64(successfulCalls)/float64(totalTokens)*100)
}
