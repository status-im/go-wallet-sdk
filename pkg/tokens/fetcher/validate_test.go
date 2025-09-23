package fetcher

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateJsonAgainstStringProvidedSchema(t *testing.T) {
	validJSON := `{"name": "Test Token List", "tokens": []}`
	invalidJSON := `{"name": "Test Token List"}`

	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"tokens": {"type": "array"}
		},
		"required": ["name", "tokens"]
	}`

	err := validateJsonAgainstSchema(validJSON, schema)
	assert.NoError(t, err)

	err = validateJsonAgainstSchema(invalidJSON, schema)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenListDoesNotMatchSchema)
}

func TestValidateJsonAgainstURLProvidedSchema(t *testing.T) {
	const (
		serverURLPlaceholder      = "SERVER-URL"
		listOfTokenListsSchemaURL = "/list-of-token-lists-schema.json" // #nosec G101
	)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.HandleFunc(listOfTokenListsSchemaURL, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(ListOfTokenListsSchema)); err != nil {
			log.Println(err.Error())
		}
	})

	validJSON := `{
		"timestamp": "2025-09-01T00:00:00.000Z",
		"version": {
			"major": 0,
			"minor": 1,
			"patch": 0
		},
		"tokenLists": [
			{
				"id": "status",
				"sourceUrl": "SERVER-URL/status-token-list.json"
			},
			{
				"id": "uniswap",
				"sourceUrl": "SERVER-URL/uniswap.json"
			}
		]
	}`
	invalidJSON := `{"tokenLists": []}`

	schema := server.URL + listOfTokenListsSchemaURL

	resp := strings.ReplaceAll(validJSON, serverURLPlaceholder, server.URL)

	err := validateJsonAgainstSchema(resp, schema)
	assert.NoError(t, err)

	err = validateJsonAgainstSchema(invalidJSON, schema)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrTokenListDoesNotMatchSchema)
}
