package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/go-wallet-sdk/pkg/ens"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
)

const (
	defaultTimeout = 30 * time.Second

	chainIDMainnet = 1
	chainIDSepolia = 11155111
	chainIDHolesky = 17000
)

func main() {
	// Define command-line flags
	rpcURL := flag.String("rpc", "", "Ethereum RPC endpoint URL (required)")
	ensName := flag.String("name", "", "ENS name for forward resolution (e.g., vitalik.eth)")
	address := flag.String("address", "", "Ethereum address for reverse resolution (e.g., 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045)")
	timeout := flag.Duration("timeout", defaultTimeout, "Operation timeout duration")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	// Show help if requested or no arguments provided
	if *help || flag.NFlag() == 0 {
		showHelp()
		return
	}

	// Validate required flags
	if *rpcURL == "" {
		fmt.Fprintf(os.Stderr, "Error: -rpc flag is required\n\n")
		showHelp()
		os.Exit(1)
	}

	if *ensName == "" && *address == "" {
		fmt.Fprintf(os.Stderr, "Error: Either -name or -address flag is required\n\n")
		showHelp()
		os.Exit(1)
	}

	if *ensName != "" && *address != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both -name and -address flags simultaneously\n\n")
		showHelp()
		os.Exit(1)
	}

	// Run the resolution
	if err := run(*rpcURL, *ensName, *address, *timeout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(rpcURL, ensName, addressStr string, timeout time.Duration) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Connect to Ethereum node
	fmt.Printf("Connecting to RPC endpoint: %s\n", rpcURL)
	rpcClient, err := rpc.DialContext(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC endpoint: %w", err)
	}
	defer rpcClient.Close()

	// Create ethclient
	client := ethclient.NewClient(rpcClient)

	// Get and display chain ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}
	chainName := getChainName(chainID.Uint64())
	fmt.Printf("Connected to %s (Chain ID: %d)\n\n", chainName, chainID.Uint64())

	// Create ENS resolver
	resolver, err := ens.NewResolver(client)
	if err != nil {
		return fmt.Errorf("failed to create ENS resolver: %w", err)
	}

	// Perform resolution based on input
	if ensName != "" {
		return resolveName(resolver, ensName)
	}
	return resolveAddress(resolver, addressStr)
}

func resolveName(resolver *ens.Resolver, name string) error {
	fmt.Printf("Resolving ENS name: %s\n", name)
	fmt.Println("----------------------------------------")

	address, err := resolver.AddressOf(name)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Successfully resolved\n")
	fmt.Printf("  ENS Name: %s\n", name)
	fmt.Printf("  Address:  %s\n", address.Hex())

	return nil
}

func resolveAddress(resolver *ens.Resolver, addressStr string) error {
	// Validate and parse address
	if !common.IsHexAddress(addressStr) {
		return fmt.Errorf("invalid Ethereum address format: %s", addressStr)
	}

	address := common.HexToAddress(addressStr)

	fmt.Printf("Performing reverse resolution for: %s\n", address.Hex())
	fmt.Println("----------------------------------------")

	name, err := resolver.GetName(address)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Successfully resolved\n")
	fmt.Printf("  Address:  %s\n", address.Hex())
	fmt.Printf("  ENS Name: %s\n", name)

	return nil
}

func getChainName(chainID uint64) string {
	switch chainID {
	case chainIDMainnet:
		return "Ethereum Mainnet"
	case chainIDSepolia:
		return "Sepolia Testnet"
	case chainIDHolesky:
		return "Holesky Testnet"
	default:
		return fmt.Sprintf("Chain ID: %d", chainID)
	}
}

func showHelp() {
	fmt.Println("ENS Resolver Example")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("A command-line tool for Ethereum Name Service (ENS) resolution.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ens-resolver-example -rpc <url> -name <ens-name>")
	fmt.Println("  ens-resolver-example -rpc <url> -address <eth-address>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -rpc string")
	fmt.Println("        Ethereum RPC endpoint URL (required)")
	fmt.Println("        Example: https://mainnet.infura.io/v3/YOUR-PROJECT-ID")
	fmt.Println()
	fmt.Println("  -name string")
	fmt.Println("        ENS name for forward resolution (name -> address)")
	fmt.Println("        Example: vitalik.eth")
	fmt.Println()
	fmt.Println("  -address string")
	fmt.Println("        Ethereum address for reverse resolution (address -> name)")
	fmt.Println("        Example: 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
	fmt.Println()
	fmt.Println("  -timeout duration")
	fmt.Println("        Operation timeout duration (default: 30s)")
	fmt.Println("        Example: 1m, 30s, 500ms")
	fmt.Println()
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println()
	fmt.Println("  Forward resolution (ENS name to address):")
	fmt.Println("    ./ens-resolver-example -rpc https://mainnet.infura.io/v3/YOUR-PROJECT-ID -name vitalik.eth")
	fmt.Println()
	fmt.Println("  Reverse resolution (address to ENS name):")
	fmt.Println("    ./ens-resolver-example -rpc https://mainnet.infura.io/v3/YOUR-PROJECT-ID -address 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
	fmt.Println()
	fmt.Println("  With custom timeout:")
	fmt.Println("    ./ens-resolver-example -rpc https://mainnet.infura.io/v3/YOUR-PROJECT-ID -name vitalik.eth -timeout 1m")
	fmt.Println()
	fmt.Println("Supported Chains:")
	fmt.Println("  Any chain where the ENS registry contract is deployed.")
	fmt.Println("  Use ens.ENSContractExists() to check if ENS is available.")
	fmt.Println()
}
