// Package httptraffic counts what HTTP clients put on the wire, so that an
// application's data consumption can be attributed to hosts, endpoints and
// the features that cause it.
//
// The package keeps collecting and analysing apart. The instrumented
// transport records raw facts into a Recorder: per connection, the exact
// bytes it carried; per request, its endpoint (host, method, path, and what
// an Inspector reads in its body, such as a JSON-RPC method), who made it (a
// source tag from the context, or the calling function) and its sizes.
// Snapshot hands those facts to analyze, a pure function that classifies them
// into sources by an Attribution, merges, sums and orders them. Changing how
// traffic is classified therefore never changes what is recorded.
//
// The package knows nothing of the application it measures: which code is
// the application's, how features are named, and which destinations stay private all
// come from its Attribution.
package httptraffic
