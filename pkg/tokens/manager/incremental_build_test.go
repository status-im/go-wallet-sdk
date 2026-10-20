package manager_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/common"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/manager"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

type memContentStore struct {
	mu   sync.Mutex
	data map[string]autofetcher.Content
}

func (s *memContentStore) GetEtag(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[id].Etag, nil
}

func (s *memContentStore) Get(id string) (autofetcher.Content, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[id], nil
}

func (s *memContentStore) Set(id string, content autofetcher.Content) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = content
	return nil
}

func (s *memContentStore) GetAll() (map[string]autofetcher.Content, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all := make(map[string]autofetcher.Content, len(s.data))
	for id, c := range s.data {
		all[id] = c
	}
	return all, nil
}

type countingParser struct {
	mu     sync.Mutex
	parsed map[string]int // by list name
}

func (p *countingParser) Parse(raw []byte, chains []uint64) (*types.TokenList, error) {
	list, err := (&parsers.StandardTokenListParser{}).Parse(raw, chains)
	if err == nil {
		p.mu.Lock()
		p.parsed[list.Name]++
		p.mu.Unlock()
	}
	return list, err
}

func (p *countingParser) counts() map[string]int {
	p.mu.Lock()
	defer p.mu.Unlock()
	c := make(map[string]int, len(p.parsed))
	for k, v := range p.parsed {
		c[k] = v
	}
	return c
}

func tokenListJSON(name, symbol string) []byte {
	return []byte(`{"name":"` + name + `","timestamp":"2025-01-01T00:00:00Z","version":{"major":1,"minor":0,"patch":0},"tokens":[` +
		`{"chainId":1,"address":"0x1111111111111111111111111111111111111111","name":"` + symbol + `","symbol":"` + symbol + `","decimals":18}]}`)
}

func TestManager_RebuildParsesOnlyChangedLists(t *testing.T) {
	parser := &countingParser{parsed: map[string]int{}}
	store := &memContentStore{data: map[string]autofetcher.Content{
		"remote": {SourceURL: "https://example.com/remote.json", Data: tokenListJSON("remote", "RMT")},
	}}
	initialLists := map[string][]byte{
		"main":  tokenListJSON("main", "MAIN"),
		"other": tokenListJSON("other", "OTH"),
	}
	config := &manager.Config{
		MainListID:          "main",
		InitialListIDs:      manager.InitialListIDsFromMap(initialLists),
		InitialListProvider: manager.StaticInitialListProvider(initialLists),
		CustomParsers:       map[string]parsers.TokenListParser{"main": parser, "other": parser, "remote": parser},
		Chains:              []uint64{common.EthereumMainnet},
	}
	m, err := manager.New(config, nil, store, nil)
	require.NoError(t, err)
	require.NoError(t, m.Start(context.Background(), false, nil))
	defer func() { require.NoError(t, m.Stop()) }()
	require.Equal(t, map[string]int{"main": 1, "other": 1, "remote": 1}, parser.counts())

	// Rebuild with nothing changed: no list is parsed again.
	require.NoError(t, m.SetChains([]uint64{common.EthereumMainnet}))
	require.Equal(t, map[string]int{"main": 1, "other": 1, "remote": 1}, parser.counts())
	require.Len(t, m.TokenLists(), 4) // native + 3

	// A stored list replaces the initial one: only that list is parsed.
	require.NoError(t, store.Set("other", autofetcher.Content{SourceURL: "https://example.com/other.json", Data: tokenListJSON("other", "OTH2")}))
	require.NoError(t, m.SetChains([]uint64{common.EthereumMainnet}))
	require.Equal(t, map[string]int{"main": 1, "other": 2, "remote": 1}, parser.counts())
	other, ok := m.TokenList("other")
	require.True(t, ok)
	require.Equal(t, "OTH2", other.Tokens[0].Symbol)
	require.Equal(t, "https://example.com/other.json", other.Source)

	// Different chains invalidate every parsed list.
	require.NoError(t, m.SetChains([]uint64{common.EthereumMainnet, common.BSCMainnet}))
	require.Equal(t, map[string]int{"main": 2, "other": 3, "remote": 2}, parser.counts())
}

// SetChains must not alias the caller's slice: mutating it afterwards would silently change the
// chains the parsed-list cache was built for.
func TestManager_SetChainsDoesNotAliasCallerSlice(t *testing.T) {
	parser := &countingParser{parsed: map[string]int{}}
	initialLists := map[string][]byte{"main": tokenListJSON("main", "MAIN")}
	m, err := manager.New(&manager.Config{
		MainListID:          "main",
		InitialListIDs:      manager.InitialListIDsFromMap(initialLists),
		InitialListProvider: manager.StaticInitialListProvider(initialLists),
		CustomParsers:       map[string]parsers.TokenListParser{"main": parser},
		Chains:              []uint64{common.EthereumMainnet},
	}, nil, &memContentStore{data: map[string]autofetcher.Content{}}, nil)
	require.NoError(t, err)
	require.NoError(t, m.Start(context.Background(), false, nil))
	defer func() { require.NoError(t, m.Stop()) }()

	chains := []uint64{common.EthereumMainnet}
	require.NoError(t, m.SetChains(chains))
	chains[0] = common.BSCMainnet
	require.NoError(t, m.SetChains([]uint64{common.EthereumMainnet}))
	require.Equal(t, 1, parser.counts()["main"], "unchanged chains must reuse the parsed list")
}
