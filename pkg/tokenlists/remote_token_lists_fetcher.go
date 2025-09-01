package tokenlists

import (
	"context"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

// fetchTokenList fetches a token list from the URL specified in the list.
func (t *tokenListsFetcher) fetchTokenList(ctx context.Context, list tokenList, etag string, ch chan<- fetchedTokenList) error {
	body, newEtag, err := t.httpClient.DoGetRequestWithEtag(ctx, list.SourceURL, etag)
	if err != nil {
		return err
	}

	if etag != "" && newEtag == etag {
		return nil
	}

	if list.Schema != "" {
		err = validateJsonAgainstSchema(string(body), gojsonschema.NewReferenceLoader(list.Schema))
		if err != nil {
			return err
		}
	}

	// check if the channel is closed
	if ch == nil {
		return ErrChannelClosed
	}

	ch <- fetchedTokenList{
		tokenList: tokenList{
			ID:        list.ID,
			SourceURL: list.SourceURL,
			Schema:    list.Schema,
		},
		Etag:     newEtag,
		Fetched:  time.Now(),
		JsonData: body,
	}

	return nil
}
