package fetcher_test

import (
	"context"
	"testing"
	"time"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"

	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	f := fetcher.New(fetcher.DefaultConfig())

	tests := []struct {
		name          string
		list          fetcher.FetchDetails
		expectedError bool
		expectedEtag  string
		expectedData  []byte
	}{
		{
			name: "successful fetch without schema without etag",
			list: fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "test-list",
					SourceURL: server.URL + uniswapURL,
					Schema:    "",
				},
				Etag: "",
			},
			expectedError: false,
			expectedEtag:  "",
			expectedData:  []byte(uniswapTokenListJsonResponse),
		},
		{
			name: "successful fetch with wrong schema without etag",
			list: fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "test-list",
					SourceURL: server.URL + uniswapURL,
					Schema:    server.URL + wrongSchemaURL,
				},
				Etag: "",
			},
			expectedError: true,
			expectedEtag:  "",
			expectedData:  []byte(uniswapTokenListJsonResponse),
		},
		{
			name: "successful fetch with right schema without etag",
			list: fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "test-list",
					SourceURL: server.URL + uniswapURL,
					Schema:    server.URL + uniswapSchemaURL,
				},
				Etag: "",
			},
			expectedError: false,
			expectedEtag:  "",
			expectedData:  []byte(uniswapTokenListJsonResponse),
		},
		{
			name: "successful fetch with etag",
			list: fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "test-list",
					SourceURL: server.URL + uniswapWithEtagURL,
					Schema:    "",
				},
				Etag: "",
			},
			expectedError: false,
			expectedEtag:  uniswapEtag,
			expectedData:  []byte(uniswapTokenListJsonResponse1),
		},
		{
			name: "successful fetch with the same etag",
			list: fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "test-list",
					SourceURL: server.URL + uniswapSameEtagURL,
					Schema:    "",
				},
				Etag: uniswapEtag,
			},
			expectedError: false,
			expectedEtag:  uniswapEtag,
			expectedData:  nil,
		},
		{
			name: "successful fetch with the new etag",
			list: fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "test-list",
					SourceURL: server.URL + uniswapNewEtagURL,
					Schema:    "",
				},
				Etag: uniswapEtag,
			},
			expectedError: false,
			expectedEtag:  uniswapNewEtag,
			expectedData:  []byte(uniswapTokenListJsonResponse2),
		},
		{
			name: "fetch from non-existent URL returns error",
			list: fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "test-list",
					SourceURL: server.URL + "/non-existent.json",
					Schema:    "",
				},
				Etag: uniswapEtag,
			},
			expectedError: true,
			expectedEtag:  uniswapEtag,
			expectedData:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetched, err := f.Fetch(context.Background(), tt.list)

			assert.Equal(t, tt.list.ID, fetched.ID)
			assert.Equal(t, tt.list.SourceURL, fetched.SourceURL)
			assert.Equal(t, tt.list.Schema, fetched.Schema)
			assert.Equal(t, tt.expectedEtag, fetched.Etag)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Error(t, fetched.Error)
				assert.Nil(t, fetched.JsonData)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, fetched.Error)

			if tt.expectedData != nil {
				assert.Equal(t, tt.expectedData, fetched.JsonData)
				assert.WithinDuration(t, time.Now(), fetched.Fetched, 2*time.Second)
			} else {
				assert.Nil(t, fetched.JsonData)
			}
		})
	}
}

func TestFetch_ContextCancellation(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	f := fetcher.New(fetcher.DefaultConfig())

	list := fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			ID:        "slow-response",
			SourceURL: server.URL + delayedResponseURL,
			Schema:    "",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fetched, err := f.Fetch(ctx, list)

	assert.Error(t, err)
	assert.Error(t, fetched.Error)
	assert.Contains(t, fetched.Error.Error(), "context")
}

func TestFetchConcurrent(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	f := fetcher.New(fetcher.DefaultConfig())

	tokenLists := []fetcher.FetchDetails{
		{
			ListDetails: types.ListDetails{
				ID:        "uniswap",
				SourceURL: server.URL + uniswapURL,
				Schema:    "",
			},
			Etag: "",
		},
		{
			ListDetails: types.ListDetails{
				ID:        "uniswap-with-etag",
				SourceURL: server.URL + uniswapWithEtagURL,
				Schema:    "",
			},
			Etag: "",
		},
		{
			ListDetails: types.ListDetails{
				ID:        "uniswap-with-schema",
				SourceURL: server.URL + uniswapURL,
				Schema:    server.URL + uniswapSchemaURL,
			},
			Etag: "",
		},
		{
			ListDetails: types.ListDetails{
				ID:        "uniswap-with-wrong-schema",
				SourceURL: server.URL + uniswapURL,
				Schema:    server.URL + wrongSchemaURL,
			},
			Etag: "",
		},
		{
			ListDetails: types.ListDetails{
				ID:        "uniswap-same-etag",
				SourceURL: server.URL + uniswapSameEtagURL,
				Schema:    "",
			},
			Etag: uniswapEtag,
		},
		{
			ListDetails: types.ListDetails{
				ID:        "uniswap-new-etag",
				SourceURL: server.URL + uniswapNewEtagURL,
				Schema:    "",
			},
			Etag: uniswapEtag,
		},
		{
			ListDetails: types.ListDetails{
				ID:        "invalid-url",
				SourceURL: "invalid-url",
				Schema:    "",
			},
			Etag: "",
		},
		{
			ListDetails: types.ListDetails{
				ID:        "non-existent",
				SourceURL: server.URL + "/non-existent.json",
				Schema:    "",
			},
			Etag: "",
		},
	}

	fetchedLists, err := f.FetchConcurrent(context.Background(), tokenLists)

	assert.NoError(t, err)
	assert.Len(t, fetchedLists, 8)

	// Check that we got results for all token lists (some with errors, some without)
	ids := make(map[string]bool)
	for _, fetched := range fetchedLists {
		ids[fetched.ID] = true

		switch fetched.ID {
		case "uniswap":
			assert.NoError(t, fetched.Error)
			assert.NotNil(t, fetched.JsonData)
			assert.Equal(t, []byte(uniswapTokenListJsonResponse), fetched.JsonData)
		case "uniswap-with-etag":
			assert.NoError(t, fetched.Error)
			assert.NotNil(t, fetched.JsonData)
			assert.Equal(t, uniswapEtag, fetched.Etag)
			assert.Equal(t, []byte(uniswapTokenListJsonResponse1), fetched.JsonData)
		case "uniswap-with-schema":
			assert.NoError(t, fetched.Error)
			assert.NotNil(t, fetched.JsonData)
			assert.Equal(t, []byte(uniswapTokenListJsonResponse), fetched.JsonData)
		case "uniswap-with-wrong-schema":
			assert.Error(t, fetched.Error)
			assert.Nil(t, fetched.JsonData)
		case "uniswap-same-etag":
			assert.NoError(t, fetched.Error)
			assert.Nil(t, fetched.JsonData)
			assert.Equal(t, uniswapEtag, fetched.Etag)
		case "uniswap-new-etag":
			assert.NoError(t, fetched.Error)
			assert.NotNil(t, fetched.JsonData)
			assert.Equal(t, uniswapNewEtag, fetched.Etag)
			assert.Equal(t, []byte(uniswapTokenListJsonResponse2), fetched.JsonData)
		case "invalid-url":
			assert.Error(t, fetched.Error)
			assert.Nil(t, fetched.JsonData)
		case "non-existent":
			assert.Error(t, fetched.Error)
			assert.Nil(t, fetched.JsonData)
		}
	}

	assert.True(t, ids["uniswap"])
	assert.True(t, ids["uniswap-with-etag"])
	assert.True(t, ids["uniswap-with-schema"])
	assert.True(t, ids["uniswap-with-wrong-schema"])
	assert.True(t, ids["uniswap-same-etag"])
	assert.True(t, ids["uniswap-new-etag"])
	assert.True(t, ids["invalid-url"])
	assert.True(t, ids["non-existent"])
}

func TestFetchConcurrent_EmptyList(t *testing.T) {
	f := fetcher.New(fetcher.DefaultConfig())

	fetchedLists, err := f.FetchConcurrent(context.Background(), []fetcher.FetchDetails{})

	assert.NoError(t, err)
	assert.Empty(t, fetchedLists)
}

func TestFetchConcurrent_ContextCancellation(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	f := fetcher.New(fetcher.DefaultConfig())

	tokenLists := []fetcher.FetchDetails{
		{
			ListDetails: types.ListDetails{
				ID:        "slow-response",
				SourceURL: server.URL + delayedResponseURL,
				Schema:    "",
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fetchedLists, err := f.FetchConcurrent(ctx, tokenLists)

	assert.NoError(t, err)
	// We might get 0 or 1 results depending on timing, if we get a result, it should have an error
	for _, fetched := range fetchedLists {
		if fetched.Error != nil {
			assert.Contains(t, fetched.Error.Error(), "context")
		}
	}
}
