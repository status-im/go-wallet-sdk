# HTTP Traffic Package

The `httptraffic` package counts what HTTP clients put on the wire and attributes it to hosts, endpoints and the features of an application that cause it.

## Use it when

- You need to know which feature of an app costs how much data.
- You want per-endpoint request counts, status codes, latencies and sizes, including JSON-RPC methods and Multicall3 bundles.
- You want to tell traffic sent in the background apart.

## Key entrypoints

- `httptraffic.NewRecorder(opts...)` with `WithAttribution`, `WithInspector`, `WithClock`
- `(*Recorder).Instrument(rt)` and `(*Recorder).WrapDialContext(dial)`
- `httptraffic.WithSource(ctx, name)` to tag requests with their feature
- `(*Recorder).Snapshot()`, `Snapshot.Filter(query)`, `Snapshot.Sub(prev)`
- `(*Recorder).Report(ctx, interval, samplesPerReport, report)` for periodic samples
- `jsonrpc.Inspector` to label JSON-RPC requests by method

## Features

- **Wire bytes**: per connection, and exact per request over HTTP/1.1
- **Attribution**: a context tag, or the first application frame on the stack, named by configurable rules
- **Private paths**: segments that look like keys, addresses or ids are recorded as `{id}`; chosen callers keep no path
- **Bounded memory**: host and endpoint tables are capped
- **No globals**: every recorder is configured by its caller
- **Cheap when off**: a disabled recorder passes requests straight through

## Example

```go
rec := httptraffic.NewRecorder(
    httptraffic.WithAttribution(httptraffic.Attribution{
        Modules:  []string{"github.com/example/app/"},
        Plumbing: []string{"github.com/example/app/internal/rpc"},
        Sources:  []httptraffic.SourceRule{{Function: "/balance", Source: "Balances"}},
    }),
    httptraffic.WithInspector(jsonrpc.Inspector{}),
)
client := &http.Client{Transport: rec.Instrument(nil)}

ctx := httptraffic.WithSource(context.Background(), "Balances")
// requests made with ctx count towards "Balances"

snapshot := rec.Snapshot()
fmt.Println(snapshot.Insights.TopSource, snapshot.Totals.BytesSent)
```
