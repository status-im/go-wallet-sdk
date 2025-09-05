package tokenlists

import (
	"context"
	"encoding/json"
	"time"

	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

// fetchRemoteListOfTokenLists fetches the remote list of token lists from the URL specified in the config.
func (t *tokenListsFetcher) fetchRemoteListOfTokenLists(ctx context.Context, etag string) ([]byte, string, error) {
	body, newEtag, err := t.httpClient.DoGetRequestWithEtag(ctx, t.config.RemoteListOfTokenListsURL, etag)
	if err != nil {
		return nil, "", err
	}

	if etag != "" && newEtag == etag {
		return nil, "", nil
	}

	err = validateJsonAgainstSchema(string(body), gojsonschema.NewStringLoader(listOfTokenListsSchema))
	if err != nil {
		return nil, "", err
	}

	return body, newEtag, nil
}

// resolveListOfTokenLists resolves the list of token lists from the remote list of token lists.
func (t *tokenListsFetcher) resolveListOfTokenLists(ctx context.Context) (remoteListOfTokenLists, error) {
	var remoteListOfTokenLists remoteListOfTokenLists
	var ok = true
	storedContent, err := t.config.ContentStore.Get(StatusListOfTokenListsID)
	if err != nil {
		ok = false
		t.config.logger.Error("failed to get stored content", zap.Error(err))
	}

	if t.config.RemoteListOfTokenListsURL != "" {
		fetchedData, newEtag, err := t.fetchRemoteListOfTokenLists(ctx, storedContent.Etag)
		if err != nil {
			// don't return, but instead try to use the last cached list of token lists
			t.config.logger.Error("failed to fetch remote list of token lists", zap.Error(err))
			goto useStoredList
		}

		if fetchedData != nil {
			if err = json.Unmarshal(fetchedData, &remoteListOfTokenLists); err != nil {
				t.config.logger.Error("failed to unmarshal remote list of token lists", zap.Error(err))
				goto useStoredList
			}

			t.config.logger.Info("new remote list of token lists fetched successfully",
				zap.String("version", remoteListOfTokenLists.Version.String()),
				zap.String("timestamp", remoteListOfTokenLists.Timestamp),
			)

			err := t.config.ContentStore.Set(StatusListOfTokenListsID, Content{
				SourceURL: t.config.RemoteListOfTokenListsURL,
				Etag:      newEtag,
				Data:      fetchedData,
				Fetched:   time.Now(),
			})
			if err != nil {
				t.config.logger.Error("failed to store remote list of token lists", zap.Error(err))
			}

			return remoteListOfTokenLists, nil
		}
	}

useStoredList:
	if ok && len(storedContent.Data) > 0 {
		err := json.Unmarshal(storedContent.Data, &remoteListOfTokenLists)
		if err != nil {
			t.config.logger.Error("failed to unmarshal stored list of token lists", zap.Error(err))
		}
	}

	return remoteListOfTokenLists, nil
}
