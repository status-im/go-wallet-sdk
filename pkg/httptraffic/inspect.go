package httptraffic

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// An Inspector reads what the start of a request body says about the
// request, such as the JSON-RPC methods it calls.
type Inspector interface {
	// Accepts tells whether req's body is worth reading.
	Accepts(req *http.Request) bool
	// Inspect reads the start of the body, decoded when it was gzipped.
	Inspect(prefix []byte) Inspection
}

// Inspection is what an Inspector found in a request body.
type Inspection struct {
	// Label is appended to the endpoint path, e.g. "#eth_call", so that
	// requests to one URL are told apart by what they ask for.
	Label string
	// Calls counts the calls the request carries, e.g. the members of a
	// JSON-RPC batch.
	Calls uint64
	// Bundles counts the calls that bundle further calls, such as eth_calls
	// to Multicall3; BundledCalls and MaxBundledCalls count what they bundle.
	Bundles         uint64
	BundledCalls    uint64
	MaxBundledCalls uint64
}

// MaxInspectedBody bounds how much of a request body an Inspector sees. A
// request is told by its start, so a large body is attributed from its first
// bytes without being copied whole.
const MaxInspectedBody = 64 << 10

// inspect hands the start of req's body, read through GetBody, to the
// inspector, leaving the request itself untouched.
func inspect(in Inspector, req *http.Request) Inspection {
	if in == nil || req.GetBody == nil || req.ContentLength <= 0 || !in.Accepts(req) {
		return Inspection{}
	}
	body, err := req.GetBody()
	if err != nil {
		return Inspection{}
	}
	defer body.Close()
	buf := prefixBuffers.Get().(*[]byte)
	defer prefixBuffers.Put(buf)
	prefix := (*buf)[:min(req.ContentLength, MaxInspectedBody)]
	var r io.Reader = body
	if strings.EqualFold(req.Header.Get("Content-Encoding"), "gzip") {
		zr, err := gzip.NewReader(body)
		if err != nil {
			return Inspection{}
		}
		defer zr.Close()
		r, prefix = zr, (*buf)[:MaxInspectedBody]
	}
	n, _ := io.ReadFull(r, prefix)
	return in.Inspect(prefix[:n])
}

// prefixBuffers recycles the buffers request bodies are inspected in.
var prefixBuffers = sync.Pool{New: func() any {
	b := make([]byte, MaxInspectedBody)
	return &b
}}
