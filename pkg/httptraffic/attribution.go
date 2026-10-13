package httptraffic

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

// Attribution tells a recorder how to name the source of a request: which
// function on the stack made it, and which feature that function belongs to.
// A request tagged with WithSource needs none of it.
//
// Without an attribution only tags name sources; everything else is
// Unattributed.
type Attribution struct {
	// Modules are the import path prefixes of the application's own code,
	// e.g. "github.com/example/app/". The function that made an untagged
	// request is the frame of these modules nearest the goroutine's root.
	// Without modules the stack is not walked.
	Modules []string
	// Plumbing are function name prefixes, with import path, that carry
	// requests for others: clients, transports, retries. A caller is never one
	// of them but the function behind them.
	Plumbing []string
	// EntryPoints matches functions that only receive calls from outside the
	// application, such as RPC API methods: like plumbing, they are skipped.
	EntryPoints *regexp.Regexp
	// Sources name the source of a request; the first rule that matches wins.
	// When none does, the source is the caller's package.
	Sources []SourceRule
	// PrivateCallers are function name prefixes whose requests go to addresses
	// users gave, such as links to preview: neither their hosts nor their paths
	// are recorded. A request tagged with WithSource has no caller; mark it
	// with WithPrivateDestination instead.
	PrivateCallers []string
	// Unattributed names the requests nothing claims. "Other" when empty.
	Unattributed string
}

// SourceRule names Source as the source of the requests whose path contains
// Path, or whose caller's name, with import path, contains Function. A rule
// sets one of the two.
type SourceRule struct {
	Path     string
	Function string
	Source   string
}

const maxStackDepth = 128

// selfPackage prefixes the functions of this package, which are always plumbing.
var selfPackage = reflect.TypeOf(Recorder{}).PkgPath() + "."

func (a *Attribution) unattributed() string {
	if a == nil || a.Unattributed == "" {
		return "Other"
	}
	return a.Unattributed
}

// callerOf returns the application function that started the current
// request, with its import path: the frame nearest the goroutine's root, past
// the plumbing. It is empty when no such frame is on the stack.
func (a *Attribution) callerOf() string {
	if a == nil || len(a.Modules) == 0 {
		return ""
	}
	var pcs [maxStackDepth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	caller := ""
	for {
		frame, more := frames.Next()
		if a.IsFeature(frame.Function) {
			caller = frame.Function
		}
		if !more {
			break
		}
	}
	return caller
}

// IsFeature tells whether function, given with its import path, can be the
// caller of a request: application code that is neither plumbing nor an entry
// point.
func (a *Attribution) IsFeature(function string) bool {
	if a == nil || strings.HasPrefix(function, selfPackage) || !hasAnyPrefix(function, a.Modules) || hasAnyPrefix(function, a.Plumbing) {
		return false
	}
	return a.EntryPoints == nil || !a.EntryPoints.MatchString(function)
}

// IsPrivate tells whether the requests of function keep no path.
func (a *Attribution) IsPrivate(function string) bool {
	return a != nil && function != "" && hasAnyPrefix(function, a.PrivateCallers)
}

// SourceOf names the source of an untagged request from its caller, given with
// its import path and empty when unknown, and the request path.
func (a *Attribution) SourceOf(function, path string) string {
	if a != nil {
		for _, rule := range a.Sources {
			if rule.Path != "" && strings.Contains(path, rule.Path) {
				return rule.Source
			}
			if rule.Function != "" && function != "" && strings.Contains(function, rule.Function) {
				return rule.Source
			}
		}
	}
	if function == "" {
		return a.unattributed()
	}
	pkg := shortFunction(function)
	if i := strings.Index(pkg, "."); i >= 0 {
		pkg = pkg[:i]
	}
	return pkg
}

var closureSuffix = regexp.MustCompile(`(\.(func|gowrap)\d+)+$`)

// shortFunction drops the import path and closure suffixes:
// ".../balance.(*Controller).fetchChain.func1" becomes
// "balance.(*Controller).fetchChain".
func shortFunction(function string) string {
	if i := strings.LastIndex(function, "/"); i >= 0 {
		function = function[i+1:]
	}
	return closureSuffix.ReplaceAllString(function, "")
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
