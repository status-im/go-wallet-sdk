package main

// BalanceResult represents the result of a balance fetch operation
type BalanceResult struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	Wei     string `json:"wei"`
	Error   string `json:"error,omitempty"`
}

// ERC20BalanceResult represents the result of an ERC20 balance fetch operation
type ERC20BalanceResult struct {
	TokenAddress string `json:"tokenAddress"`
	TokenSymbol  string `json:"tokenSymbol,omitempty"`
	TokenName    string `json:"tokenName,omitempty"`
	Balance      string `json:"balance"`
	Wei          string `json:"wei"`
	Decimals     int    `json:"decimals"`
	Error        string `json:"error,omitempty"`
}

// AccountBalances represents all balances for a single account
type AccountBalances struct {
	Address       string                        `json:"address"`
	NativeBalance BalanceResult                 `json:"nativeBalance"`
	ERC20Balances map[string]ERC20BalanceResult `json:"erc20Balances"` // tokenAddress -> balance
}

// ChainConfig represents a user-supplied chain config
type ChainConfig struct {
	ChainID        uint64   `json:"chainId"`
	RPCURL         string   `json:"rpcUrl"`
	Name           string   `json:"name,omitempty"`
	TokenAddresses []string `json:"tokenAddresses,omitempty"` // ERC20 token addresses to fetch
}

// FetchRequest represents the request from the frontend
type FetchRequest struct {
	Chains    []ChainConfig `json:"chains"`
	Addresses []string      `json:"addresses"`
	BlockNum  string        `json:"blockNum"`
}

// FetchResponse represents the response to the frontend
type FetchResponse struct {
	Results map[string]map[string]AccountBalances `json:"results"` // chainID -> address -> account balances
	Errors  []string                              `json:"errors"`
}
