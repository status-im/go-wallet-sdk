package main

// BalanceResult represents the result of a balance fetch operation
type BalanceResult struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	Wei     string `json:"wei"`
	Error   string `json:"error,omitempty"`
}

// ChainConfig represents a user-supplied chain config
type ChainConfig struct {
	ChainID uint64 `json:"chainId"`
	RPCURL  string `json:"rpcUrl"`
	Name    string `json:"name,omitempty"`
}

// FetchRequest represents the request from the frontend
type FetchRequest struct {
	Chains    []ChainConfig `json:"chains"`
	Addresses []string      `json:"addresses"`
	BlockNum  string        `json:"blockNum"`
}

// FetchResponse represents the response to the frontend
type FetchResponse struct {
	Results map[string]map[string]BalanceResult `json:"results"` // chainID -> address -> result
	Errors  []string                            `json:"errors"`
}
