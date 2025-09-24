package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/status-im/go-wallet-sdk/pkg/balance/multistandardfetcher"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
)

func main() {
	// Configuration
	chainID := int64(1)                                           // Ethereum mainnet
	walletAddress := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045" // vitalik.eth

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		log.Fatalf("RPC_URL environment variable is not set")
	}

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

	// Define tokens and collectibles to check
	vitalikAddr := common.HexToAddress(walletAddress)

	// Popular ERC20 tokens
	erc20Tokens := []common.Address{
		common.HexToAddress("0xA0b86a33E6441b8C4C8C0C4C0C4C0C4C0C4C0C4C0"), // USDC
		common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"),  // DAI
		common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),  // USDT
		common.HexToAddress("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599"),  // WBTC
		common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),  // LINK
		common.HexToAddress("0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"),  // UNI
		common.HexToAddress("0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0"),  // MATIC
		common.HexToAddress("0x95aD61b0a150d79219dCF64E1E6Cc01f0B64C4cE"),  // SHIB
	}

	// Popular ERC721 NFTs
	erc721NFTs := []common.Address{
		common.HexToAddress("0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D"), // Bored Ape Yacht Club
		common.HexToAddress("0x60E4d786628Fea6478F785A6d7e704777c86a7c6"), // Mutant Ape Yacht Club
		common.HexToAddress("0x49cF6f5d44E70224e2E95fd1F953Fd3A0069f1"),   // CryptoPunks
		common.HexToAddress("0xed5AF388653567Af2F388E6224dC7C4b3241C544"), // Azuki
		common.HexToAddress("0x23581767a106ae21c074b2276D25e5C3e136a68b"), // Moonbirds
		common.HexToAddress("0xa1eb40c284c5b44419425c4202fa8dabff31006b"), // POAP
	}

	// Popular ERC1155 collectibles
	erc1155Collectibles := []multistandardfetcher.CollectibleID{
		{
			ContractAddress: common.HexToAddress("0x495f947276749Ce646f68AC8c248420045cb7b5e"), // OpenSea Shared Storefront
			TokenID:         big.NewInt(1),
		},
		{
			ContractAddress: common.HexToAddress("0x495f947276749Ce646f68AC8c248420045cb7b5e"),
			TokenID:         big.NewInt(2),
		},
		{
			ContractAddress: common.HexToAddress("0x76BE3b62873462d2142405439777e971754E8E77"), // Parallel Alpha
			TokenID:         big.NewInt(1),
		},
		{
			ContractAddress: common.HexToAddress("0xd07dc4262bcdbf85190c01c996b4c06a461d2430"), // Rarible
			TokenID:         big.NewInt(24775),
		},
	}

	// Create fetch configuration
	config := multistandardfetcher.FetchConfig{
		Native: []multistandardfetcher.AccountAddress{vitalikAddr},
		ERC20: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
			vitalikAddr: erc20Tokens,
		},
		ERC721: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
			vitalikAddr: erc721NFTs,
		},
		ERC1155: map[multistandardfetcher.AccountAddress][]multistandardfetcher.CollectibleID{
			vitalikAddr: erc1155Collectibles,
		},
	}

	log.Printf("Fetching balances for %s across all token standards...", walletAddress)
	log.Printf("Checking %d ERC20 tokens, %d ERC721 NFTs, and %d ERC1155 collectibles",
		len(erc20Tokens), len(erc721NFTs), len(erc1155Collectibles))

	// Execute fetch
	ctx := context.Background()
	resultsCh := multistandardfetcher.FetchBalances(ctx, multicallAddr, multicallContract, config, 100)

	// Process results
	var totalResults int
	var nativeBalance *big.Int
	var erc20Balances = make(map[common.Address]*big.Int)
	var erc721Balances = make(map[common.Address]*big.Int)
	var erc1155Balances = make(map[multistandardfetcher.CollectibleID]*big.Int)

	for result := range resultsCh {
		totalResults++

		switch result.ResultType {
		case multistandardfetcher.ResultTypeNative:
			native := result.Result.(multistandardfetcher.NativeResult)
			if native.Err != nil {
				log.Printf("❌ Failed to fetch native balance: %v", native.Err)
			} else {
				nativeBalance = native.Result
				log.Printf("✅ Native ETH balance: %s wei (block %s)",
					native.Result.String(), native.AtBlockNumber.String())
			}

		case multistandardfetcher.ResultTypeERC20:
			erc20 := result.Result.(multistandardfetcher.ERC20Result)
			if erc20.Err != nil {
				log.Printf("❌ Failed to fetch ERC20 balances: %v", erc20.Err)
			} else {
				erc20Balances = erc20.Results
				log.Printf("✅ ERC20 balances fetched (block %s)", erc20.AtBlockNumber.String())
			}

		case multistandardfetcher.ResultTypeERC721:
			erc721 := result.Result.(multistandardfetcher.ERC721Result)
			if erc721.Err != nil {
				log.Printf("❌ Failed to fetch ERC721 balances: %v", erc721.Err)
			} else {
				erc721Balances = erc721.Results
				log.Printf("✅ ERC721 balances fetched (block %s)", erc721.AtBlockNumber.String())
			}

		case multistandardfetcher.ResultTypeERC1155:
			erc1155 := result.Result.(multistandardfetcher.ERC1155Result)
			if erc1155.Err != nil {
				log.Printf("❌ Failed to fetch ERC1155 balances: %v", erc1155.Err)
			} else {
				// Convert hashable IDs back to regular IDs for display
				for hashableID, balance := range erc1155.Results {
					originalID := hashableID.ToCollectibleID()
					erc1155Balances[originalID] = balance
				}
				log.Printf("✅ ERC1155 balances fetched (block %s)", erc1155.AtBlockNumber.String())
			}
		}
	}

	// Display results
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("BALANCE REPORT FOR %s\n", walletAddress)
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	// Native balance
	if nativeBalance != nil {
		fmt.Printf("\n💰 NATIVE ETH BALANCE\n")
		fmt.Printf(strings.Repeat("-", 40) + "\n")
		ethBalance := new(big.Float).SetInt(nativeBalance)
		ethBalance.Quo(ethBalance, big.NewFloat(1e18))
		fmt.Printf("ETH: %s ETH (%s wei)\n", ethBalance.Text('f', 6), nativeBalance.String())
	}

	// ERC20 balances
	fmt.Printf("\n🪙 ERC20 TOKEN BALANCES\n")
	fmt.Printf(strings.Repeat("-", 40) + "\n")
	erc20Count := 0
	for token, balance := range erc20Balances {
		if balance.Cmp(big.NewInt(0)) > 0 {
			erc20Count++
			fmt.Printf("%s: %s\n", getTokenSymbol(token), balance.String())
		}
	}
	if erc20Count == 0 {
		fmt.Printf("No ERC20 token balances found\n")
	} else {
		fmt.Printf("Found %d tokens with non-zero balances\n", erc20Count)
	}

	// ERC721 balances
	fmt.Printf("\n🖼️  ERC721 NFT BALANCES\n")
	fmt.Printf(strings.Repeat("-", 40) + "\n")
	erc721Count := 0
	for nft, balance := range erc721Balances {
		if balance.Cmp(big.NewInt(0)) > 0 {
			erc721Count++
			fmt.Printf("%s: %s NFTs\n", getNFTSymbol(nft), balance.String())
		}
	}
	if erc721Count == 0 {
		fmt.Printf("No ERC721 NFT balances found\n")
	} else {
		fmt.Printf("Found %d NFT collections with non-zero balances\n", erc721Count)
	}

	// ERC1155 balances
	fmt.Printf("\n🎨 ERC1155 COLLECTIBLE BALANCES\n")
	fmt.Printf(strings.Repeat("-", 40) + "\n")
	erc1155Count := 0
	for collectible, balance := range erc1155Balances {
		if balance.Cmp(big.NewInt(0)) > 0 {
			erc1155Count++
			fmt.Printf("%s (token %s): %s\n",
				collectible.ContractAddress.Hex()[:10]+"...",
				collectible.TokenID.String(),
				balance.String())
		}
	}
	if erc1155Count == 0 {
		fmt.Printf("No ERC1155 collectible balances found\n")
	} else {
		fmt.Printf("Found %d collectibles with non-zero balances\n", erc1155Count)
	}

	// Summary
	fmt.Printf("\n📊 SUMMARY\n")
	fmt.Printf(strings.Repeat("-", 40) + "\n")
	fmt.Printf("Total result sets processed: %d\n", totalResults)
	fmt.Printf("Native ETH: %s\n", formatBalance(nativeBalance))
	fmt.Printf("ERC20 tokens with balance: %d\n", erc20Count)
	fmt.Printf("ERC721 NFT collections with balance: %d\n", erc721Count)
	fmt.Printf("ERC1155 collectibles with balance: %d\n", erc1155Count)
}

// Helper function to get token symbols (simplified mapping)
func getTokenSymbol(address common.Address) string {
	symbols := map[string]string{
		"0xA0b86a33E6441b8C4C8C0C4C0C4C0C4C0C4C0C4C0": "USDC",
		"0x6B175474E89094C44Da98b954EedeAC495271d0F":  "DAI",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7":  "USDT",
		"0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599":  "WBTC",
		"0x514910771AF9Ca656af840dff83E8264EcF986CA":  "LINK",
		"0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984":  "UNI",
		"0x7D1AfA7B718fb893dB30A3aBc0Cfc608AaCfeBB0":  "MATIC",
		"0x95aD61b0a150d79219dCF64E1E6Cc01f0B64C4cE":  "SHIB",
	}
	if symbol, exists := symbols[address.Hex()]; exists {
		return symbol
	}
	return address.Hex()[:10] + "..."
}

// Helper function to get NFT collection names (simplified mapping)
func getNFTSymbol(address common.Address) string {
	names := map[string]string{
		"0xBC4CA0EdA7647A8aB7C2061c2E118A18a936f13D": "BAYC",
		"0x60E4d786628Fea6478F785A6d7e704777c86a7c6": "MAYC",
		"0x49cF6f5d44E70224e2E95fd1F953Fd3A0069f1":   "CryptoPunks",
		"0xed5AF388653567Af2F388E6224dC7C4b3241C544": "Azuki",
		"0x23581767a106ae21c074b2276D25e5C3e136a68b": "Moonbirds",
		"0x8a90CAb2b38dba80c64b7734e58Ee1dB38B8992e": "Doodles",
	}
	if name, exists := names[address.Hex()]; exists {
		return name
	}
	return address.Hex()[:10] + "..."
}

// Helper function to format balance
func formatBalance(balance *big.Int) string {
	if balance == nil {
		return "N/A"
	}
	if balance.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}
	return balance.String()
}
