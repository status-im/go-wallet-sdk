package builder

import (
	"errors"
	"testing"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/status-im/go-wallet-sdk/pkg/common"
	mock_parsers "github.com/status-im/go-wallet-sdk/pkg/tokens/parsers/mock"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	testChains = []uint64{common.EthereumMainnet, common.BSCMainnet, common.OptimismMainnet}

	testToken1 = &types.Token{
		CrossChainID: "test-token-1",
		ChainID:      common.EthereumMainnet,
		Address:      gethcommon.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		Decimals:     18,
		Name:         "Test Token 1",
		Symbol:       "TT1",
		LogoURI:      "https://example.com/token1.png",
	}

	testToken2 = &types.Token{
		CrossChainID: "test-token-2",
		ChainID:      common.BSCMainnet,
		Address:      gethcommon.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12"),
		Decimals:     8,
		Name:         "Test Token 2",
		Symbol:       "TT2",
		LogoURI:      "https://example.com/token2.png",
	}

	testTokenList1 = &types.TokenList{
		Name:      "Test Token List 1",
		Timestamp: "2025-01-01T00:00:00Z",
		Source:    "https://example.com/list1.json",
		Version:   types.Version{Major: 1, Minor: 0, Patch: 0},
		Tokens:    []*types.Token{testToken1},
	}

	testTokenList2 = &types.TokenList{
		Name:      "Test Token List 2",
		Timestamp: "2025-01-02T00:00:00Z",
		Source:    "https://example.com/list2.json",
		Version:   types.Version{Major: 2, Minor: 1, Patch: 0},
		Tokens:    []*types.Token{testToken2},
	}
)

func TestNew(t *testing.T) {
	chains := []uint64{common.EthereumMainnet, common.BSCMainnet}

	builder := New(chains)

	assert.NotNil(t, builder)
	assert.Equal(t, chains, builder.chains)
	assert.NotNil(t, builder.tokens)
	assert.NotNil(t, builder.tokenLists)
	assert.Empty(t, builder.tokens)
	assert.Empty(t, builder.tokenLists)
}

func TestBuilder_GetTokens(t *testing.T) {
	builder := New(testChains)
	builder.AddTokenList("test-list", testTokenList1)

	result := builder.GetTokens()
	assert.Contains(t, result, testToken1.Key())
}

func TestBuilder_GetTokenLists(t *testing.T) {
	builder := New(testChains)
	builder.AddTokenList("test-list", testTokenList1)

	result := builder.GetTokenLists()
	assert.Contains(t, result, "test-list")
}

func TestGetNativeToken(t *testing.T) {
	tests := []struct {
		name            string
		chainID         uint64
		expectedSymbol  string
		expectedName    string
		expectedCrossID string
	}{
		{
			name:            "Ethereum mainnet",
			chainID:         common.EthereumMainnet,
			expectedSymbol:  EthereumNativeSymbol,
			expectedName:    EthereumNativeName,
			expectedCrossID: EthereumNativeCrossChainID,
		},
		{
			name:            "BSC mainnet",
			chainID:         common.BSCMainnet,
			expectedSymbol:  BinanceSmartChainNativeSymbol,
			expectedName:    BinanceSmartChainNativeName,
			expectedCrossID: BinanceSmartChainNativeCrossChainID,
		},
		{
			name:            "Sepolia",
			chainID:         common.EthereumSepolia,
			expectedSymbol:  EthereumNativeSymbol,
			expectedName:    EthereumNativeName,
			expectedCrossID: EthereumNativeCrossChainID,
		},
		{
			name:            "BSC testnet",
			chainID:         common.BSCTestnet,
			expectedSymbol:  BinanceSmartChainNativeSymbol,
			expectedName:    BinanceSmartChainNativeName,
			expectedCrossID: BinanceSmartChainNativeCrossChainID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := getNativeToken(tt.chainID)

			assert.Equal(t, tt.expectedCrossID, token.CrossChainID)
			assert.Equal(t, tt.chainID, token.ChainID)
			assert.Equal(t, tt.expectedSymbol, token.Symbol)
			assert.Equal(t, tt.expectedName, token.Name)
			assert.Equal(t, uint(18), token.Decimals)
			assert.NotEmpty(t, token.LogoURI)
			assert.True(t, token.IsNative())
		})
	}
}

func TestBuilder_AddNativeTokenList(t *testing.T) {
	chains := []uint64{common.EthereumMainnet, common.BSCMainnet, common.OptimismMainnet}
	builder := New(chains)

	err := builder.AddNativeTokenList()
	require.NoError(t, err)

	tokenLists := builder.GetTokenLists()
	assert.Contains(t, tokenLists, NativeTokenListID)

	nativeList := tokenLists[NativeTokenListID]
	assert.Equal(t, "Native tokens", nativeList.Name)
	assert.Len(t, nativeList.Tokens, len(chains))

	tokens := builder.GetTokens()
	assert.Len(t, tokens, len(chains))

	chainTokenMap := make(map[uint64]*types.Token)
	for _, token := range nativeList.Tokens {
		chainTokenMap[token.ChainID] = token
		assert.True(t, token.IsNative())
		assert.Contains(t, tokens, token.Key())
	}

	ethToken := chainTokenMap[common.EthereumMainnet]
	assert.Equal(t, EthereumNativeSymbol, ethToken.Symbol)
	assert.Equal(t, EthereumNativeCrossChainID, ethToken.CrossChainID)

	bscToken := chainTokenMap[common.BSCMainnet]
	assert.Equal(t, BinanceSmartChainNativeSymbol, bscToken.Symbol)
	assert.Equal(t, BinanceSmartChainNativeCrossChainID, bscToken.CrossChainID)
}

func TestBuilder_AddTokenList(t *testing.T) {
	builder := New(testChains)

	tokenListID := "test-list"
	builder.AddTokenList(tokenListID, testTokenList1)

	tokenLists := builder.GetTokenLists()
	assert.Contains(t, tokenLists, tokenListID)
	assert.Equal(t, testTokenList1, tokenLists[tokenListID])

	tokens := builder.GetTokens()
	for _, token := range testTokenList1.Tokens {
		assert.Contains(t, tokens, token.Key())
		assert.Equal(t, token, tokens[token.Key()])
	}
}

func TestBuilder_AddTokenList_DuplicateTokens(t *testing.T) {
	builder := New(testChains)

	tokenList1 := &types.TokenList{
		Name:   "List 1",
		Tokens: []*types.Token{testToken1},
	}
	tokenList2 := &types.TokenList{
		Name:   "List 2",
		Tokens: []*types.Token{testToken1}, // same token
	}

	builder.AddTokenList("list1", tokenList1)
	builder.AddTokenList("list2", tokenList2)

	tokenLists := builder.GetTokenLists()
	assert.Len(t, tokenLists, 2)
	assert.Contains(t, tokenLists, "list1")
	assert.Contains(t, tokenLists, "list2")

	tokens := builder.GetTokens()
	assert.Len(t, tokens, 1)
	assert.Contains(t, tokens, testToken1.Key())
}

func TestBuilder_AddRawTokenList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := New(testChains)

	rawData := []byte(`{"name": "Test List", "tokens": []}`)
	sourceURL := "https://example.com/test.json"
	fetchedAt := time.Now()

	mockParser := mock_parsers.NewMockTokenListParser(ctrl)
	mockParser.EXPECT().Parse(rawData, testChains).Return(testTokenList1, nil)

	err := builder.AddRawTokenList("test-list", rawData, sourceURL, fetchedAt, mockParser)
	require.NoError(t, err)

	tokenLists := builder.GetTokenLists()
	assert.Contains(t, tokenLists, "test-list")
	assert.Equal(t, testTokenList1, tokenLists["test-list"])

	tokens := builder.GetTokens()
	for _, token := range testTokenList1.Tokens {
		assert.Contains(t, tokens, token.Key())
	}
}

func TestBuilder_AddRawTokenList_EmptyRawData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := New(testChains)
	mockParser := mock_parsers.NewMockTokenListParser(ctrl)

	err := builder.AddRawTokenList("test-list", []byte{}, "url", time.Now(), mockParser)
	assert.ErrorIs(t, err, ErrEmptyRawTokenList)

	err = builder.AddRawTokenList("test-list", nil, "url", time.Now(), mockParser)
	assert.ErrorIs(t, err, ErrEmptyRawTokenList)
}

func TestBuilder_AddRawTokenList_NilParser(t *testing.T) {
	builder := New(testChains)

	err := builder.AddRawTokenList("test-list", []byte(`{}`), "url", time.Now(), nil)
	assert.ErrorIs(t, err, ErrParserIsNil)
}

func TestBuilder_AddRawTokenList_ParserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := New(testChains)

	expectedError := errors.New("parser error")
	rawData := []byte(`{}`)

	mockParser := mock_parsers.NewMockTokenListParser(ctrl)
	mockParser.EXPECT().Parse(rawData, testChains).Return(nil, expectedError)

	err := builder.AddRawTokenList("test-list", rawData, "url", time.Now(), mockParser)
	assert.ErrorIs(t, err, expectedError)

	tokenLists := builder.GetTokenLists()
	assert.NotContains(t, tokenLists, "test-list")
}

func TestBuilder_ComplexBuildScenario(t *testing.T) {
	builder := New([]uint64{common.EthereumMainnet, common.BSCMainnet})

	err := builder.AddNativeTokenList()
	require.NoError(t, err)

	builder.AddTokenList("list1", testTokenList1)

	builder.AddTokenList("list2", testTokenList2)

	duplicateTokenList := &types.TokenList{
		Name:   "Duplicate List",
		Tokens: []*types.Token{testToken1}, // same as in list1
	}
	builder.AddTokenList("duplicate", duplicateTokenList)

	tokenLists := builder.GetTokenLists()
	assert.Len(t, tokenLists, 4) // native + list1 + list2 + duplicate

	tokens := builder.GetTokens()
	// Should have: 2 native tokens + testToken1 + testToken2 = 4 unique tokens
	assert.Len(t, tokens, 4)

	assert.Contains(t, tokens, testToken1.Key())
	assert.Contains(t, tokens, testToken2.Key())

	ethNative := getNativeToken(common.EthereumMainnet)
	bscNative := getNativeToken(common.BSCMainnet)
	assert.Contains(t, tokens, ethNative.Key())
	assert.Contains(t, tokens, bscNative.Key())
}

func TestBuilder_EmptyChains(t *testing.T) {
	builder := New([]uint64{})

	err := builder.AddNativeTokenList()
	require.NoError(t, err)

	tokenLists := builder.GetTokenLists()
	assert.Contains(t, tokenLists, NativeTokenListID)

	nativeList := tokenLists[NativeTokenListID]
	assert.Equal(t, "Native tokens", nativeList.Name)
	assert.Empty(t, nativeList.Tokens)

	tokens := builder.GetTokens()
	assert.Empty(t, tokens)
}

func TestBuilder_API(t *testing.T) {
	builder := New([]uint64{common.EthereumMainnet})

	err := builder.AddNativeTokenList()
	require.NoError(t, err)

	builder.AddTokenList("list1", testTokenList1)
	builder.AddTokenList("list2", testTokenList2)

	tokens := builder.GetTokens()
	assert.NotEmpty(t, tokens)

	tokenLists := builder.GetTokenLists()
	assert.Len(t, tokenLists, 3)
}

func TestBuilder_BuilderPattern_EmptyInitialization(t *testing.T) {
	builder := New(testChains)

	assert.Empty(t, builder.GetTokens())
	assert.Empty(t, builder.GetTokenLists())

	builder.AddTokenList("first", testTokenList1)
	assert.Len(t, builder.GetTokens(), 1)
	assert.Len(t, builder.GetTokenLists(), 1)

	builder.AddTokenList("second", testTokenList2)
	assert.Len(t, builder.GetTokens(), 2)
	assert.Len(t, builder.GetTokenLists(), 2)

	err := builder.AddNativeTokenList()
	require.NoError(t, err)
	assert.Len(t, builder.GetTokens(), 2+len(testChains))
	assert.Len(t, builder.GetTokenLists(), 3)
}
