package tokenlists

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/common"

	"go.uber.org/zap"
)

// tokensList implements the TokensList interface with thread-safe state management.
type tokensList struct {
	mu sync.RWMutex

	config        *Config
	refreshWorker *refreshWorker

	state atomic.Pointer[state]

	notifyCh chan struct{}

	started atomic.Bool
}

// NewTokensList creates a new TokensList instance.
func NewTokensList(config *Config) (TokensList, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &tokensList{
		config:        config,
		refreshWorker: newRefreshWorker(config),
	}, nil
}

// Start begins the TokensList service.
func (tl *tokensList) Start(ctx context.Context, notifyCh chan struct{}) error {
	if tl.started.Load() {
		return fmt.Errorf("starting tokens list which has already been started")
	}

	tl.mu.Lock()
	defer tl.mu.Unlock()

	if err := tl.buildState(); err != nil {
		return fmt.Errorf("failed to build initial state: %w", err)
	}

	tl.notifyCh = notifyCh

	if err := tl.manageRefreshWorker(ctx); err != nil {
		return fmt.Errorf("failed to manage refresh worker: %w", err)
	}

	tl.started.Store(true)

	return nil
}

// Stop stops the TokensList service.
func (tl *tokensList) Stop() error {
	if !tl.started.Load() {
		return fmt.Errorf("stopping tokens list which has not been started")
	}

	tl.mu.Lock()
	defer tl.mu.Unlock()

	tl.refreshWorker.stop()

	return nil
}

func (tl *tokensList) manageRefreshWorker(ctx context.Context) error {
	privacyOn, err := tl.config.PrivacyGuard.IsPrivacyOn()
	if err != nil {
		return err
	}
	if privacyOn {
		tl.refreshWorker.stop()
		return nil
	}

	refreshCh := tl.refreshWorker.start(ctx)
	go func() {
		for {
			select {
			case _, ok := <-refreshCh:
				if !ok {
					return
				}
				tl.mu.Lock()
				err := tl.buildState()
				if err != nil {
					tl.config.logger.Error("failed to build state", zap.Error(err))
				} else {
					tl.notifyCh <- struct{}{}
				}
				tl.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// PrivacyModeUpdated is called when privacy mode is updated.
func (tl *tokensList) PrivacyModeUpdated(ctx context.Context) error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if err := tl.manageRefreshWorker(ctx); err != nil {
		return fmt.Errorf("failed to manage refresh worker: %w", err)
	}

	return nil
}

// LastRefreshTime returns the last refresh time.
func (tl *tokensList) LastRefreshTime() (time.Time, error) {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	return tl.config.LastTokenListsUpdateTimeStore.Get()
}

// RefreshNow refreshes the tokens list.
func (tl *tokensList) RefreshNow(ctx context.Context) error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	privacyOn, err := tl.config.PrivacyGuard.IsPrivacyOn()
	if err != nil {
		return err
	}
	if !privacyOn {
		// if privacy mode is off, we need to fetch the remote lists, build the state and and notify the client, all that is done by the manageRefreshWorker.
		if err := tl.manageRefreshWorker(ctx); err != nil {
			return fmt.Errorf("failed to manage refresh worker: %w", err)
		}
		return nil
	}

	// if privacy mode is on, we need to build the state and notify the client.
	if err := tl.buildState(); err != nil {
		return fmt.Errorf("failed to build state: %w", err)
	}
	tl.notifyCh <- struct{}{}

	return nil
}

// UniqueTokens returns all unique tokens.
func (tl *tokensList) UniqueTokens() []*Token {
	state := tl.state.Load()
	if state == nil {
		return nil
	}
	tokens := make([]*Token, 0, len(state.tokens))
	for _, token := range state.tokens {
		tokens = append(tokens, token)
	}
	return tokens
}

// GetTokenByChainAddress retrieves a token by chain ID and address.
func (tl *tokensList) GetTokenByChainAddress(chainID uint64, addr gethcommon.Address) (*Token, bool) {
	state := tl.state.Load()
	if state == nil {
		return nil, false
	}
	key := TokenKey(chainID, addr)
	token, exists := state.tokens[key]
	return token, exists
}

// GetTokensByChain returns all tokens for a specific chain.
func (tl *tokensList) GetTokensByChain(chainID uint64) []*Token {
	state := tl.state.Load()
	if state == nil {
		return nil
	}
	var tokens []*Token
	for _, token := range state.tokens {
		if token.ChainID != chainID {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

// TokenList returns a token list by ID.
func (tl *tokensList) TokenList(id string) (*TokenList, bool) {
	state := tl.state.Load()
	if state == nil {
		return nil, false
	}
	tokenList, exists := state.tokenLists[id]
	return tokenList, exists
}

// TokenLists returns all token lists.
func (tl *tokensList) TokenLists() []*TokenList {
	state := tl.state.Load()
	if state == nil {
		return nil
	}
	tokenLists := make([]*TokenList, 0, len(state.tokenLists))
	for _, tokenList := range state.tokenLists {
		tokenLists = append(tokenLists, tokenList)
	}
	return tokenLists
}

func (tl *tokensList) buildState() error {
	newState := &state{
		tokens:     make(map[string]*Token),
		tokenLists: make(map[string]*TokenList),
	}

	// 1. native token list
	if err := tl.mergeNativeTokenList(newState); err != nil {
		tl.config.logger.Error("failed to merge native token list", zap.Error(err))
	}

	// merge tokens from all sources in the specified order.
	// 2. main list (remote if available, otherwise initial)
	if err := tl.mergeMainList(newState); err != nil {
		tl.config.logger.Error("failed to merge main list", zap.Error(err))
	}

	// 3. other initial lists (in deterministic order), remote if available, otherwise initial list
	if err := tl.mergeInitialLists(newState); err != nil {
		tl.config.logger.Error("failed to merge initial lists", zap.Error(err))
	}

	// 4. remote lists that are not main or initial lists (in deterministic order)
	if err := tl.mergeRemoteLists(newState); err != nil {
		tl.config.logger.Error("failed to merge remote lists", zap.Error(err))
	}

	// 5. custom tokens
	if err := tl.mergeCustomTokens(newState); err != nil {
		tl.config.logger.Error("failed to merge custom tokens", zap.Error(err))
	}

	tl.state.Store(newState)
	return nil
}

func getNativeToken(chainID uint64) *Token {
	crossChainID := EthereumNativeCrossChainID
	symbol := EthereumNativeSymbol
	name := EthereumNativeName
	logoURI := "https://raw.githubusercontent.com/trustwallet/assets/master/blockchains/ethereum/assets/0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2/logo.png"
	if chainID == common.BSCMainnet || chainID == common.BSCTestnet {
		crossChainID = BinanceSmartChainNativeCrossChainID
		symbol = BinanceSmartChainNativeSymbol
		name = BinanceSmartChainNativeName
		logoURI = "https://assets.coingecko.com/coins/images/825/thumb/bnb-icon2_2x.png?1696501970"
	}
	return &Token{
		CrossChainID: crossChainID,
		ChainID:      chainID,
		Symbol:       symbol,
		Name:         name,
		Decimals:     18,
		LogoURI:      logoURI,
	}
}

func (tl *tokensList) mergeNativeTokenList(state *state) error {
	nativeTokenList := &TokenList{
		Name:   "Native tokens",
		Tokens: make([]*Token, 0),
	}

	for _, chainID := range tl.config.Chains {
		nativeToken := getNativeToken(chainID)
		nativeTokenList.Tokens = append(nativeTokenList.Tokens, nativeToken)
	}

	tl.addTokenListToState(state, NativeTokenListID, nativeTokenList)
	return nil
}

func (tl *tokensList) mergeMainList(state *state) error {
	parser, exists := tl.config.Parsers[tl.config.MainListID]
	if !exists {
		// because we validate config in NewTokensList, this should never happen
		return fmt.Errorf("main list parser not found for list ID %s", tl.config.MainListID)
	}

	// process last fetched main list
	storedContent, err := tl.config.ContentStore.Get(tl.config.MainListID)
	if err != nil {
		tl.config.logger.Error("failed to get stored content for main list", zap.Error(err))
		goto processProvidedMainList
	}

	if len(storedContent.Data) > 0 {
		tokenList, err := parser.Parse(storedContent.Data, storedContent.SourceURL, storedContent.Fetched, tl.config.Chains)
		if err != nil {
			tl.config.logger.Error("failed to parse main list", zap.Error(err))
			goto processProvidedMainList
		}
		tl.addTokenListToState(state, tl.config.MainListID, tokenList)
		return nil
	}

	tl.config.logger.Info("main list not found in content store")

processProvidedMainList: // process provided main list
	if tl.config.MainList != nil {
		if tokenList, err := parser.Parse(tl.config.MainList, LocalSourceURL, time.Time{}, tl.config.Chains); err == nil {
			tl.addTokenListToState(state, tl.config.MainListID, tokenList)
		}
	}

	return nil
}

func (tl *tokensList) mergeInitialLists(state *state) error {
	// sort keys for deterministic order
	keys := make([]string, 0, len(tl.config.InitialLists))
	for key := range tl.config.InitialLists {
		if key != tl.config.MainListID {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	for _, key := range keys {
		data := tl.config.InitialLists[key]
		parser, exists := tl.config.Parsers[key]
		if !exists {
			// because we validate config in NewTokensList, this should never happen
			tl.config.logger.Error("initial list parser not found for list ID", zap.String("listID", key))
			continue
		}

		// process last fetched list
		storedContent, err := tl.config.ContentStore.Get(key)
		if err != nil {
			tl.config.logger.Error("failed to get stored content for initial list", zap.Error(err))
			goto processProvidedList
		}
		if len(storedContent.Data) > 0 {
			tokenList, err := parser.Parse(storedContent.Data, storedContent.SourceURL, storedContent.Fetched, tl.config.Chains)
			if err != nil {
				tl.config.logger.Error("failed to parse initial list", zap.Error(err))
				goto processProvidedList
			}
			tl.addTokenListToState(state, key, tokenList)
			continue
		}

		tl.config.logger.Info("initial list not found in content store", zap.String("listID", key))

	processProvidedList: // process provided list
		if tokens, err := parser.Parse(data, LocalSourceURL, time.Time{}, tl.config.Chains); err == nil {
			tl.addTokenListToState(state, key, tokens)
		}
	}

	return nil
}

func (tl *tokensList) mergeRemoteLists(state *state) error {
	allStoredContent, err := tl.config.ContentStore.GetAll()
	if err != nil {
		return err
	}

	// sort keys for deterministic order
	keys := make([]string, 0, len(allStoredContent))
	for key := range allStoredContent {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, contentID := range keys {
		storedContent, ok := allStoredContent[contentID]
		if !ok {
			continue
		}

		if contentID == tl.config.MainListID {
			continue
		}

		if _, ok := tl.config.InitialLists[contentID]; ok {
			continue
		}

		parser, exists := tl.config.Parsers[contentID]
		if !exists {
			tl.config.logger.Error("remote list parser not found for list ID", zap.String("listID", contentID))
			continue
		}

		if tokenList, err := parser.Parse(storedContent.Data, storedContent.SourceURL, storedContent.Fetched, tl.config.Chains); err == nil {
			tl.addTokenListToState(state, contentID, tokenList)
		}
	}

	return nil
}

func (tl *tokensList) mergeCustomTokens(state *state) error {
	if tl.config.CustomTokenStore == nil {
		return nil
	}
	customTokens, err := tl.config.CustomTokenStore.GetAll()
	if err != nil {
		return err
	}

	customTokenList := &TokenList{
		Name:   "Custom tokens",
		Tokens: make([]*Token, 0, len(customTokens)),
	}
	for _, token := range customTokens {
		if err := validateToken(token, tl.config.Chains); err != nil {
			tl.config.logger.Error("invalid token", zap.String("symbol", token.Symbol), zap.Error(err))
			continue
		}
		customTokenList.Tokens = append(customTokenList.Tokens, token)
	}

	tl.addTokenListToState(state, CustomTokenListID, customTokenList)

	return nil
}

func (tl *tokensList) addTokenListToState(currentState *state, tokenListID string, tokenList *TokenList) {
	currentState.tokenLists[tokenListID] = tokenList
	for _, token := range tokenList.Tokens {
		if _, exists := currentState.tokens[token.Key()]; !exists {
			currentState.tokens[token.Key()] = token
		}
	}
}
