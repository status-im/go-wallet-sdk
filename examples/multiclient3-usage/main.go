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

	// Prepare jobs using the new multicall package
	jobs := prepareMulticallJobs(multicallAddr, walletAddress, chainTokens)

	log.Printf("Prepared %d jobs for multicall", len(jobs))

	// Execute multicall using the new RunSync function
	results := multicall.RunSync(ctx, jobs, nil, multicallContract, 100)

	// Process results
	blockNumber := results[0].BlockNumber
	nativeBalance, balanceResults := processResults(chainTokens, results[0], results[1])

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

func prepareMulticallJobs(contractAddress common.Address, walletAddress string, tokens []Token) []multicall.Job {
	var jobs []multicall.Job

	// Create native ETH balance job
	nativeCalls := []multicall3.IMulticall3Call{
		multicall.BuildNativeBalanceCall(common.HexToAddress(walletAddress), contractAddress),
	}
	nativeJob := multicall.Job{
		Calls: nativeCalls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return multicall.ProcessNativeBalanceResult(result)
		},
	}
	jobs = append(jobs, nativeJob)

	// Create ERC20 balance job
	erc20Calls := make([]multicall3.IMulticall3Call, 0, len(tokens))
	for _, token := range tokens {
		erc20Calls = append(erc20Calls, multicall.BuildERC20BalanceCall(common.HexToAddress(walletAddress), common.HexToAddress(token.Address)))
	}
	erc20Job := multicall.Job{
		Calls: erc20Calls,
		CallResultFn: func(result multicall3.IMulticall3Result) (any, error) {
			return multicall.ProcessERC20BalanceResult(result)
		},
	}
	jobs = append(jobs, erc20Job)

	return jobs
}

func processResults(tokens []Token, nativeJobResult multicall.JobResult, erc20JobResult multicall.JobResult) (*big.Int, []BalanceResult) {
	var balanceResults []BalanceResult

	// Process native ETH balance
	var nativeBalance *big.Int
	if nativeJobResult.Err != nil {
		log.Printf("Failed to process native balance job: %v", nativeJobResult.Err)
		nativeBalance = big.NewInt(0)
	} else if len(nativeJobResult.Results) == 0 {
		log.Printf("No results returned from native balance job")
		nativeBalance = big.NewInt(0)
	} else {
		callResult := nativeJobResult.Results[0]
		if callResult.Err != nil {
			log.Printf("Failed to process native balance: %v", callResult.Err)
			nativeBalance = big.NewInt(0)
		} else {
			balance, ok := callResult.Value.(*big.Int)
			if !ok || balance == nil {
				log.Printf("Failed to parse native balance result")
				nativeBalance = big.NewInt(0)
			} else {
				nativeBalance = balance
			}
		}
	}

	// Process ERC20 token balances
	if erc20JobResult.Err != nil {
		log.Printf("Failed to process ERC20 balance job: %v", erc20JobResult.Err)
		// Create empty results for all tokens
		for _, token := range tokens {
			balanceResults = append(balanceResults, BalanceResult{
				Token:   token,
				Balance: big.NewInt(0),
				Success: false,
			})
		}
		return nativeBalance, balanceResults
	}

	erc20Results := erc20JobResult.Results
	if len(erc20Results) != len(tokens) {
		log.Printf("Expected %d ERC20 results, got %d", len(tokens), len(erc20Results))
		// Create empty results for all tokens
		for _, token := range tokens {
			balanceResults = append(balanceResults, BalanceResult{
				Token:   token,
				Balance: big.NewInt(0),
				Success: false,
			})
		}
		return nativeBalance, balanceResults
	}

	// Process each token balance
	for i, callResult := range erc20Results {
		token := tokens[i]
		balanceResult := BalanceResult{
			Token: token,
		}

		if callResult.Err != nil {
			log.Printf("Failed to process balance for token %s: %v", token.Symbol, callResult.Err)
			balanceResult.Balance = big.NewInt(0)
			balanceResult.Success = false
		} else {
			balance, ok := callResult.Value.(*big.Int)
			if !ok || balance == nil {
				log.Printf("Failed to parse balance for token %s", token.Symbol)
				balanceResult.Balance = big.NewInt(0)
				balanceResult.Success = false
			} else {
				balanceResult.Balance = balance
				balanceResult.Success = true
			}
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
