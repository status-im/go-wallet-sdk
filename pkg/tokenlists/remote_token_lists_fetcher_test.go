package tokenlists

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTokenListsFetcher_FetchTokenList(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger: zap.NewNop(),
	}

	fetcher := newTokenListsFetcher(config)

	tests := []struct {
		name          string
		list          tokenList
		etag          string
		expectedError bool
		expectedEtag  string
		expectedData  []byte
	}{
		{
			name: "successful fetch without schema without etag",
			list: tokenList{
				ID:        "test-list",
				SourceURL: server.URL + uniswapURL,
				Schema:    "",
			},
			etag:          "",
			expectedError: false,
			expectedEtag:  "",
			expectedData:  []byte(uniswapTokenListJsonResponse),
		},
		{
			name: "successful fetch with wrong schema without etag",
			list: tokenList{
				ID:        "test-list",
				SourceURL: server.URL + uniswapURL,
				Schema:    "wrong-schema.json",
			},
			etag:          "",
			expectedError: true,
			expectedEtag:  "",
			expectedData:  nil,
		},
		{
			name: "successful fetch with right schema without etag",
			list: tokenList{
				ID:        "test-list",
				SourceURL: server.URL + uniswapURL,
				Schema:    server.URL + uniswapSchemaURL,
			},
			etag:          "",
			expectedError: false,
			expectedEtag:  "",
			expectedData:  []byte(uniswapTokenListJsonResponse),
		},
		{
			name: "successful fetch with etag",
			list: tokenList{
				ID:        "test-list",
				SourceURL: server.URL + uniswapWithEtagURL,
				Schema:    "",
			},
			etag:          "",
			expectedError: false,
			expectedEtag:  uniswapEtag,
			expectedData:  []byte(uniswapTokenListJsonResponse1),
		},
		{
			name: "successful fetch with the same etag",
			list: tokenList{
				ID:        "test-list",
				SourceURL: server.URL + uniswapSameEtagURL,
				Schema:    "",
			},
			etag:          uniswapEtag,
			expectedError: false,
			expectedEtag:  uniswapEtag,
			expectedData:  nil,
		},
		{
			name: "successful fetch with the new etag",
			list: tokenList{
				ID:        "test-list",
				SourceURL: server.URL + uniswapNewEtagURL,
				Schema:    "",
			},
			etag:          uniswapEtag,
			expectedError: false,
			expectedEtag:  uniswapNewEtag,
			expectedData:  []byte(uniswapTokenListJsonResponse2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan fetchedTokenList, 1)
			defer close(ch)

			err := fetcher.fetchTokenList(context.Background(), tt.list, tt.etag, ch)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Check if we got data (for non-304 responses)
			if tt.expectedData != nil {
				select {
				case fetched := <-ch:
					assert.Equal(t, tt.list.ID, fetched.ID)
					assert.Equal(t, tt.list.SourceURL, fetched.SourceURL)
					assert.Equal(t, tt.list.Schema, fetched.Schema)
					assert.Equal(t, tt.expectedEtag, fetched.Etag)
					assert.Equal(t, tt.expectedData, fetched.JsonData)
					assert.WithinDuration(t, time.Now(), fetched.Fetched, 2*time.Second)
				case <-time.After(1 * time.Second):
					t.Fatal("timeout waiting for fetched token list")
				}
			} else {
				select {
				case <-ch:
					t.Fatal("unexpected data received for 304 response")
				case <-time.After(100 * time.Millisecond):
					// This is expected - no data should be sent
				}
			}
		})
	}
}

func TestTokenListsFetcher_FetchTokenList_ContextCancellation(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:       zap.NewNop(),
		ContentStore: NewDefaultContentStore(),
	}

	fetcher := newTokenListsFetcher(config)

	list := tokenList{
		ID:        "slow-response",
		SourceURL: server.URL + delayedResponseURL,
		Schema:    "",
	}

	ch := make(chan fetchedTokenList, 1)
	defer close(ch)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := fetcher.fetchTokenList(ctx, list, "", ch)
	assert.Error(t, err)
}

func TestTokenListsFetcher_FetchTokenList_ChannelClosed(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	config := &Config{
		logger:       zap.NewNop(),
		ContentStore: NewDefaultContentStore(),
	}

	fetcher := newTokenListsFetcher(config)

	list := tokenList{
		ID:        "status",
		SourceURL: server.URL + uniswapURL,
		Schema:    "",
	}

	// Create a closed channel
	ch := make(chan fetchedTokenList)
	close(ch)
	ch = nil

	err := fetcher.fetchTokenList(context.Background(), list, "", ch)
	assert.Error(t, err)
}
