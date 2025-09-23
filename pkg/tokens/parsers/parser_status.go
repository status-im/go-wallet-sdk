package parsers

import (
	"encoding/json"
	"slices"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
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

// Parse parses raw bytes as StatusTokenList and converts to TokenList.
// ID, Source, FetchedTimestamp are set by the caller.
func (p *StatusTokenListParser) Parse(raw []byte, supportedChains []uint64) (*types.TokenList, error) {
	var tokenList StatusTokenList
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
		for chainID, address := range t.Contracts {
			if !common.IsHexAddress(address) || !slices.Contains(supportedChains, chainID) {
				continue
			}

			token := types.Token{
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
