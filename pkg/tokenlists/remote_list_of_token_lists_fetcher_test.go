package tokenlists

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTokenListsFetcher_ResolveListOfTokenLists(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:       zap.NewNop(),
		ContentStore: NewDefaultContentStore(),
	}

	fetcher := newTokenListsFetcher(config)

	expectedListOfTokenListsJsonResponse := strings.ReplaceAll(listOfTokenListsJsonResponse, serverURLPlaceholder, server.URL)
	expectedListOfTokenListsCopy := copyRemoteListOfTokenLists(expectedListOfTokenLists)
	for i := range expectedListOfTokenListsCopy.TokenLists {
		tokenList := &expectedListOfTokenListsCopy.TokenLists[i]
		tokenList.SourceURL = strings.ReplaceAll(tokenList.SourceURL, serverURLPlaceholder, server.URL)
		tokenList.Schema = strings.ReplaceAll(tokenList.Schema, serverURLPlaceholder, server.URL)
	}

	expectedListOfTokenListsJsonResponse1 := strings.ReplaceAll(listOfTokenListsJsonResponse1, serverURLPlaceholder, server.URL)
	expectedListOfTokenListsCopy1 := copyRemoteListOfTokenLists(expectedListOfTokenLists1)
	for i := range expectedListOfTokenListsCopy1.TokenLists {
		tokenList := &expectedListOfTokenListsCopy1.TokenLists[i]
		tokenList.SourceURL = strings.ReplaceAll(tokenList.SourceURL, serverURLPlaceholder, server.URL)
		tokenList.Schema = strings.ReplaceAll(tokenList.Schema, serverURLPlaceholder, server.URL)
	}

	expectedListOfTokenListsJsonResponse2 := strings.ReplaceAll(listOfTokenListsJsonResponse2, serverURLPlaceholder, server.URL)
	expectedListOfTokenListsCopy2 := copyRemoteListOfTokenLists(expectedListOfTokenLists2)
	for i := range expectedListOfTokenListsCopy2.TokenLists {
		tokenList := &expectedListOfTokenListsCopy2.TokenLists[i]
		tokenList.SourceURL = strings.ReplaceAll(tokenList.SourceURL, serverURLPlaceholder, server.URL)
		tokenList.Schema = strings.ReplaceAll(tokenList.Schema, serverURLPlaceholder, server.URL)
	}

	// no stored content/etag
	content, err := fetcher.config.ContentStore.Get(StatusListOfTokenListsID)
	require.Error(t, err)
	require.Equal(t, "", content.Etag)
	require.Equal(t, "", string(content.Data))

	// fetch remote list of token lists with no tag
	config.RemoteListOfTokenListsURL = server.URL + listOfTokenListsURL
	tokenLists, err := fetcher.resolveListOfTokenLists(context.TODO())
	require.NoError(t, err)
	require.Equal(t, expectedListOfTokenListsCopy, tokenLists)

	// check stored content/etag
	content, err = fetcher.config.ContentStore.Get(StatusListOfTokenListsID)
	require.NoError(t, err)
	require.Equal(t, "", content.Etag)
	require.Equal(t, expectedListOfTokenListsJsonResponse, string(content.Data))

	// fetch remote list of token lists with etag
	config.RemoteListOfTokenListsURL = server.URL + listOfTokenListsWithEtagURL
	tokenLists, err = fetcher.resolveListOfTokenLists(context.TODO())
	require.NoError(t, err)
	require.Equal(t, expectedListOfTokenListsCopy1, tokenLists)

	// check stored content/etag
	content, err = fetcher.config.ContentStore.Get(StatusListOfTokenListsID)
	require.NoError(t, err)
	require.Equal(t, listOfTokenListsEtag, content.Etag)
	require.Equal(t, expectedListOfTokenListsJsonResponse1, string(content.Data))

	// fetch remote list of token lists with the same etag
	config.RemoteListOfTokenListsURL = server.URL + listOfTokenListsWithSameEtagURL
	tokenLists, err = fetcher.resolveListOfTokenLists(context.TODO())
	require.NoError(t, err)
	require.Equal(t, expectedListOfTokenListsCopy1, tokenLists)

	// check stored content/etag
	content, err = fetcher.config.ContentStore.Get(StatusListOfTokenListsID)
	require.NoError(t, err)
	require.Equal(t, listOfTokenListsEtag, content.Etag)
	require.Equal(t, expectedListOfTokenListsJsonResponse1, string(content.Data))

	// fetch remote list of token lists with a new etag
	config.RemoteListOfTokenListsURL = server.URL + listOfTokenListsWithNewEtagURL
	tokenLists, err = fetcher.resolveListOfTokenLists(context.TODO())
	require.NoError(t, err)
	require.Equal(t, expectedListOfTokenListsCopy2, tokenLists)

	// check stored content/etag, should be the new one
	content, err = fetcher.config.ContentStore.Get(StatusListOfTokenListsID)
	require.NoError(t, err)
	require.Equal(t, listOfTokenListsNewEtag, content.Etag)
	require.Equal(t, expectedListOfTokenListsJsonResponse2, string(content.Data))
}

func TestTokenListsFetcher_ResolveListOfTokenLists_WithIssues(t *testing.T) {

	expectedListOfTokenListsCopy := copyRemoteListOfTokenLists(expectedListOfTokenLists)
	for i := range expectedListOfTokenListsCopy.TokenLists {
		tokenList := &expectedListOfTokenListsCopy.TokenLists[i]
		tokenList.SourceURL = strings.ReplaceAll(tokenList.SourceURL, serverURLPlaceholder, serverURLPlaceholder)
		tokenList.Schema = strings.ReplaceAll(tokenList.Schema, serverURLPlaceholder, serverURLPlaceholder)
	}

	var tests = []struct {
		name          string
		remoteURL     string
		storedContent *Content
		expected      remoteListOfTokenLists
	}{
		{
			name:      "page not found no stored content",
			remoteURL: "/not-found",
		},
		{
			name:      "page not found has stored content",
			remoteURL: "/not-found",
			storedContent: &Content{
				SourceURL: "",
				Etag:      "some-etag",
				Data:      []byte(listOfTokenListsJsonResponse),
				Fetched:   time.Now(),
			},
			expected: expectedListOfTokenListsCopy,
		},
		{
			name:      "content of the response does not match the schema",
			remoteURL: listOfTokenListsWrongSchemaURL,
			expected:  expectedListOfTokenListsCopy,
		},
		{
			name:     "remote url not set",
			expected: expectedListOfTokenListsCopy,
		},
	}

	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:       zap.NewNop(),
		ContentStore: NewDefaultContentStore(),
	}

	fetcher := newTokenListsFetcher(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			config.RemoteListOfTokenListsURL = ""
			if tt.remoteURL != "" {
				config.RemoteListOfTokenListsURL = server.URL + tt.remoteURL
			}

			if tt.storedContent != nil {
				err := config.ContentStore.Set(StatusListOfTokenListsID, *tt.storedContent)
				require.NoError(t, err)
			}

			tokenLists, err := fetcher.resolveListOfTokenLists(context.TODO())

			require.NoError(t, err)
			require.Equal(t, tt.expected, tokenLists)
		})
	}
}

func TestTokenListsFetcher_ResolveListOfTokenLists_ContextCancellation(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:                    zap.NewNop(),
		ContentStore:              NewDefaultContentStore(),
		RemoteListOfTokenListsURL: server.URL + delayedResponseURL,
	}

	fetcher := newTokenListsFetcher(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := fetcher.resolveListOfTokenLists(ctx)
	assert.NoError(t, err) // Should not return an error - the function is designed to be resilient
}
