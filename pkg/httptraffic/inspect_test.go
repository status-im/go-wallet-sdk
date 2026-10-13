package httptraffic

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// prefixInspector labels every request with the length of the prefix it saw
// and remembers the prefix.
type prefixInspector struct{ seen []byte }

func (*prefixInspector) Accepts(*http.Request) bool { return true }

func (p *prefixInspector) Inspect(prefix []byte) Inspection {
	p.seen = append(p.seen[:0], prefix...)
	return Inspection{Label: "#seen", Calls: 1}
}

func TestInspect_ReadsOnlyThePrefix(t *testing.T) {
	in := &prefixInspector{}
	req, err := http.NewRequest(http.MethodPost, "http://rpc/", strings.NewReader(strings.Repeat("a", 1<<20)))
	require.NoError(t, err)

	require.Equal(t, "#seen", inspect(in, req).Label)
	require.Len(t, in.seen, MaxInspectedBody, "a 1 MB body is inspected by its first 64 KB")
}

func TestInspect_DecodesAGzippedBody(t *testing.T) {
	in := &prefixInspector{}
	req, err := http.NewRequest(http.MethodPost, "http://rpc/", bytes.NewReader(gzipped(t, `{"method":"eth_call"}`)))
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	inspect(in, req)
	require.Equal(t, `{"method":"eth_call"}`, string(in.seen))
}

func TestInspect_LeavesBodiesAloneWithoutAnInspector(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://rpc/", strings.NewReader("x"))
	require.NoError(t, err)
	require.Equal(t, Inspection{}, inspect(nil, req))
}

func TestRecorder_LabelsEndpointsByTheInspection(t *testing.T) {
	server := newServer(t)
	rec := NewRecorder(WithInspector(&prefixInspector{}))
	client := &http.Client{Transport: rec.Instrument(nil)}

	resp, err := client.Post(server.URL+"/rpc", "application/json", strings.NewReader("body")) //nolint:noctx
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	require.Equal(t, uint64(1), endpointFor(t, rec.Snapshot(), "/rpc#seen").Calls)
}
