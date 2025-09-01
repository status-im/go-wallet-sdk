package tokenlists

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
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

// Parse parses raw bytes as CoinGecko tokens and converts to Token objects.
func (p *CoinGeckoAllTokensParser) Parse(raw []byte, sourceURL string, fetchedAt time.Time, supportedChains []uint64) (*TokenList, error) {
	var tokens []CoinGeckoAllTokens
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return nil, err
	}

	result := &TokenList{
		Source: sourceURL,
		Tokens: make([]*Token, 0),
	}

	if !fetchedAt.IsZero() {
		result.FetchedTimestamp = fetchedAt.Format(time.RFC3339)
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

			token := Token{
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
