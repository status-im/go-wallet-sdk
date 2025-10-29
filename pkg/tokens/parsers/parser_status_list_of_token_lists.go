package parsers

import (
	"encoding/json"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

// StatusListOfTokenListsParser parses tokens in the StatusListOfTokenLists format.
type StatusListOfTokenListsParser struct{}

// Parse parses raw bytes as a StatusListOfTokenLists and converts to ListOfTokenLists.
func (p *StatusListOfTokenListsParser) Parse(raw []byte) (*types.ListOfTokenLists, error) {
	var listOfTokenLists types.ListOfTokenLists
	if err := json.Unmarshal(raw, &listOfTokenLists); err != nil {
		return nil, err
	}

	return &listOfTokenLists, nil
}
