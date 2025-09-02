package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

func main() {
	// Get RPC endpoint from environment or use default
	rpcEndpoints := strings.Fields(os.Getenv("ETH_RPC_ENDPOINTS"))
	if len(rpcEndpoints) == 0 {
		rpcEndpoints = []string{
			"https://ethereum-rpc.publicnode.com",
			"https://optimism-rpc.publicnode.com",
			"https://arbitrum-rpc.publicnode.com",
			"https://public.sepolia.rpc.status.network",
		}
	}

	for _, rpcEndpoint := range rpcEndpoints {
		testRPC(rpcEndpoint)
	}

	fmt.Println("\n✅ Example completed successfully!")
}

func testRPC(rpcEndpoint string) {
	fmt.Println("Testing RPC endpoint: ", rpcEndpoint)

	// Create RPC client
	rpcClient, err := rpc.Dial(rpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to dial RPC: %v", err)
	}
	defer rpcClient.Close()

	// Create our client
	client := ethclient.NewClient(rpcClient)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("🚀 Ethereum JSON-RPC Client Example (using eth.go methods)")
	fmt.Println("==========================================================")

	zeroAddress := common.Address{}

	// Example 1: Get basic network information
	fmt.Println("\n📡 Network Information")

	// Get client version
	version, err := client.Web3ClientVersion(ctx)
	if err != nil {
		log.Printf("Error getting client version: %v", err)
	} else {
		fmt.Printf("Client Version: %s\n", version)
	}

	// Get network ID
	networkID, err := client.NetVersion(ctx)
	if err != nil {
		log.Printf("Error getting network ID: %v", err)
	} else {
		fmt.Printf("Network ID: %s\n", networkID)
	}

	// Get chain ID
	chainID, err := client.EthChainId(ctx)
	if err != nil {
		log.Printf("Error getting chain ID: %v", err)
	} else {
		fmt.Printf("Chain ID: %s\n", chainID.String())
	}

	// Example 2: Get blockchain information
	fmt.Println("\n⛓️  Blockchain Information")

	// Get latest block number
	blockNumber, err := client.EthBlockNumber(ctx)
	if err != nil {
		log.Printf("Error getting block number: %v", err)
	} else {
		fmt.Printf("Latest Block Number: %d\n", blockNumber)
	}

	// Get latest block
	firstTxHash := common.Hash{}
	block, err := client.EthGetBlockByNumberWithFullTxs(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		log.Printf("Error getting latest block: %v", err)
	} else {
		fmt.Printf("Latest Block Hash: %s\n", block.Hash.Hex())
		fmt.Printf("Block Number: %d\n", block.Number)
		fmt.Printf("Found %d Transactions\n", len(block.Transactions))
		for i, tx := range block.Transactions {
			if i == 0 {
				firstTxHash = tx.Hash
			}
			if i >= 3 { // Show only first 3
				break
			}
			fmt.Printf("Transaction %d:\n", i+1)
			fmt.Printf(" Hash: %s\n", tx.Hash.Hex())
			fmt.Printf(" From: %s\n", tx.From.Hex())
			fmt.Printf(" Gas: %d\n", tx.Gas)
		}
		fmt.Printf("Block Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Gas Used: %d\n", block.GasUsed)
		fmt.Printf("Gas Limit: %d\n", block.GasLimit)
		if block.BaseFeePerGas != nil {
			fmt.Printf("Base Fee Per Gas: %s wei\n", block.BaseFeePerGas.String())
		}
	}

	// Get gas price
	gasPrice, err := client.EthGasPrice(ctx)
	if err != nil {
		log.Printf("Error getting gas price: %v", err)
	} else {
		fmt.Printf("Current Gas Price: %s wei\n", gasPrice.String())
	}

	// Example 3: Account information
	fmt.Println("\n👤 Account Information")

	// Example address (Vitalik's address)
	address := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")

	// Get balance
	balance, err := client.EthGetBalance(ctx, address, nil)
	if err != nil {
		log.Printf("Error getting balance: %v", err)
	} else {
		fmt.Printf("Balance of %s: %s wei\n", address.Hex(), balance.String())
		// Convert to ETH
		ethBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(big.NewInt(1e18)))
		fmt.Printf("Balance in ETH: %f\n", ethBalance)
	}

	// Get nonce
	nonce, err := client.EthGetTransactionCount(ctx, address, nil)
	if err != nil {
		log.Printf("Error getting nonce: %v", err)
	} else {
		fmt.Printf("Nonce of %s: %d\n", address.Hex(), nonce)
	}

	// Example 4: Contract interaction
	fmt.Println("\n📄 Contract Interaction")

	// Multicall3 contract address
	multicall3Address := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Get contract code
	code, err := client.EthGetCode(ctx, multicall3Address, nil)
	if err != nil {
		log.Printf("Error getting contract code: %v", err)
	} else {
		fmt.Printf("Multicall3 Contract Code Length: %d bytes\n", len(code))
		if len(code) > 0 {
			fmt.Println("✅ Contract exists (has code)")
		} else {
			fmt.Println("❌ No contract at address")
		}
	}

	// Example 5: Event filtering
	fmt.Println("\n🔍 Event Filtering")

	// Create a filter for Transfer events
	// Transfer event signature: Transfer(address,address,uint256)
	transferEventSig := common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	filterQuery := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(blockNumber - 10)), // Last 10 blocks
		ToBlock:   big.NewInt(int64(blockNumber)),
		Addresses: []common.Address{common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")},
		Topics: [][]common.Hash{
			{transferEventSig}, // Event signature
		},
	}

	// Get logs directly
	logs, err := client.EthGetLogs(ctx, filterQuery)
	if err != nil {
		log.Printf("Error getting logs: %v", err)
	} else {
		fmt.Printf("Found %d Transfer events in the last 10 blocks\n", len(logs))

		// Show first few events
		for i, log := range logs {
			if i >= 3 { // Show only first 3
				break
			}
			fmt.Printf("  Event %d: Block %d, Tx %s\n",
				i+1,
				log.BlockNumber,
				log.TxHash.Hex())
		}
	}

	// Example 6: Transaction information
	fmt.Println("\n💸 Transaction Information")

	// Get transaction
	tx, err := client.EthGetTransactionByHash(ctx, firstTxHash)
	if err != nil {
		fmt.Printf("Error getting transaction: %v\n", err)
	} else if tx == nil {
		fmt.Println("Transaction not found")
	} else {
		fmt.Printf("Transaction Hash: %s\n", tx.Hash.Hex())
		fmt.Printf("From: %s\n", tx.From.Hex())
		if tx.To != nil {
			fmt.Printf("To: %s\n", tx.To.Hex())
		}
		fmt.Printf("Value: %s wei\n", tx.Value.String())
		fmt.Printf("Gas: %d\n", tx.Gas)
		if tx.GasPrice != nil {
			fmt.Printf("Gas Price: %s wei\n", tx.GasPrice.String())
		}
	}

	// Example 7: Network status
	fmt.Println("\n🌐 Network Status")

	// Check net version
	netVersion, err := client.NetVersion(ctx)
	if err != nil {
		log.Printf("Error checking net version: %v", err)
	} else {
		fmt.Printf("Net Version: %v\n", netVersion)
	}

	// Example 8: Gas estimation
	fmt.Println("\n⛽ Gas Estimation")

	// Create a call message for gas estimation
	callMsg := ethereum.CallMsg{
		From:  zeroAddress,
		To:    &zeroAddress,
		Value: big.NewInt(0),
	}

	// Estimate gas
	estimatedGas, err := client.EthEstimateGas(ctx, callMsg)
	if err != nil {
		log.Printf("Error estimating gas: %v", err)
	} else {
		fmt.Printf("Estimated gas for call: %d\n", estimatedGas)
	}

	fmt.Println("--------------------------------")
	fmt.Println("--------------------------------")
	fmt.Println("--------------------------------")
}
