package httptraffic

import "context"

type sourceKey struct{}

// WithSource tags the requests made with ctx as coming from source. It is
// precise where the call stack is not, and saves walking it.
func WithSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, sourceKey{}, source)
}

func sourceFrom(ctx context.Context) (string, bool) {
	source, ok := ctx.Value(sourceKey{}).(string)
	return source, ok && source != ""
}

type privateKey struct{}

// WithPrivateDestination marks the requests made with ctx as going to an
// address a user gave: neither their host nor their path is recorded.
func WithPrivateDestination(ctx context.Context) context.Context {
	return context.WithValue(ctx, privateKey{}, true)
}

func isPrivateDestination(ctx context.Context) bool {
	private, _ := ctx.Value(privateKey{}).(bool)
	return private
}

// tagPrefix marks a raw caller that is a source tag rather than a function.
const tagPrefix = "@"

// callerKey names who makes a request: its source tag, or else the function
// that started it.
func (a *Attribution) callerKey(ctx context.Context) string {
	if source, ok := sourceFrom(ctx); ok {
		return tagPrefix + source
	}
	return a.callerOf()
}

// callerFunction is the function of a raw caller, empty for a source tag.
func callerFunction(caller string) string {
	if len(caller) > 0 && caller[:1] == tagPrefix {
		return ""
	}
	return caller
}

// classify names the source of a raw caller and path: a tag names it outright.
func (a *Attribution) classify(caller, path string) string {
	if len(caller) > 0 && caller[:1] == tagPrefix {
		return caller[1:]
	}
	return a.SourceOf(caller, path)
}
