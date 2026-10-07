package jsonrpc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/status-im/go-wallet-sdk/pkg/httptraffic"
)

func post(t *testing.T, client *http.Client, url, contentType, body string) {
	resp, err := client.Post(url, contentType, strings.NewReader(body)) //nolint:noctx
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func endpointFor(t *testing.T, s httptraffic.Snapshot, path string) httptraffic.EndpointStats {
	for _, e := range s.Endpoints {
		if e.Path == path {
			return e
		}
	}
	t.Fatalf("no stats for endpoint %s", path)
	return httptraffic.EndpointStats{}
}

// multicallData encodes a Multicall3 call with n calls, the way go-ethereum's
// ABI packer lays it out: an offset to the array, then its length.
func multicallData(selector string, arrayArg, n int) string {
	word := func(v int) string { return fmt.Sprintf("%064x", v) }
	head := ""
	for i := 0; i < arrayArg; i++ {
		head += word(1)
	}
	offset := (arrayArg + 1) * 32
	return "0x" + selector + head + word(offset) + word(n) + strings.Repeat("ab", 40*n)
}

func call(method, data string) string {
	return fmt.Sprintf(`{"method":%q,"params":[{"to":"0xcA11bde05977b3631167028862bE2a173976CA11","data":%q},"latest"]}`, method, data)
}

func TestInspector_LabelsRequestsWithTheirMethods(t *testing.T) {
	var received []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = append(received, string(body))
		_, _ = w.Write([]byte(`{"result":"0x1"}`))
	}))
	defer server.Close()
	rec := httptraffic.NewRecorder(httptraffic.WithInspector(Inspector{}))
	client := &http.Client{Transport: rec.Instrument(nil)}

	single := `{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[]}`
	batch := `[{"method":"eth_getBalance"},{"method":"eth_call"},{"method":"eth_call"}]`
	post(t, client, server.URL+"/ethereum/mainnet/", "application/json", single)
	post(t, client, server.URL+"/ethereum/mainnet/", "application/json", batch)
	post(t, client, server.URL+"/ethereum/mainnet/", "text/plain", single)

	s := rec.Snapshot()
	require.Equal(t, uint64(1), endpointFor(t, s, "/ethereum/mainnet/#eth_call").Requests)
	batchStats := endpointFor(t, s, "/ethereum/mainnet/#batch:eth_call,eth_getBalance")
	require.Equal(t, uint64(1), batchStats.Requests)
	require.Equal(t, uint64(3), batchStats.Calls)
	require.Equal(t, uint64(1), endpointFor(t, s, "/ethereum/mainnet/").Requests, "only JSON bodies are inspected")
	require.Equal(t, []string{single, batch, single}, received, "the request has to reach the server untouched")
}

func TestInspector_CountsMulticallCalls(t *testing.T) {
	in := Inspector{}
	require.Equal(t, httptraffic.Inspection{Label: "#eth_call", Calls: 1, Bundles: 1, BundledCalls: 2500, MaxBundledCalls: 2500},
		in.Inspect([]byte(call("eth_call", multicallData("82ad56cb", 0, 2500)))))
	require.Equal(t, uint64(7), in.Inspect([]byte(call("eth_call", multicallData("bce38bd7", 1, 7)))).BundledCalls,
		"tryAggregate keeps its array in the second argument")
	require.Equal(t, httptraffic.Inspection{Label: "#batch:eth_call", Calls: 2, Bundles: 1, BundledCalls: 3, MaxBundledCalls: 3},
		in.Inspect([]byte("["+call("eth_call", multicallData("252dba42", 0, 3))+","+call("eth_call", "0x70a08231")+"]")))
	require.Equal(t, httptraffic.Inspection{Label: "#eth_call", Calls: 1}, in.Inspect([]byte(call("eth_call", "0x70a08231"))),
		"a plain eth_call is not a multicall")
}

func TestInspector_ReadsATruncatedCall(t *testing.T) {
	body := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xca11","data":%q},"latest"]}`,
		multicallData("82ad56cb", 0, 2500))

	in := Inspector{}.Inspect([]byte(body[:300]))
	require.Equal(t, "#eth_call", in.Label)
	require.Equal(t, uint64(2500), in.MaxBundledCalls, "the array length sits at the start of the calldata")
}

func TestInspector_IgnoresWhatIsNotJSONRPC(t *testing.T) {
	for _, body := range []string{"", "plain", `{"jsonrpc":"2.0"}`, `{"method":42}`, `["x"]`} {
		require.Equal(t, httptraffic.Inspection{}, Inspector{}.Inspect([]byte(body)), body)
	}
}

func TestInspector_AcceptsOnlyJSONPosts(t *testing.T) {
	for _, c := range []struct {
		method, contentType string
		want                bool
	}{
		{http.MethodPost, "application/json", true},
		{http.MethodPost, "application/json; charset=utf-8", true},
		{http.MethodPost, "text/plain", false},
		{http.MethodGet, "application/json", false},
	} {
		req, err := http.NewRequest(c.method, "http://rpc/", nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", c.contentType)
		require.Equal(t, c.want, Inspector{}.Accepts(req), c.method+" "+c.contentType)
	}
}

func BenchmarkInspector_Multicall(b *testing.B) {
	body := []byte(call("eth_call", multicallData("82ad56cb", 0, 2500)))
	prefix := body[:min(len(body), httptraffic.MaxInspectedBody)]
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Inspector{}.Inspect(prefix)
	}
}
