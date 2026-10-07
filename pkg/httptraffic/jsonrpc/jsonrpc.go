// Package jsonrpc tells Ethereum JSON-RPC requests apart for httptraffic: by
// the methods they call, and for eth_calls to Multicall3 by how many calls
// they bundle.
package jsonrpc

import (
	"bytes"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/status-im/go-wallet-sdk/pkg/httptraffic"
)

// Inspector labels a JSON-RPC request "#eth_call" for a single call and
// "#batch:eth_call,eth_getBalance" for a batch. It scans for the keys rather
// than decoding the JSON, so a 960 KB eth_call costs microseconds; a batch
// longer than httptraffic.MaxInspectedBody counts the calls it starts with.
type Inspector struct{}

// Accepts takes JSON POSTs, which is what JSON-RPC over HTTP sends.
func (Inspector) Accepts(req *http.Request) bool {
	return req.Method == http.MethodPost && strings.Contains(req.Header.Get("Content-Type"), "json")
}

// Inspect reads the methods and Multicall3 sizes of the calls prefix starts with.
func (Inspector) Inspect(prefix []byte) httptraffic.Inspection {
	return parse(prefix)
}

// callWindow is how far after its method a call's data is looked for.
const callWindow = 1 << 10

// multicallPrefix is how much calldata multicallSize needs: the selector and
// the first four ABI words.
const multicallPrefix = 2 + 8 + 4*64

var (
	methodKey = []byte(`"method"`)
	dataKeys  = [][]byte{[]byte(`"input"`), []byte(`"data"`)}
)

// parse finds the methods and Multicall3 sizes of the calls a JSON-RPC body
// starts with.
func parse(body []byte) httptraffic.Inspection {
	body = bytes.TrimSpace(body)
	if len(body) == 0 || (body[0] != '{' && body[0] != '[') {
		return httptraffic.Inspection{}
	}
	batch := body[0] == '['

	var in httptraffic.Inspection
	unique := make(map[string]struct{})
	for rest := body; ; {
		value, after, ok := stringValue(rest, methodKey)
		if !ok {
			break
		}
		method := string(value)
		in.Calls++
		unique[method] = struct{}{}
		// The call's data follows its method closely, before the next call's; a
		// single call has no next one to look for.
		callEnd := min(len(after), callWindow)
		if batch {
			if next := bytes.Index(after[:callEnd], methodKey); next >= 0 {
				callEnd = next
			}
		}
		if method == "eth_call" {
			if n, ok := multicallSize(callData(after[:callEnd])); ok {
				in.Bundles++
				in.BundledCalls += n
				in.MaxBundledCalls = max(in.MaxBundledCalls, n)
			}
		}
		if !batch {
			break
		}
		rest = after
	}
	if in.Calls == 0 {
		return httptraffic.Inspection{}
	}

	methods := make([]string, 0, len(unique))
	for m := range unique {
		methods = append(methods, m)
	}
	sort.Strings(methods)
	if batch {
		in.Label = "#batch:" + strings.Join(methods, ",")
	} else {
		in.Label = "#" + methods[0]
	}
	return in
}

// callData returns the start of the hex of a call's transaction data, "input"
// preferred: as much as multicallSize reads.
func callData(call []byte) string {
	for _, key := range dataKeys {
		if v, _, ok := stringValue(call, key); ok {
			return string(v[:min(len(v), multicallPrefix)])
		}
	}
	return ""
}

// stringValue finds key in data and returns the JSON string that follows it,
// without copying it, and the bytes after that string. Escapes are not
// decoded: the values looked for are method names and hex.
func stringValue(data, key []byte) (value, after []byte, ok bool) {
	for {
		i := bytes.Index(data, key)
		if i < 0 {
			return nil, nil, false
		}
		data = data[i+len(key):]
		j := skipSpace(data, 0)
		if j >= len(data) || data[j] != ':' {
			continue
		}
		j = skipSpace(data, j+1)
		if j >= len(data) || data[j] != '"' {
			continue
		}
		end := bytes.IndexByte(data[j+1:], '"')
		if end < 0 {
			// Cut off at the end of the inspected prefix: its start still counts.
			return data[j+1:], nil, true
		}
		return data[j+1 : j+1+end], data[j+2+end:], true
	}
}

func skipSpace(data []byte, i int) int {
	for i < len(data) && (data[i] == ' ' || data[i] == '\t' || data[i] == '\n' || data[i] == '\r') {
		i++
	}
	return i
}

// multicallArrayArg is the argument index of the Call[] array for each
// Multicall3 function, by selector.
var multicallArrayArg = map[string]int{
	"252dba42": 0, // aggregate((address,bytes)[])
	"82ad56cb": 0, // aggregate3((address,bool,bytes)[])
	"174dea71": 0, // aggregate3Value((address,bool,uint256,bytes)[])
	"c3077fa9": 0, // blockAndAggregate((address,bytes)[])
	"bce38bd7": 1, // tryAggregate(bool,(address,bytes)[])
	"399542e9": 1, // tryBlockAndAggregate(bool,(address,bytes)[])
}

// multicallSize returns how many calls the hex calldata of an eth_call to
// Multicall3 bundles.
func multicallSize(data string) (uint64, bool) {
	data = strings.TrimPrefix(data, "0x")
	if len(data) < 8 {
		return 0, false
	}
	arg, ok := multicallArrayArg[data[:8]]
	if !ok {
		return 0, false
	}
	args := data[8:]
	offset, ok := abiWord(args, arg*64)
	if !ok || offset > uint64(len(args)) {
		return 0, false
	}
	return abiWord(args, int(offset)*2)
}

// abiWord reads the 32-byte ABI word that starts at the given hex offset.
func abiWord(hexArgs string, at int) (uint64, bool) {
	if at < 0 || at+64 > len(hexArgs) {
		return 0, false
	}
	word := hexArgs[at : at+64]
	if strings.TrimLeft(word[:48], "0") != "" {
		return 0, false
	}
	v, err := strconv.ParseUint(word[48:], 16, 64)
	return v, err == nil
}
