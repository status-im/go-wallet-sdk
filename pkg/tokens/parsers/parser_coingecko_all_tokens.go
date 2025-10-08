package parsers

import (
	"encoding/json"
	"slices"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

// Note:
// Keeping this parser here, but it doesn't provide decimals (they all will be 0).
// This parser can be updated to use multicall3 (more in pkg/contracts/multicall3) and fetch the decimals from contracts.

// CoinGeckoAllTokens represents a token in CoinGecko format.
type CoinGeckoAllTokens struct {
	ID        string            `json:"id"`
	Symbol    string            `json:"symbol"`
	Name      string            `json:"name"`
	Platforms map[string]string `json:"platforms"`
}

// CoinGeckoAllTokensParser parses tokens from CoinGecko format.
type CoinGeckoAllTokensParser struct {
	chainsMapper map[string]uint64
}

// NewCoinGeckoAllTokensParser creates a new CoinGeckoAllTokensParser with the given chains mapper.
func NewCoinGeckoAllTokensParser(chainsMapper map[string]uint64) *CoinGeckoAllTokensParser {
	return &CoinGeckoAllTokensParser{
		chainsMapper: chainsMapper,
	}
}

// Parse parses raw bytes as CoinGecko tokens and converts to TokenList.
// ID, Source, FetchedTimestamp are set by the caller.
func (p *CoinGeckoAllTokensParser) Parse(raw []byte, supportedChains []uint64) (*types.TokenList, error) {
	var tokens []CoinGeckoAllTokens
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return nil, err
	}

	result := &types.TokenList{
		Tokens: make([]*types.Token, 0),
	}

	for _, t := range tokens {
		for platform, address := range t.Platforms {
			chainID, exists := p.chainsMapper[platform]
			if !exists {
				continue
			}

			if !common.IsHexAddress(address) || !slices.Contains(supportedChains, chainID) {
				continue
			}

			token := types.Token{
				ChainID: chainID,
				Address: common.HexToAddress(address),
				Name:    t.Name,
				Symbol:  t.Symbol,
				// CoinGecko doesn't provide decimals, logo URI, etc.
			}

			result.Tokens = append(result.Tokens, &token)
		}
	}

	return result, nil
}
