package httptraffic

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// stubTransport answers every request at once, so that a benchmark measures
// the recording around it.
type stubTransport struct{}

func (stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		_, _ = io.Copy(io.Discard, req.Body)
		_ = req.Body.Close()
	}
	return &http.Response{StatusCode: 200, ProtoMajor: 2, Header: http.Header{},
		Body: io.NopCloser(bytes.NewReader([]byte(`{"result":"0x"}`)))}, nil
}

// multicallBody is a 960 KB JSON-RPC call, the size of a Multicall3 eth_call
// bundling 2500 balance reads.
var multicallBody = []byte(`{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0xca11","data":"0x82ad56cb` +
	strings.Repeat("ab", 480<<10) + `"},"latest"]}`)

func benchmarkRoundTrip(b *testing.B, rt http.RoundTripper, ctx context.Context) {
	b.ReportAllocs()
	b.SetBytes(int64(len(multicallBody)))
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://rpc.example/ethereum/mainnet/", bytes.NewReader(multicallBody))
		req.Header.Set("Content-Type", "application/json")
		resp, err := rt.RoundTrip(req)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

func BenchmarkRoundTrip_Bare(b *testing.B) {
	benchmarkRoundTrip(b, stubTransport{}, context.Background())
}

func BenchmarkRoundTrip_Disabled(b *testing.B) {
	rec := NewRecorder()
	rec.SetEnabled(false)
	benchmarkRoundTrip(b, rec.Instrument(stubTransport{}), context.Background())
}

func BenchmarkRoundTrip_Tagged(b *testing.B) {
	benchmarkRoundTrip(b, NewRecorder().Instrument(stubTransport{}), WithSource(context.Background(), "Balances"))
}

func BenchmarkRoundTrip_StackWalk(b *testing.B) {
	rec := NewRecorder(WithAttribution(Attribution{Modules: []string{"github.com/status-im/"}}))
	benchmarkRoundTrip(b, rec.Instrument(stubTransport{}), context.Background())
}

func BenchmarkAddConnBytes(b *testing.B) {
	rec := NewRecorder()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rec.addConnBytes("rpc.example", 1500, 0)
		}
	})
}
