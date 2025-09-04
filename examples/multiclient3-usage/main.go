package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
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

	// Create multicall3 contract instance
	multicallContract, err := multicall3.NewMulticall3(multicallAddr, client)
	if err != nil {
		log.Fatalf("Failed to create multicall3 contract instance: %v", err)
	}

	// Execute multicall for batch balance reads
	ctx := context.Background()
	callOpts := &bind.CallOpts{Context: ctx}

	// Prepare calls for multicall3 contract
	calls := prepareMulticall3Calls(multicallAddr, walletAddress)
	multicall3ResultsCount := len(calls)

	// Prepare calls for balanceOf function
	calls = append(calls, prepareBalanceCalls(chainTokens, walletAddress)...)

	log.Printf("Prepared %d balance calls for multicall", len(calls))

	// Execute multicall using the view method
	results, err := multicallContract.ViewAggregate3(callOpts, calls)
	if err != nil {
		log.Fatalf("Failed to execute multicall: %v", err)
	}

	// Process results
	blockNumber, nativeBalance := processMulticall3Results(results[:multicall3ResultsCount])
	balanceResults := processBalanceResults(chainTokens, results[multicall3ResultsCount:])

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

func prepareMulticall3Calls(contractAddress common.Address, walletAddress string) []multicall3.IMulticall3Call3 {
	var calls []multicall3.IMulticall3Call3

	abi, err := multicall3.Multicall3MetaData.GetAbi()
	if err != nil {
		log.Fatalf("Failed to get Multicall3 ABI: %v", err)
	}

	if callData, err := abi.Pack("getBlockNumber"); err != nil {
		log.Fatalf("Failed to pack getBlockNumber call data: %v", err)
	} else {
		calls = append(calls, multicall3.IMulticall3Call3{
			Target:       contractAddress,
			AllowFailure: true,
			CallData:     callData,
		})
	}

	if callData, err := abi.Pack("getEthBalance", common.HexToAddress(walletAddress)); err != nil {
		log.Fatalf("Failed to pack getEthBalance call data: %v", err)
	} else {
		calls = append(calls, multicall3.IMulticall3Call3{
			Target:       contractAddress,
			AllowFailure: true,
			CallData:     callData,
		})
	}

	return calls
}

func prepareBalanceCalls(tokens []Token, walletAddress string) []multicall3.IMulticall3Call3 {
	var calls []multicall3.IMulticall3Call3

	abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		log.Fatalf("Failed to get Multicall3 ABI: %v", err)
	}

	callData, err := abi.Pack("balanceOf", common.HexToAddress(walletAddress))
	if err != nil {
		log.Fatalf("Failed to pack balanceOf call data: %v", err)
	}

	for _, token := range tokens {
		tokenAddr := common.HexToAddress(token.Address)

		call := multicall3.IMulticall3Call3{
			Target:       tokenAddr,
			AllowFailure: true, // Allow individual calls to fail
			CallData:     callData,
		}

		calls = append(calls, call)
	}

	return calls
}

func processMulticall3Results(results []multicall3.IMulticall3Result) (blockNumber *big.Int, nativeBalance *big.Int) {
	if len(results) != 2 {
		log.Fatalf("Expected 2 results, got %d", len(results))
	}

	for i, result := range results {
		switch i {
		case 0:
			blockNumber = new(big.Int).SetBytes(result.ReturnData)
		case 1:
			nativeBalance = new(big.Int).SetBytes(result.ReturnData)
		}
	}
	return
}

func processBalanceResults(tokens []Token, results []multicall3.IMulticall3Result) []BalanceResult {
	var balanceResults []BalanceResult

	for i, result := range results {
		if i >= len(tokens) {
			break
		}

		token := tokens[i]
		balanceResult := BalanceResult{
			Token:   token,
			Success: result.Success,
		}

		if result.Success && len(result.ReturnData) >= 32 {
			// Parse the balance from return data
			balance := new(big.Int).SetBytes(result.ReturnData)
			balanceResult.Balance = balance
		} else {
			balanceResult.Balance = big.NewInt(0)
		}

		balanceResults = append(balanceResults, balanceResult)
	}

	return balanceResults
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
