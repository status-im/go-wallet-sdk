package manager

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/autofetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/builder"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/parsers"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"
)

const (
	LocalSourceURL    = "local"
	CustomTokenListID = "custom"
)

var (
	ErrContentStoreNotProvided                       = fmt.Errorf("content store not provided")
	ErrStoredTokenListIsEmpty                        = fmt.Errorf("stored token list is empty")
	ErrParserNotProvided                             = fmt.Errorf("parser not provided")
	ErrAutoFetcherNotProvided                        = fmt.Errorf("auto fetcher not provided")
	ErrAutoRefreshEnabledButNotifyChannelNotProvided = fmt.Errorf("auto refresh enabled but notify channel not provided")
	ErrManagerNotConfiguredForAutoRefresh            = fmt.Errorf("manager not configured for auto refresh")
	ErrNotFoundInInitialLists                        = fmt.Errorf("not found in initial lists")
)

// manager implements the Manager interface with thread-safe state management.
type manager struct {
	mu sync.RWMutex

	builderMu sync.RWMutex
	builder   *builder.Builder

	notifyCh chan struct{}

	autoRefreshEnabled       bool
	remoteListOfTokenListsID string

	autoFetcher      autofetcher.AutoFetcher
	contentStore     autofetcher.ContentStore
	customTokenStore CustomTokenStore

	mainListID    string
	initialLists  map[string][]byte
	customParsers map[string]parsers.TokenListParser

	chains []uint64

	started         bool
	refreshCancelFn context.CancelFunc
}

// New creates a new Manager instance.
func New(config *Config,
	fetcher fetcher.Fetcher,
	contentStore autofetcher.ContentStore,
	customTokenStore CustomTokenStore) (Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if contentStore == nil {
		return nil, ErrContentStoreNotProvided
	}

	manager := &manager{
		mainListID:    config.MainListID,
		initialLists:  config.InitialLists,
		customParsers: config.CustomParsers,
		chains:        config.Chains,

		contentStore:     contentStore,
		customTokenStore: customTokenStore,
	}

	if config.AutoFetcherConfig != nil {
		var err error
		manager.autoFetcher, err = autofetcher.NewAutofetcherFromRemoteListOfTokenLists(*config.AutoFetcherConfig, fetcher,
			contentStore)
		if err != nil {
			return nil, err
		}
		manager.remoteListOfTokenListsID = config.AutoFetcherConfig.RemoteListOfTokenListsFetchDetails.ID
	}

	return manager, nil
}

// Start begins the Manager service, if notify channel is provided, it will be notified when the token lists are refreshed.
// Once the manager is started, the initial state is built and then the manager will start to manage the refresh of the token lists
// if auto refresh is enabled.
func (m *manager) Start(ctx context.Context, autoRefreshEnabled bool, notifyCh chan struct{}) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return nil
	}

	defer func() {
		if err == nil {
			m.started = true
		}
	}()

	// if auto refresh is enabled, notify channel must be provided, otherwise client cannot be notified about the refresh
	if autoRefreshEnabled && notifyCh == nil {
		err = ErrAutoRefreshEnabledButNotifyChannelNotProvided
		return
	}
	m.autoRefreshEnabled = autoRefreshEnabled

	// if notify channel is provided, auto fetcher must be provided, otherwise there is nothing to notify about
	if notifyCh != nil {
		if m.autoFetcher == nil {
			err = ErrAutoFetcherNotProvided
			return
		}
		m.notifyCh = notifyCh
	}

	// build the initial state
	if err = m.buildState(); err != nil {
		return
	}

	if err = m.manageRefresh(ctx); err != nil {
		return
	}

	return
}

// Stop stops the Manager service.
func (m *manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	// Stop refresh goroutine first
	if m.refreshCancelFn != nil {
		m.refreshCancelFn()
		m.refreshCancelFn = nil
	}

	// Stop autofetcher
	if m.autoFetcher != nil {
		m.autoFetcher.Stop()
	}

	m.started = false
	return nil
}

func (m *manager) manageRefresh(ctx context.Context) error {
	if m.autoFetcher == nil {
		return nil
	}

	// Stop existing refresh goroutine first
	if m.refreshCancelFn != nil {
		m.refreshCancelFn()
	}

	if !m.autoRefreshEnabled {
		m.autoFetcher.Stop()
		m.refreshCancelFn = nil
		return nil
	}

	// Create new context for refresh goroutine
	refreshCtx, cancel := context.WithCancel(ctx)
	m.refreshCancelFn = cancel

	refreshCh := m.autoFetcher.Start(refreshCtx)
	go func() {
		defer cancel()

		for {
			select {
			case refreshErr, ok := <-refreshCh:
				if !ok {
					return
				}
				if refreshErr != nil {
					// an error occurred while refreshing the token lists, continue to wait for the next refresh
					continue
				}

				m.mu.Lock()
				err := m.buildState()
				if err != nil {
					m.mu.Unlock()
					continue
				}

				if m.notifyCh != nil {
					select {
					case m.notifyCh <- struct{}{}:
						// notification sent
					default:
						// Channel is full or closed, skip notification
					}
				}
				m.mu.Unlock()

			case <-refreshCtx.Done():
				return
			}
		}
	}()

	return nil
}

func (m *manager) supportsAutoRefresh() error {
	if m.autoFetcher == nil || m.notifyCh == nil {
		return ErrManagerNotConfiguredForAutoRefresh
	}
	return nil
}

// EnableAutoRefresh enables auto refresh.
func (m *manager) EnableAutoRefresh(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.supportsAutoRefresh(); err != nil {
		return err
	}

	if m.autoRefreshEnabled {
		return nil
	}

	m.autoRefreshEnabled = true

	return m.manageRefresh(ctx)
}

// DisableAutoRefresh disables auto refresh.
func (m *manager) DisableAutoRefresh(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.supportsAutoRefresh(); err != nil {
		return err
	}

	if !m.autoRefreshEnabled {
		return nil
	}
	m.autoRefreshEnabled = false

	return m.manageRefresh(ctx)
}

func (m *manager) TriggerRefresh(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.supportsAutoRefresh(); err != nil {
		return err
	}

	return m.manageRefresh(ctx)
}

// UniqueTokens returns all unique tokens.
func (m *manager) UniqueTokens() []*types.Token {
	m.builderMu.RLock()
	defer m.builderMu.RUnlock()

	if m.builder == nil {
		return nil
	}
	tokens := make([]*types.Token, 0, len(m.builder.GetTokens()))
	for _, token := range m.builder.GetTokens() {
		tokens = append(tokens, token)
	}
	return tokens
}

// GetTokenByChainAddress retrieves a token by chain ID and address.
func (m *manager) GetTokenByChainAddress(chainID uint64, addr common.Address) (*types.Token, bool) {
	m.builderMu.RLock()
	defer m.builderMu.RUnlock()

	if m.builder == nil {
		return nil, false
	}
	key := types.TokenKey(chainID, addr)
	token, exists := m.builder.GetTokens()[key]
	return token, exists
}

// GetTokensByChain returns all tokens for a specific chain.
func (m *manager) GetTokensByChain(chainID uint64) []*types.Token {
	m.builderMu.RLock()
	defer m.builderMu.RUnlock()

	if m.builder == nil {
		return nil
	}
	var tokens []*types.Token
	for _, token := range m.builder.GetTokens() {
		if token.ChainID != chainID {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

// GetTokensByKeys returns tokens by keys.
func (m *manager) GetTokensByKeys(keys []string) ([]*types.Token, error) {
	m.builderMu.RLock()
	defer m.builderMu.RUnlock()

	if m.builder == nil {
		return nil, nil
	}

	tokensMap := m.builder.GetTokens()

	tokens := make([]*types.Token, 0)
	for _, key := range keys {
		token, exists := tokensMap[strings.ToLower(key)]
		if exists {
			tokens = append(tokens, token)
		}
	}
	return tokens, nil
}

// TokenList returns a token list by ID.
func (m *manager) TokenList(id string) (*types.TokenList, bool) {
	m.builderMu.RLock()
	defer m.builderMu.RUnlock()

	if m.builder == nil {
		return nil, false
	}
	tokenList, exists := m.builder.GetTokenLists()[id]
	return tokenList, exists
}

// TokenLists returns all token lists.
func (m *manager) TokenLists() []*types.TokenList {
	m.builderMu.RLock()
	defer m.builderMu.RUnlock()

	if m.builder == nil {
		return nil
	}
	tokenLists := make([]*types.TokenList, 0, len(m.builder.GetTokenLists()))
	for _, tokenList := range m.builder.GetTokenLists() {
		tokenLists = append(tokenLists, tokenList)
	}
	return tokenLists
}

func (m *manager) buildState() error {
	builder := builder.New(m.chains)

	// 1. native token list
	if err := builder.AddNativeTokenList(); err != nil {
		return err
	}

	// merge tokens from all sources in the specified order.
	// 2. main list (remote if available, otherwise initial)
	if err := m.mergeMainList(builder); err != nil {
		return err
	}

	// 3. other initial lists (in deterministic order), remote if available, otherwise initial list
	if err := m.mergeInitialLists(builder); err != nil {
		return err
	}

	// 4. remote lists that are not main or initial lists (in deterministic order)
	if err := m.mergeRemoteLists(builder); err != nil {
		return err
	}

	// 5. custom tokens
	if err := m.mergeCustomTokens(builder); err != nil {
		return err
	}

	m.builderMu.Lock()
	m.builder = builder
	m.builderMu.Unlock()

	return nil
}

func (m *manager) tryToGetLastFetchedTokenList(tokenListID string) (content autofetcher.Content, err error) {
	content, err = m.contentStore.Get(tokenListID)
	if err != nil {
		return
	}
	if len(content.Data) == 0 {
		err = ErrStoredTokenListIsEmpty
		return
	}
	return
}

func (m *manager) mergeList(builder *builder.Builder, tokenListID string, fallbackToInitialList bool) error {
	parser, exists := m.customParsers[tokenListID]
	if !exists {
		// if no custom parser is provided, use the standard parser
		parser = &parsers.StandardTokenListParser{}
	}

	// try to get last fetched main list if available, otherwise use the provided main list
	var (
		content autofetcher.Content
		err     error
	)
	content, err = m.tryToGetLastFetchedTokenList(tokenListID)
	if err != nil {
		if !fallbackToInitialList {
			return err
		}

		// don't return error but instead use the provided initial list
		content.Data, exists = m.initialLists[tokenListID]
		if !exists {
			// this should never happen, because execution gets here only if fallbackToInitialList is true and that's the case
			// for the initial lists (main list and other initial lists) only.
			return ErrNotFoundInInitialLists
		}

		content.SourceURL = LocalSourceURL
		content.Fetched = time.Time{}
	}

	return builder.AddRawTokenList(tokenListID, content.Data, content.SourceURL, content.Fetched, parser)
}

func (m *manager) mergeMainList(builder *builder.Builder) error {
	return m.mergeList(builder, m.mainListID, true)
}

func (m *manager) mergeInitialLists(builder *builder.Builder) error {
	// sort keys for deterministic order, skip main list
	keys := make([]string, 0, len(m.initialLists))
	for key := range m.initialLists {
		if key == m.mainListID {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		err := m.mergeList(builder, key, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) mergeRemoteLists(builder *builder.Builder) error {
	allStoredContent, err := m.contentStore.GetAll()
	if err != nil {
		return err
	}

	// sort keys for deterministic order, skip main list and initial lists
	keys := make([]string, 0, len(allStoredContent))
	for key := range allStoredContent {
		if _, exists := m.initialLists[key]; exists { // main list is also in initial lists
			continue
		}
		if m.remoteListOfTokenListsID != "" && key == m.remoteListOfTokenListsID {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		_ = m.mergeList(builder, key, false) // ignore error, and try to process as many remote lists as possible
	}

	return nil
}

func (m *manager) mergeCustomTokens(builder *builder.Builder) error {
	if m.customTokenStore == nil {
		return nil
	}

	customTokens, err := m.customTokenStore.GetAll()
	if err != nil {
		return err
	}

	customTokenList := &types.TokenList{
		ID:     CustomTokenListID,
		Name:   "Custom tokens",
		Tokens: make([]*types.Token, 0, len(customTokens)),
	}
	for _, token := range customTokens {
		if err := token.Validate(m.chains); err != nil {
			continue
		}
		customTokenList.Tokens = append(customTokenList.Tokens, token)
	}

	builder.AddTokenList(CustomTokenListID, customTokenList)

	return nil
}
