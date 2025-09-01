package tokenlists

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// StandardTokenListParser parses tokens in the StandardTokenList format.
type StandardTokenListParser struct{}

// Parse parses raw bytes as a StandardTokenList and converts to Token objects.
func (p *StandardTokenListParser) Parse(raw []byte, sourceURL string, fetchedAt time.Time, supportedChains []uint64) (*TokenList, error) {
	var tokenList StandardTokenList
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
		if !common.IsHexAddress(t.Address) || !slices.Contains(supportedChains, t.ChainID) {
			continue
		}

		token := Token{
			ChainID:  t.ChainID,
			Address:  common.HexToAddress(t.Address),
			Name:     t.Name,
			Symbol:   t.Symbol,
			Decimals: t.Decimals,
			LogoURI:  t.LogoURI,
		}

		result.Tokens = append(result.Tokens, &token)
	}

	return result, nil
}
