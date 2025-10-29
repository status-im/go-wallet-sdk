package fetcher_test

import (
	"context"
	"strings"
	"testing"

	"github.com/status-im/go-wallet-sdk/pkg/tokens/fetcher"
	"github.com/status-im/go-wallet-sdk/pkg/tokens/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchRemoteListOfTokenLists(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	f := fetcher.New(fetcher.DefaultConfig())

	tests := []struct {
		name          string
		url           string
		etag          string
		expectedError bool
		expectedEtag  string
		expectedData  []byte
	}{
		{
			name:          "successful fetch without etag",
			url:           server.URL + listOfTokenListsURL,
			etag:          "",
			expectedError: false,
			expectedEtag:  "",
			expectedData:  []byte(strings.ReplaceAll(listOfTokenListsJsonResponse, serverURLPlaceholder, server.URL)),
		},
		{
			name:          "successful fetch with etag",
			url:           server.URL + listOfTokenListsWithEtagURL,
			etag:          "",
			expectedError: false,
			expectedEtag:  listOfTokenListsEtag,
			expectedData:  []byte(strings.ReplaceAll(listOfTokenListsJsonResponse1, serverURLPlaceholder, server.URL)),
		},
		{
			name:          "fetch with same etag",
			url:           server.URL + listOfTokenListsWithSameEtagURL,
			etag:          listOfTokenListsEtag,
			expectedError: false,
			expectedEtag:  listOfTokenListsEtag,
			expectedData:  nil,
		},
		{
			name:          "fetch with new etag returns new data",
			url:           server.URL + listOfTokenListsWithNewEtagURL,
			etag:          listOfTokenListsEtag,
			expectedError: false,
			expectedEtag:  listOfTokenListsNewEtag,
			expectedData:  []byte(strings.ReplaceAll(listOfTokenListsJsonResponse2, serverURLPlaceholder, server.URL)),
		},
		{
			name:          "fetch from non-existent URL returns error",
			url:           server.URL + "/non-existent.json",
			etag:          "",
			expectedError: true,
			expectedEtag:  "",
			expectedData:  nil,
		},
		{
			name:          "fetch returns valid data with wrong URLs",
			url:           server.URL + listOfTokenListsSomeWrongUrlsURL,
			etag:          "",
			expectedError: false, // This is valid JSON that passes schema validation
			expectedEtag:  "",
			expectedData:  []byte(strings.ReplaceAll(listOfTokenListsWrongUrlsJsonResponse, serverURLPlaceholder, server.URL)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := f.Fetch(context.Background(), fetcher.FetchDetails{
				ListDetails: types.ListDetails{
					ID:        "test-list",
					SourceURL: tt.url,
					Schema:    "",
				},
				Etag: tt.etag,
			})

			if tt.expectedError {
				assert.Error(t, err)
				assert.Empty(t, data.JsonData)
				assert.Empty(t, data.Etag)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedEtag, data.Etag)

			if tt.expectedData != nil {
				assert.Equal(t, tt.expectedData, data.JsonData)
			} else {
				assert.Empty(t, data.JsonData)
			}
		})
	}
}

func TestFetchRemoteListOfTokenLists_ContextCancellation(t *testing.T) {
	server, closeServer := GetTestServer()
	t.Cleanup(func() {
		closeServer()
	})

	f := fetcher.New(fetcher.DefaultConfig())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	data, err := f.Fetch(ctx, fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			ID:        "test-list",
			SourceURL: server.URL + delayedResponseURL,
			Schema:    "",
		},
		Etag: "",
	})

	assert.Error(t, err)
	assert.Empty(t, data.JsonData)
	assert.Empty(t, data.Etag)
}

func TestFetchRemoteListOfTokenLists_InvalidURL(t *testing.T) {
	f := fetcher.New(fetcher.DefaultConfig())

	data, err := f.Fetch(context.Background(), fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			ID:        "test-list",
			SourceURL: "invalid-url",
			Schema:    "",
		},
		Etag: "",
	})

	assert.Error(t, err)
	assert.Empty(t, data.JsonData)
	assert.Empty(t, data.Etag)
}

func TestFetchRemoteListOfTokenLists_EmptyURL(t *testing.T) {
	f := fetcher.New(fetcher.DefaultConfig())

	data, err := f.Fetch(context.Background(), fetcher.FetchDetails{
		ListDetails: types.ListDetails{
			ID:        "test-list",
			SourceURL: "",
			Schema:    "",
		},
		Etag: "",
	})

	assert.Error(t, err)
	assert.Empty(t, data.JsonData)
	assert.Empty(t, data.Etag)
}
