package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/go-wallet-sdk/pkg/balance/multistandardfetcher"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/erc20"
	"github.com/status-im/go-wallet-sdk/pkg/contracts/multicall3"
	"github.com/status-im/go-wallet-sdk/pkg/ens"
	"github.com/status-im/go-wallet-sdk/pkg/ethclient"
	tokenbuilder "github.com/status-im/go-wallet-sdk/pkg/tokens/builder"
	tokenfetcher "github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	tokentypes "github.com/status-im/go-wallet-sdk/pkg/tokens/types"
	"github.com/status-im/go-wallet-sdk/pkg/txgenerator"
)

// --- Configuration & Constants ---

const (
	LightScryptN = 1 << 12 // 4096
	LightScryptP = 6
)

var rpcUrls = map[uint64]string{
	1:          "https://ethereum-rpc.publicnode.com",
	42161:      "https://arbitrum-one-rpc.publicnode.com",
	10:         "https://optimism-rpc.publicnode.com",
	11155111:   "https://ethereum-sepolia-rpc.publicnode.com",
	421614:     "https://arbitrum-sepolia-rpc.publicnode.com",
	11155420:   "https://optimism-sepolia-rpc.publicnode.com",
	1660990954: "https://public.sepolia.rpc.status.network",
}

// Global state
// EthClientInterface defines the interface for Ethereum clients used in this app
type EthClientInterface interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	EthGetBalance(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	EthChainId(ctx context.Context) (*big.Int, error)
	EthGasPrice(ctx context.Context) (*big.Int, error)
	EthGetTransactionCount(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	EthEstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	EthSendRawTransaction(ctx context.Context, tx []byte) (common.Hash, error)
}

var (
	walletKeystore   *keystore.KeyStore
	clients          = make(map[uint64]EthClientInterface)
	tokenLists       = make(map[uint64][]tokentypes.Token)
	tokenInfo        = make(map[uint64]map[common.Address]tokentypes.Token) // chainID -> address -> token info
	enabledChains    = make(map[uint64]bool)                                // chainID -> enabled
	unlockedAccounts = make(map[common.Address]bool)                        // address -> unlocked (password stored in keystore)
	ksMu             sync.Mutex
	chainMu          sync.RWMutex
)

var chainNames = map[uint64]string{
	1:          "Ethereum",
	42161:      "Arbitrum",
	10:         "Optimism",
	11155111:   "Sepolia",
	421614:     "Arbitrum Sepolia",
	11155420:   "Optimism Sepolia",
	1660990954: "Status Sepolia",
}

// loggingRPCClient wraps an RPCClient to log all RPC calls
// It embeds *rpc.Client to support type assertion in ethclient.NewClient
type loggingRPCClient struct {
	*rpc.Client // Embed to support type assertion
	chainID     uint64
}

func (l *loggingRPCClient) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	chainName := chainNames[l.chainID]
	if chainName == "" {
		chainName = fmt.Sprintf("Chain %d", l.chainID)
	}
	log.Printf("[RPC] %s (%d) -> %s with %d args", chainName, l.chainID, method, len(args))
	start := time.Now()
	err := l.Client.CallContext(ctx, result, method, args...)
	duration := time.Since(start)
	if err != nil {
		log.Printf("[RPC] %s (%d) -> %s ERROR after %v: %v", chainName, l.chainID, method, duration, err)
	} else {
		log.Printf("[RPC] %s (%d) -> %s SUCCESS after %v", chainName, l.chainID, method, duration)
	}
	return err
}

// loggingEthClient wraps ethclient.Client to log CallContract calls
// This is needed because CallContract uses gethEthClient which bypasses our RPC wrapper
type loggingEthClient struct {
	*ethclient.Client
	chainID uint64
}

func (l *loggingEthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	chainName := chainNames[l.chainID]
	if chainName == "" {
		chainName = fmt.Sprintf("Chain %d", l.chainID)
	}
	blockStr := "latest"
	if blockNumber != nil {
		blockStr = blockNumber.String()
	}
	toStr := "nil"
	if msg.To != nil {
		toStr = msg.To.Hex()
	}
	log.Printf("[RPC] %s (%d) -> eth_call to=%s block=%s", chainName, l.chainID, toStr, blockStr)
	start := time.Now()
	result, err := l.Client.CallContract(ctx, msg, blockNumber)
	duration := time.Since(start)
	if err != nil {
		log.Printf("[RPC] %s (%d) -> eth_call ERROR after %v: %v", chainName, l.chainID, duration, err)
	} else {
		log.Printf("[RPC] %s (%d) -> eth_call SUCCESS after %v (result length: %d)", chainName, l.chainID, duration, len(result))
	}
	return result, err
}

// ChainInfo represents chain information
type ChainInfo struct {
	ChainID uint64 `json:"chainId"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// --- Initialization ---

func main() {
	// Initialize Ethereum Clients with logging wrapper
	for chainID, url := range rpcUrls {
		rpcClient, err := rpc.Dial(url)
		if err == nil {
			// Wrap RPC client with logging (embed *rpc.Client to support type assertion)
			wrappedRPC := &loggingRPCClient{
				Client:  rpcClient,
				chainID: chainID,
			}
			baseClient := ethclient.NewClient(wrappedRPC)
			// Wrap the client to intercept CallContract calls (used by Multicall3)
			clients[chainID] = &loggingEthClient{
				Client:  baseClient,
				chainID: chainID,
			}
			// Enable all chains by default, except Ethereum Mainnet (chain ID 1)
			enabledChains[chainID] = chainID != 1
		}
	}

	// Fetch Real-time Token Lists
	loadTokens()

	// HTTP Routing
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/init-keystore", handleInitKeystore)
	http.HandleFunc("/create-account", handleCreateAccount)
	http.HandleFunc("/unlock", handleUnlockAccount)
	http.HandleFunc("/accounts", handleGetAccounts)
	http.HandleFunc("/chains", handleGetChains)
	http.HandleFunc("/chains/toggle", handleToggleChain)
	http.HandleFunc("/balances", handleBalances)
	http.HandleFunc("/transfer/preview", handleTransferPreview)
	http.HandleFunc("/transfer", handleTransfer)
	http.HandleFunc("/tokens/stats", handleTokenStats)
	http.HandleFunc("/tokens/chain", handleTokensByChain)

	fmt.Println("🚀 Mini Wallet live at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadTokens() {
	// Get supported chain IDs from rpcUrls
	supportedChains := make([]uint64, 0, len(rpcUrls))
	for chainID := range rpcUrls {
		supportedChains = append(supportedChains, chainID)
	}

	// Create token builder
	tokenBuilder := tokenbuilder.New(supportedChains)

	// Add native tokens
	if err := tokenBuilder.AddNativeTokenList(); err != nil {
		log.Printf("Failed to add native tokens: %v", err)
	}

	// Define token list sources
	tokenListSources := []struct {
		id     string
		url    string
		schema string
		parser parsers.TokenListParser
	}{
		{"status", "https://prod.market.status.im/static/token-list.json", "", &parsers.StatusTokenListParser{}},
		{"uniswap", "https://ipfs.io/ipns/tokens.uniswap.org", "https://uniswap.org/tokenlist.schema.json", &parsers.StandardTokenListParser{}},
		{"coingeckoEthereum", "https://prod.market.status.im/v1/token_lists/ethereum/all.json", "", &parsers.StandardTokenListParser{}},
		{"coingeckoOptimism", "https://prod.market.status.im/v1/token_lists/optimistic-ethereum/all.json", "", &parsers.StandardTokenListParser{}},
		{"coingeckoArbitrum", "https://prod.market.status.im/v1/token_lists/arbitrum-one/all.json", "", &parsers.StandardTokenListParser{}},
		{"aave", "https://raw.githubusercontent.com/bgd-labs/aave-address-book/main/tokenlist.json", "", &parsers.StandardTokenListParser{}},
	}

	// Create fetcher
	f := tokenfetcher.New(tokenfetcher.DefaultConfig())
	ctx := context.Background()

	// Fetch and add all token lists
	var successCount, totalTokens int
	for _, source := range tokenListSources {
		fetchDetails := tokenfetcher.FetchDetails{
			ListDetails: tokentypes.ListDetails{
				ID:        source.id,
				SourceURL: source.url,
				Schema:    source.schema,
			},
			Etag: "",
		}

		fetchedData, err := f.Fetch(ctx, fetchDetails)
		if err != nil {
			log.Printf("Failed to fetch %s: %v", source.id, err)
			continue
		}

		if len(fetchedData.JsonData) == 0 {
			log.Printf("No data from %s", source.id)
			continue
		}

		// Add to builder
		err = tokenBuilder.AddRawTokenList(
			source.id,
			fetchedData.JsonData,
			source.url,
			fetchedData.Fetched,
			source.parser,
		)
		if err != nil {
			log.Printf("Failed to parse %s: %v", source.id, err)
			continue
		}

		successCount++
		tokenCount := len(tokenBuilder.GetTokens())
		totalTokens = tokenCount
		log.Printf("✅ Loaded %s: %d total tokens", source.id, tokenCount)
	}

	// Extract tokens from builder and organize by chain
	allTokens := tokenBuilder.GetTokens()
	for _, token := range allTokens {
		tokenLists[token.ChainID] = append(tokenLists[token.ChainID], *token)

		// Create token info map for quick lookup
		if tokenInfo[token.ChainID] == nil {
			tokenInfo[token.ChainID] = make(map[common.Address]tokentypes.Token)
		}
		tokenInfo[token.ChainID][token.Address] = *token
	}

	log.Printf("📊 Loaded %d token lists with %d unique tokens across all chains", successCount, totalTokens)
}

// --- Request Handlers ---

func handleHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func handleInitKeystore(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	ksMu.Lock()
	defer ksMu.Unlock()

	walletKeystore = keystore.NewKeyStore(req.Path, LightScryptN, LightScryptP)
	accs := walletKeystore.Accounts()

	var addrs []string
	for _, a := range accs {
		addrs = append(addrs, a.Address.Hex())
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"accounts": addrs})
}

func handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	acc, err := walletKeystore.NewAccount(req.Password)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	walletKeystore.Unlock(acc, req.Password)
	json.NewEncoder(w).Encode(map[string]string{"address": acc.Address.Hex()})
}

func handleUnlockAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address string `json:"address"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Just mark account as selected/unlocked (no password needed for viewing)
	// Password will be required when signing transactions
	addr := common.HexToAddress(req.Address)
	ksMu.Lock()
	unlockedAccounts[addr] = true
	ksMu.Unlock()

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleBalances(w http.ResponseWriter, r *http.Request) {
	addrStr := r.URL.Query().Get("address")
	address := common.HexToAddress(addrStr)

	type BalRes struct {
		ChainID          uint64 `json:"chainId"`
		ChainName        string `json:"chainName"`
		Symbol           string `json:"symbol"`
		Balance          string `json:"balance"`          // Raw balance in wei/smallest unit
		BalanceFormatted string `json:"balanceFormatted"` // Formatted balance with decimals
		Token            string `json:"token"`
		Decimals         uint   `json:"decimals"`
		LogoURI          string `json:"logoURI"`
	}
	var results []BalRes

	for chainID, client := range clients {
		// Skip disabled chains
		chainMu.RLock()
		enabled := enabledChains[chainID]
		chainMu.RUnlock()
		if !enabled {
			continue
		}

		ctx := r.Context()

		// Get multicall3 address for this chain
		multicallAddr, exists := multicall3.GetMulticall3Address(int64(chainID))
		if !exists {
			// Fallback to direct balance fetch if multicall3 not available
			balance, err := client.EthGetBalance(ctx, address, nil)
			if err != nil {
			} else {
				// Look up native token info from tokenInfo map
				var symbol string = "Native"
				var decimals uint = 18
				var logoURI string = ""

				if chainTokenInfo, ok := tokenInfo[chainID]; ok {
					if nativeToken, ok := chainTokenInfo[common.Address{}]; ok {
						symbol = nativeToken.Symbol
						decimals = nativeToken.Decimals
						logoURI = nativeToken.LogoURI
					}
				}

				formatted := formatBalance(balance, decimals)
				results = append(results, BalRes{
					ChainID:          chainID,
					ChainName:        chainNames[chainID],
					Symbol:           symbol,
					Balance:          balance.String(),
					BalanceFormatted: formatted,
					Token:            common.Address{}.Hex(),
					Decimals:         decimals,
					LogoURI:          logoURI,
				})
			}
			continue
		}

		// Create multicall3 contract instance (client is already wrapped with logging)
		multicallContract, err := multicall3.NewMulticall3Caller(multicallAddr, client)
		if err != nil {
			continue
		}

		// Get first 10 tokens for this chain
		chainTokens := tokenLists[chainID]

		// Build token addresses list
		tokenAddresses := make([]common.Address, len(chainTokens))
		for i := 0; i < len(chainTokens); i++ {
			tokenAddresses[i] = chainTokens[i].Address
		}

		// Create fetch configuration
		config := multistandardfetcher.FetchConfig{
			Native: []multistandardfetcher.AccountAddress{address},
			ERC20: map[multistandardfetcher.AccountAddress][]multistandardfetcher.ContractAddress{
				address: tokenAddresses,
			},
		}

		// Fetch balances
		resultsCh := multistandardfetcher.FetchBalances(ctx, multicallAddr, multicallContract, config, 500)

		// Process results
		resultCount := 0
		for result := range resultsCh {
			resultCount++
			switch result.ResultType {
			case multistandardfetcher.ResultTypeNative:
				native := result.Result.(multistandardfetcher.NativeResult)
				if native.Err != nil {
				} else {
					// Look up native token info from tokenInfo map
					var symbol string = "Native"
					var decimals uint = 18
					var logoURI string = ""

					if chainTokenInfo, ok := tokenInfo[chainID]; ok {
						if nativeToken, ok := chainTokenInfo[common.Address{}]; ok {
							symbol = nativeToken.Symbol
							decimals = nativeToken.Decimals
							logoURI = nativeToken.LogoURI
						}
					}

					formatted := formatBalance(native.Result, decimals)
					results = append(results, BalRes{
						ChainID:          chainID,
						ChainName:        chainNames[chainID],
						Symbol:           symbol,
						Balance:          native.Result.String(),
						BalanceFormatted: formatted,
						Token:            common.Address{}.Hex(),
						Decimals:         decimals,
						LogoURI:          logoURI,
					})
				}

			case multistandardfetcher.ResultTypeERC20:
				erc20 := result.Result.(multistandardfetcher.ERC20Result)
				if erc20.Err != nil {
				} else {
					addedCount := 0
					for i, tokenAddr := range tokenAddresses {
						if bal, ok := erc20.Results[tokenAddr]; ok {
							if bal.Cmp(common.Big0) > 0 {
								token := chainTokens[i]
								decimals := token.Decimals
								formatted := formatBalance(bal, decimals)
								results = append(results, BalRes{
									ChainID:          chainID,
									ChainName:        chainNames[chainID],
									Symbol:           token.Symbol,
									Balance:          bal.String(),
									BalanceFormatted: formatted,
									Token:            tokenAddr.Hex(),
									Decimals:         decimals,
									LogoURI:          token.LogoURI,
								})
								addedCount++
							} else {
							}
						}
					}
				}
			}
		}
	}

	// Sort results by chain name, then by token symbol
	sort.Slice(results, func(i, j int) bool {
		if results[i].ChainName != results[j].ChainName {
			return results[i].ChainName < results[j].ChainName
		}
		return results[i].Symbol < results[j].Symbol
	})

	json.NewEncoder(w).Encode(results)
}

// getMainnetClient returns the Ethereum mainnet client for ENS resolution
func getMainnetClient() (*ethclient.Client, error) {
	mainnetClient, exists := clients[1]
	if !exists {
		return nil, fmt.Errorf("ethereum mainnet client not available")
	}
	// Type assert to get the underlying *ethclient.Client
	if loggingClient, ok := mainnetClient.(*loggingEthClient); ok {
		return loggingClient.Client, nil
	}
	if baseClient, ok := mainnetClient.(*ethclient.Client); ok {
		return baseClient, nil
	}
	return nil, fmt.Errorf("unexpected client type")
}

// resolveENS resolves an ENS name to an address using Ethereum mainnet
func resolveENS(name string) (common.Address, error) {
	mainnetClient, err := getMainnetClient()
	if err != nil {
		return common.Address{}, err
	}

	res, err := ens.NewResolver(mainnetClient)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to create ENS resolver: %v", err)
	}

	resolvedAddr, err := res.AddressOf(name)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to resolve ENS name '%s': %v", name, err)
	}

	if resolvedAddr == (common.Address{}) {
		return common.Address{}, fmt.Errorf("ENS name '%s' resolved to zero address (name may not exist)", name)
	}

	return resolvedAddr, nil
}

func handleTransferPreview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Amount  string `json:"amount"`
		ChainID uint64 `json:"chainId"`
		Token   string `json:"token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	client := clients[req.ChainID]
	if client == nil {
		http.Error(w, "Chain client not available", http.StatusInternalServerError)
		return
	}

	fromAddr := common.HexToAddress(req.From)
	toAddr := common.HexToAddress(req.To)
	originalTo := req.To
	resolvedENS := ""

	// 1. ENS Resolution (always use mainnet)
	if strings.HasSuffix(req.To, ".eth") {
		resolvedAddr, err := resolveENS(req.To)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		toAddr = resolvedAddr
		resolvedENS = resolvedAddr.Hex()
	}

	// 2. Get transaction parameters
	_, err := client.EthChainId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nonce, err := client.EthGetTransactionCount(r.Context(), fromAddr, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gasPrice, err := client.EthGasPrice(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Prepare transaction data for gas estimation
	tokenAddr := common.HexToAddress(req.Token)
	var decimals uint = 18
	var tokenSymbol string = "Native"
	var gasLimit uint64 = 21000

	if tokenAddr != (common.Address{}) {
		// Look up token info
		if chainTokenInfo, ok := tokenInfo[req.ChainID]; ok {
			if token, ok := chainTokenInfo[tokenAddr]; ok {
				decimals = token.Decimals
				tokenSymbol = token.Symbol
			}
		}

		// Estimate gas for ERC20 transfer
		erc20ABI, _ := erc20.Erc20MetaData.GetAbi()
		amtFloat, _, _ := new(big.Float).Parse(req.Amount, 10)
		multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
		amtFloat.Mul(amtFloat, multiplier)
		amt, _ := amtFloat.Int(nil)

		data, _ := erc20ABI.Pack("transfer", toAddr, amt)
		callMsg := ethereum.CallMsg{
			From: fromAddr,
			To:   &tokenAddr,
			Data: data,
		}
		if estimatedGas, err := client.EthEstimateGas(r.Context(), callMsg); err == nil {
			gasLimit = estimatedGas
		} else {
			gasLimit = 65000 // Default fallback
		}
	}

	// Calculate total cost
	totalCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))

	// Convert wei to gwei (1 gwei = 1e9 wei)
	gasPriceGwei := new(big.Float).Quo(new(big.Float).SetInt(gasPrice), big.NewFloat(1e9))
	totalCostGwei := new(big.Float).Quo(new(big.Float).SetInt(totalCost), big.NewFloat(1e9))

	type PreviewResponse struct {
		From          string `json:"from"`
		To            string `json:"to"`
		OriginalTo    string `json:"originalTo"`
		ResolvedENS   string `json:"resolvedENS,omitempty"`
		Amount        string `json:"amount"`
		TokenSymbol   string `json:"tokenSymbol"`
		TokenAddress  string `json:"tokenAddress"`
		ChainID       uint64 `json:"chainId"`
		ChainName     string `json:"chainName"`
		Nonce         uint64 `json:"nonce"`
		GasLimit      uint64 `json:"gasLimit"`
		GasPrice      string `json:"gasPrice"`
		GasPriceGwei  string `json:"gasPriceGwei"`
		TotalCost     string `json:"totalCost"`
		TotalCostGwei string `json:"totalCostGwei"`
		IsNative      bool   `json:"isNative"`
	}

	response := PreviewResponse{
		From:          fromAddr.Hex(),
		To:            toAddr.Hex(),
		OriginalTo:    originalTo,
		ResolvedENS:   resolvedENS,
		Amount:        req.Amount,
		TokenSymbol:   tokenSymbol,
		TokenAddress:  tokenAddr.Hex(),
		ChainID:       req.ChainID,
		ChainName:     chainNames[req.ChainID],
		Nonce:         nonce,
		GasLimit:      gasLimit,
		GasPrice:      gasPrice.String(),
		GasPriceGwei:  gasPriceGwei.Text('f', 6),
		TotalCost:     totalCost.String(),
		TotalCostGwei: totalCostGwei.Text('f', 6),
		IsNative:      tokenAddr == (common.Address{}),
	}

	json.NewEncoder(w).Encode(response)
}

func handleTransfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		From     string `json:"from"`
		To       string `json:"to"`
		Amount   string `json:"amount"`
		ChainID  uint64 `json:"chainId"`
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Password is required for signing
	if req.Password == "" {
		http.Error(w, "Password required for signing", 401)
		return
	}

	// Check if chain is enabled
	chainMu.RLock()
	enabled := enabledChains[req.ChainID]
	chainMu.RUnlock()
	if !enabled {
		http.Error(w, "Chain is disabled", 400)
		return
	}

	client := clients[req.ChainID]
	if client == nil {
		http.Error(w, "Chain client not available", 500)
		return
	}

	fromAddr := common.HexToAddress(req.From)
	toAddr := common.HexToAddress(req.To)

	// 1. ENS Resolution (always use mainnet)
	if strings.HasSuffix(req.To, ".eth") {
		resolvedAddr, err := resolveENS(req.To)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		toAddr = resolvedAddr
	}

	// 2. Get transaction parameters
	chainID, err := client.EthChainId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	nonce, err := client.EthGetTransactionCount(r.Context(), fromAddr, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	gasPrice, err := client.EthGasPrice(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 3. Tx Generation
	// Convert amount from decimal to wei/smallest unit
	tokenAddr := common.HexToAddress(req.Token)
	var decimals uint = 18 // Default for native ETH

	if tokenAddr != (common.Address{}) {
		// Look up token decimals
		if chainTokenInfo, ok := tokenInfo[req.ChainID]; ok {
			if token, ok := chainTokenInfo[tokenAddr]; ok {
				decimals = token.Decimals
			}
		}
	}

	// Parse amount as float and convert to wei
	amtFloat, _, err := new(big.Float).Parse(req.Amount, 10)
	if err != nil {
		http.Error(w, "Invalid amount format", 400)
		return
	}

	// Multiply by 10^decimals
	multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	amtFloat.Mul(amtFloat, multiplier)

	// Convert to big.Int (truncate, not round)
	amt, _ := amtFloat.Int(nil)

	var tx *types.Transaction
	if tokenAddr == (common.Address{}) {
		// Native ETH transfer
		tx, err = txgenerator.TransferETH(txgenerator.TransferETHParams{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:    nonce,
				GasLimit: 21000,
				ChainID:  chainID,
				GasPrice: gasPrice,
			},
			To:    toAddr,
			Value: amt,
		})
	} else {
		// ERC20 token transfer
		// Encode transfer call for gas estimation
		erc20ABI, _ := erc20.Erc20MetaData.GetAbi()
		data, _ := erc20ABI.Pack("transfer", toAddr, amt)
		gasLimit := uint64(65000) // Default for ERC20
		callMsg := ethereum.CallMsg{
			From: fromAddr,
			To:   &tokenAddr,
			Data: data,
		}
		if estimatedGas, err := client.EthEstimateGas(r.Context(), callMsg); err == nil {
			gasLimit = estimatedGas
		}

		tx, err = txgenerator.TransferERC20(txgenerator.TransferERC20Params{
			BaseTxParams: txgenerator.BaseTxParams{
				Nonce:    nonce,
				GasLimit: gasLimit,
				ChainID:  chainID,
				GasPrice: gasPrice,
			},
			TokenAddress: tokenAddr,
			To:           toAddr,
			Amount:       amt,
		})
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 4. Signing - unlock account temporarily for signing
	acc := accounts.Account{Address: fromAddr}
	err = walletKeystore.Unlock(acc, req.Password)
	if err != nil {
		http.Error(w, "Invalid password", 401)
		return
	}
	defer walletKeystore.Lock(acc.Address)

	signedTx, err := walletKeystore.SignTx(acc, tx, chainID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	rlpTx, err := signedTx.MarshalBinary()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 5. Send Raw Transaction
	hash, err := client.EthSendRawTransaction(r.Context(), rlpTx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"hash": hash.Hex()})
}

func handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	ksMu.Lock()
	defer ksMu.Unlock()

	if walletKeystore == nil {
		json.NewEncoder(w).Encode([]string{})
		return
	}

	accs := walletKeystore.Accounts()
	var addrs []string
	for _, a := range accs {
		addrs = append(addrs, a.Address.Hex())
	}
	json.NewEncoder(w).Encode(addrs)
}

func handleGetChains(w http.ResponseWriter, r *http.Request) {
	fullChainNames := map[uint64]string{
		1:          "Ethereum Mainnet",
		42161:      "Arbitrum One",
		10:         "Optimism",
		11155111:   "Sepolia",
		421614:     "Arbitrum Sepolia",
		11155420:   "Optimism Sepolia",
		1660990954: "Status Sepolia",
	}

	chainMu.RLock()
	defer chainMu.RUnlock()

	var chains []ChainInfo
	for chainID := range rpcUrls {
		chains = append(chains, ChainInfo{
			ChainID: chainID,
			Name:    fullChainNames[chainID],
			Enabled: enabledChains[chainID],
		})
	}

	// Sort chains alphabetically by name
	sort.Slice(chains, func(i, j int) bool {
		return chains[i].Name < chains[j].Name
	})

	json.NewEncoder(w).Encode(chains)
}

func handleToggleChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		ChainID uint64 `json:"chainId"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	chainMu.Lock()
	enabledChains[req.ChainID] = req.Enabled
	chainMu.Unlock()

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleTokenStats(w http.ResponseWriter, r *http.Request) {
	type ChainStats struct {
		ChainID   uint64 `json:"chainId"`
		ChainName string `json:"chainName"`
		Count     int    `json:"count"`
	}

	type Stats struct {
		TotalTokens    int          `json:"totalTokens"`
		TotalChains    int          `json:"totalChains"`
		TokensPerChain []ChainStats `json:"tokensPerChain"`
	}

	ksMu.Lock()
	defer ksMu.Unlock()

	stats := Stats{
		TokensPerChain: make([]ChainStats, 0),
	}

	// Count tokens per chain
	for chainID, tokens := range tokenLists {
		stats.TokensPerChain = append(stats.TokensPerChain, ChainStats{
			ChainID:   chainID,
			ChainName: chainNames[chainID],
			Count:     len(tokens),
		})
		stats.TotalTokens += len(tokens)
	}

	stats.TotalChains = len(stats.TokensPerChain)

	json.NewEncoder(w).Encode(stats)
}

func handleTokensByChain(w http.ResponseWriter, r *http.Request) {
	chainIDStr := r.URL.Query().Get("chainId")
	if chainIDStr == "" {
		http.Error(w, "chainId parameter required", http.StatusBadRequest)
		return
	}

	var chainID uint64
	fmt.Sscanf(chainIDStr, "%d", &chainID)

	ksMu.Lock()
	defer ksMu.Unlock()

	tokens := tokenLists[chainID]
	if tokens == nil {
		tokens = []tokentypes.Token{}
	}

	type TokenInfo struct {
		Address      string `json:"address"`
		Symbol       string `json:"symbol"`
		Name         string `json:"name"`
		Decimals     uint   `json:"decimals"`
		LogoURI      string `json:"logoURI"`
		CrossChainID string `json:"crossChainId"`
	}

	result := make([]TokenInfo, len(tokens))
	for i, token := range tokens {
		result[i] = TokenInfo{
			Address:      token.Address.Hex(),
			Symbol:       token.Symbol,
			Name:         token.Name,
			Decimals:     token.Decimals,
			LogoURI:      token.LogoURI,
			CrossChainID: token.CrossChainID,
		}
	}

	json.NewEncoder(w).Encode(result)
}

// formatBalance converts a balance from wei/smallest unit to decimal format
func formatBalance(balance *big.Int, decimals uint) string {
	if balance == nil || balance.Sign() == 0 {
		return "0"
	}

	// Create divisor: 10^decimals
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))

	// Convert balance to float and divide
	balanceFloat := new(big.Float).SetInt(balance)
	balanceFloat.Quo(balanceFloat, divisor)

	// Format with appropriate precision
	return balanceFloat.Text('f', int(decimals))
}
