package parsers

import (
	"encoding/json"
	"slices"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

// StandardTokenList represents the TokenLists standard format.
type StandardTokenList struct {
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
	Version   struct {
		Major int `json:"major"`
		Minor int `json:"minor"`
		Patch int `json:"patch"`
	} `json:"version"`
	Tags     map[string]interface{} `json:"tags"`
	LogoURI  string                 `json:"logoURI"`
	Keywords []string               `json:"keywords"`
	Tokens   []struct {
		ChainID  uint64 `json:"chainId"`
		Address  string `json:"address"`
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Decimals uint   `json:"decimals"`
		LogoURI  string `json:"logoURI"`
	} `json:"tokens"`
}

// StandardTokenListParser parses tokens in the StandardTokenList format.
type StandardTokenListParser struct{}

// Parse parses raw bytes as a StandardTokenList and converts to TokenList.
// ID, Source, FetchedTimestamp are set by the caller.
func (p *StandardTokenListParser) Parse(raw []byte, supportedChains []uint64) (*types.TokenList, error) {
	var tokenList StandardTokenList
	if err := json.Unmarshal(raw, &tokenList); err != nil {
		return nil, err
	}

	result := &types.TokenList{
		Name:      tokenList.Name,
		Timestamp: tokenList.Timestamp,
		Version:   tokenList.Version,
		Tags:      tokenList.Tags,
		LogoURI:   tokenList.LogoURI,
		Keywords:  tokenList.Keywords,
		Tokens:    make([]*types.Token, 0),
	}

	for _, t := range tokenList.Tokens {
		if !common.IsHexAddress(t.Address) || !slices.Contains(supportedChains, t.ChainID) {
			continue
		}

		token := types.Token{
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
