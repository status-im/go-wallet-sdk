package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc1155"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc721"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	"github.com/status-im/go-wallet-sdk/pkg/eventfilter"
	"github.com/status-im/go-wallet-sdk/pkg/eventlog"
)

// displayEventMetadata displays common event metadata from the Raw log
func displayEventMetadata(raw interface{}) {
	switch r := raw.(type) {
	case erc20.Erc20Transfer:
		fmt.Printf("   Event Signature: %s\n", r.Raw.Topics[0].Hex())
		fmt.Printf("   Data Length: %d bytes\n", len(r.Raw.Data))
		fmt.Printf("   Removed: %t\n", r.Raw.Removed)
	case erc721.Erc721Transfer:
		fmt.Printf("   Event Signature: %s\n", r.Raw.Topics[0].Hex())
		fmt.Printf("   Data Length: %d bytes\n", len(r.Raw.Data))
		fmt.Printf("   Removed: %t\n", r.Raw.Removed)
	case erc1155.Erc1155TransferSingle:
		fmt.Printf("   Event Signature: %s\n", r.Raw.Topics[0].Hex())
		fmt.Printf("   Data Length: %d bytes\n", len(r.Raw.Data))
		fmt.Printf("   Removed: %t\n", r.Raw.Removed)
	case erc1155.Erc1155TransferBatch:
		fmt.Printf("   Event Signature: %s\n", r.Raw.Topics[0].Hex())
		fmt.Printf("   Data Length: %d bytes\n", len(r.Raw.Data))
		fmt.Printf("   Removed: %t\n", r.Raw.Removed)
	}
}

// formatAddress shortens an address for display
func formatAddress(addr common.Address) string {
	addrStr := addr.Hex()
	if len(addrStr) > 10 {
		return addrStr[:6] + "..." + addrStr[len(addrStr)-4:]
	}
	return addrStr
}

// formatBigInt formats a big.Int for display
func formatBigInt(value *big.Int) string {
	if value.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}
	// For very large numbers, show in scientific notation
	if value.BitLen() > 64 {
		return fmt.Sprintf("%.2e", new(big.Float).SetInt(value))
	}
	return value.String()
}

func main() {
	// Parse command line arguments
	var (
		rpcURL     = flag.String("rpc", "https://mainnet.infura.io/v3/YOUR_KEY", "Ethereum RPC URL")
		account    = flag.String("account", "", "Account address to filter transfers for (required)")
		startBlock = flag.Int64("start", 0, "Start block number (required)")
		endBlock   = flag.Int64("end", 0, "End block number (required)")
		direction  = flag.String("direction", "both", "Transfer direction: send, receive, or both")
		help       = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	// Show help if requested
	if *help {
		showHelp()
		return
	}

	// Validate required arguments
	if *account == "" || *startBlock == 0 || *endBlock == 0 {
		fmt.Fprintf(os.Stderr, "Error: account, start, and end block are required\n\n")
		showHelp()
		os.Exit(1)
	}

	// Validate account address
	if !common.IsHexAddress(*account) {
		fmt.Fprintf(os.Stderr, "Error: invalid account address: %s\n", *account)
		os.Exit(1)
	}

	// Validate block numbers
	if *startBlock >= *endBlock {
		fmt.Fprintf(os.Stderr, "Error: start block (%d) must be less than end block (%d)\n", *startBlock, *endBlock)
		os.Exit(1)
	}

	// Parse direction
	var dir eventfilter.Direction
	switch strings.ToLower(*direction) {
	case "send":
		dir = eventfilter.Send
	case "receive":
		dir = eventfilter.Receive
	case "both":
		dir = eventfilter.Both
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid direction '%s'. Must be 'send', 'receive', or 'both'\n", *direction)
		os.Exit(1)
	}

	// Connect to Ethereum client
	fmt.Printf("Connecting to Ethereum RPC: %s\n", *rpcURL)
	rpcClient, err := rpc.Dial(*rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}
	defer rpcClient.Close()

	client := ethclient.NewClient(rpcClient)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}
	defer client.Close()

	// Get latest block number for context
	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Failed to get latest block number: %v", err)
	}

	fmt.Printf("Latest block: %d\n", latestBlock)
	fmt.Printf("Scanning blocks %d to %d for account %s\n", *startBlock, *endBlock, *account)
	fmt.Printf("Direction: %s\n\n", strings.Title(*direction))

	// Create filter configuration
	config := eventfilter.TransferQueryConfig{
		FromBlock: big.NewInt(*startBlock),
		ToBlock:   big.NewInt(*endBlock),
		Accounts:  []common.Address{common.HexToAddress(*account)},
		TransferTypes: []eventfilter.TransferType{
			eventfilter.TransferTypeERC20,
			eventfilter.TransferTypeERC721,
			eventfilter.TransferTypeERC1155,
		},
		Direction: dir,
	}

	// Filter transfer events with concurrent processing
	fmt.Println("Filtering transfer events...")
	events, err := eventfilter.FilterTransfers(context.Background(), client, config)
	if err != nil {
		log.Fatalf("Failed to filter events: %v", err)
	}

	// Display results
	fmt.Printf("\nFound %d transfer events:\n\n", len(events))
	if len(events) == 0 {
		fmt.Println("No transfer events found for the specified criteria.")
		return
	}

	// Group events by type for better display
	erc20Events := make([]eventlog.Event, 0)
	erc721Events := make([]eventlog.Event, 0)
	erc1155Events := make([]eventlog.Event, 0)

	for _, event := range events {
		switch event.EventKey {
		case eventlog.ERC20Transfer:
			erc20Events = append(erc20Events, event)
		case eventlog.ERC721Transfer:
			erc721Events = append(erc721Events, event)
		case eventlog.ERC1155TransferSingle, eventlog.ERC1155TransferBatch:
			erc1155Events = append(erc1155Events, event)
		}
	}

	// Display ERC20 transfers
	if len(erc20Events) > 0 {
		fmt.Printf("=== ERC20 Transfers (%d) ===\n", len(erc20Events))
		for i, event := range erc20Events {
			if transfer, ok := event.Unpacked.(erc20.Erc20Transfer); ok {
				fmt.Printf("%d. ERC20 Transfer\n", i+1)
				fmt.Printf("   Block: %d\n", transfer.Raw.BlockNumber)
				fmt.Printf("   Transaction: %s\n", transfer.Raw.TxHash.Hex())
				fmt.Printf("   From: %s (%s)\n", transfer.From.Hex(), formatAddress(transfer.From))
				fmt.Printf("   To: %s (%s)\n", transfer.To.Hex(), formatAddress(transfer.To))
				fmt.Printf("   Amount: %s\n", formatBigInt(transfer.Value))
				fmt.Printf("   Token Contract: %s (%s)\n", transfer.Raw.Address.Hex(), formatAddress(transfer.Raw.Address))
				fmt.Printf("   Log Index: %d\n", transfer.Raw.Index)
				fmt.Printf("   Topics: %d\n", len(transfer.Raw.Topics))
				displayEventMetadata(transfer)
				fmt.Println()
			}
		}
	}

	// Display ERC721 transfers
	if len(erc721Events) > 0 {
		fmt.Printf("=== ERC721 Transfers (%d) ===\n", len(erc721Events))
		for i, event := range erc721Events {
			if transfer, ok := event.Unpacked.(erc721.Erc721Transfer); ok {
				fmt.Printf("%d. ERC721 Transfer (NFT)\n", i+1)
				fmt.Printf("   Block: %d\n", transfer.Raw.BlockNumber)
				fmt.Printf("   Transaction: %s\n", transfer.Raw.TxHash.Hex())
				fmt.Printf("   From: %s (%s)\n", transfer.From.Hex(), formatAddress(transfer.From))
				fmt.Printf("   To: %s (%s)\n", transfer.To.Hex(), formatAddress(transfer.To))
				fmt.Printf("   Token ID: %s\n", formatBigInt(transfer.TokenId))
				fmt.Printf("   NFT Contract: %s (%s)\n", transfer.Raw.Address.Hex(), formatAddress(transfer.Raw.Address))
				fmt.Printf("   Log Index: %d\n", transfer.Raw.Index)
				fmt.Printf("   Topics: %d\n", len(transfer.Raw.Topics))
				displayEventMetadata(transfer)
				fmt.Println()
			}
		}
	}

	// Display ERC1155 transfers
	if len(erc1155Events) > 0 {
		fmt.Printf("=== ERC1155 Transfers (%d) ===\n", len(erc1155Events))
		for i, event := range erc1155Events {
			if event.EventKey == eventlog.ERC1155TransferSingle {
				if transfer, ok := event.Unpacked.(erc1155.Erc1155TransferSingle); ok {
					fmt.Printf("%d. ERC1155 Transfer Single\n", i+1)
					fmt.Printf("   Block: %d\n", transfer.Raw.BlockNumber)
					fmt.Printf("   Transaction: %s\n", transfer.Raw.TxHash.Hex())
					fmt.Printf("   Operator: %s (%s)\n", transfer.Operator.Hex(), formatAddress(transfer.Operator))
					fmt.Printf("   From: %s (%s)\n", transfer.From.Hex(), formatAddress(transfer.From))
					fmt.Printf("   To: %s (%s)\n", transfer.To.Hex(), formatAddress(transfer.To))
					fmt.Printf("   Token ID: %s\n", formatBigInt(transfer.Id))
					fmt.Printf("   Amount: %s\n", formatBigInt(transfer.Value))
					fmt.Printf("   Contract: %s (%s)\n", transfer.Raw.Address.Hex(), formatAddress(transfer.Raw.Address))
					fmt.Printf("   Log Index: %d\n", transfer.Raw.Index)
					fmt.Printf("   Topics: %d\n", len(transfer.Raw.Topics))
					displayEventMetadata(transfer)
					fmt.Println()
				}
			} else if event.EventKey == eventlog.ERC1155TransferBatch {
				if transfer, ok := event.Unpacked.(erc1155.Erc1155TransferBatch); ok {
					fmt.Printf("%d. ERC1155 Transfer Batch\n", i+1)
					fmt.Printf("   Block: %d\n", transfer.Raw.BlockNumber)
					fmt.Printf("   Transaction: %s\n", transfer.Raw.TxHash.Hex())
					fmt.Printf("   Operator: %s (%s)\n", transfer.Operator.Hex(), formatAddress(transfer.Operator))
					fmt.Printf("   From: %s (%s)\n", transfer.From.Hex(), formatAddress(transfer.From))
					fmt.Printf("   To: %s (%s)\n", transfer.To.Hex(), formatAddress(transfer.To))
					fmt.Printf("   Contract: %s (%s)\n", transfer.Raw.Address.Hex(), formatAddress(transfer.Raw.Address))
					fmt.Printf("   Log Index: %d\n", transfer.Raw.Index)
					fmt.Printf("   Topics: %d\n", len(transfer.Raw.Topics))
					fmt.Printf("   Batch Items (%d):\n", len(transfer.Ids))
					// Show individual token IDs and amounts
					for j, id := range transfer.Ids {
						if j < len(transfer.Values) {
							fmt.Printf("     - Token ID: %s, Amount: %s\n", formatBigInt(id), formatBigInt(transfer.Values[j]))
						}
					}
					displayEventMetadata(transfer)
					fmt.Println()
				}
			}
		}
	}

	// Display raw event data for debugging (if verbose flag is added)
	if len(events) > 0 {
		fmt.Printf("=== Raw Event Data (First 3 events) ===\n")
		for i, event := range events {
			if i >= 3 { // Limit to first 3 events to avoid spam
				fmt.Printf("... and %d more events\n", len(events)-3)
				break
			}
			fmt.Printf("Event %d:\n", i+1)
			fmt.Printf("  Contract Key: %s\n", event.ContractKey)
			fmt.Printf("  Event Key: %s\n", event.EventKey)
			fmt.Printf("  ABI Event Name: %s\n", event.ABIEvent.Name)
			fmt.Printf("  Unpacked Type: %T\n", event.Unpacked)
			fmt.Println()
		}
	}

	// Summary
	fmt.Printf("Summary:\n")
	fmt.Printf("- ERC20 transfers: %d\n", len(erc20Events))
	fmt.Printf("- ERC721 transfers: %d\n", len(erc721Events))
	fmt.Printf("- ERC1155 transfers: %d\n", len(erc1155Events))
	fmt.Printf("- Total transfers: %d\n", len(events))
}

func showHelp() {
	fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
	fmt.Println("Options:")
	fmt.Println("  -rpc string")
	fmt.Println("        Ethereum RPC URL (default: https://mainnet.infura.io/v3/YOUR_KEY)")
	fmt.Println("  -account string")
	fmt.Println("        Account address to filter transfers for (required)")
	fmt.Println("  -start int")
	fmt.Println("        Start block number (required)")
	fmt.Println("  -end int")
	fmt.Println("        End block number (required)")
	fmt.Println("  -direction string")
	fmt.Println("        Transfer direction: send, receive, or both (default: both)")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s -account 0x1234... -start 18000000 -end 18001000\n", os.Args[0])
	fmt.Printf("  %s -account 0x1234... -start 18000000 -end 18001000 -direction send\n", os.Args[0])
	fmt.Printf("  %s -rpc https://mainnet.infura.io/v3/YOUR_KEY -account 0x1234... -start 18000000 -end 18001000\n", os.Args[0])
}
