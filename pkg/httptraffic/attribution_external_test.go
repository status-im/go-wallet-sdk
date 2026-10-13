package httptraffic_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/httptraffic"
)

type marker struct{}

// thisPackage prefixes the functions of this test package, which stands for
// the application here.
var thisPackage = reflect.TypeOf(marker{}).PkgPath() + "."

func fetchPrices(t *testing.T, client *http.Client, url string) {
	resp, err := client.Get(url) //nolint:noctx
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	require.NoError(t, resp.Body.Close())
}

func TestRecorder_AttributesUntaggedRequestsToTheFunctionThatMadeThem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()
	rec := httptraffic.NewRecorder(httptraffic.WithAttribution(httptraffic.Attribution{
		Modules:  []string{thisPackage},
		Plumbing: []string{thisPackage + "TestRecorder_"},
		Sources:  []httptraffic.SourceRule{{Function: ".fetchPrices", Source: "Prices"}},
	}))
	client := &http.Client{Transport: rec.Instrument(nil)}

	fetchPrices(t, client, server.URL+"/v1/simple/price")

	s := rec.Snapshot()
	require.Len(t, s.Endpoints, 1)
	require.Equal(t, "Prices", s.Endpoints[0].Source)
	require.Equal(t, "httptraffic_test.fetchPrices", s.Endpoints[0].Caller)
}
