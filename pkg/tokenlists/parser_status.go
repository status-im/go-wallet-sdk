package tokenlists

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// StatusTokenList represents a token list in Status format.
type StatusTokenList struct {
	StandardTokenList
	Tokens []struct {
		CrossChainID string            `json:"crossChainId"`
		Symbol       string            `json:"symbol"`
		Name         string            `json:"name"`
		Decimals     uint              `json:"decimals"`
		LogoURI      string            `json:"logoURI"`
		Contracts    map[uint64]string `json:"contracts"`
	} `json:"tokens"`
}

// StatusTokenListParser parses tokens from Status format.
type StatusTokenListParser struct{}

// Parse parses raw bytes as StatusTokenList and converts to TokenList objects.
func (p *StatusTokenListParser) Parse(raw []byte, sourceURL string, fetchedAt time.Time, supportedChains []uint64) (*TokenList, error) {
	var tokenList StatusTokenList
	if err := json.Unmarshal(raw, &tokenList); err != nil {
		return nil, err
	}

	result := &TokenList{
		Name:             tokenList.Name,
		Timestamp:        tokenList.Timestamp,
		FetchedTimestamp: tokenList.Timestamp, // by default (if fetchedAt is not provided) the list's `FetchedTimestamp` is the list's `Timestamp` (used for local lists)
		Source:           sourceURL,
		Version:          tokenList.Version,
		Tags:             tokenList.Tags,
		LogoURI:          tokenList.LogoURI,
		Keywords:         tokenList.Keywords,
		Tokens:           make([]*Token, 0),
	}

	if !fetchedAt.IsZero() {
		result.FetchedTimestamp = fetchedAt.Format(time.RFC3339)
	}

	for _, t := range tokenList.Tokens {
		for chainID, address := range t.Contracts {
			if !common.IsHexAddress(address) || !slices.Contains(supportedChains, chainID) {
				continue
			}

			token := Token{
				CrossChainID: t.CrossChainID,
				ChainID:      chainID,
				Address:      common.HexToAddress(address),
				Name:         t.Name,
				Symbol:       t.Symbol,
				Decimals:     t.Decimals,
				LogoURI:      t.LogoURI,
			}

			result.Tokens = append(result.Tokens, &token)
		}
	}

	return result, nil
}
