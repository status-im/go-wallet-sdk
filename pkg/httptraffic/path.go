package httptraffic

import "strings"

// Paths end up in snapshots, logs and the clipboard, so a segment that
// identifies something — a provider token, an address, a hash, an id — is
// recorded as idSegment. Provider tokens in particular sit in the path of
// token-auth RPC URLs.
const idSegment = "{id}"

// userPath is the whole path of a request to an address a user gave, such as
// a link to preview: where people go is not ours to log.
const userPath = "/…"

func (a *Attribution) recordedPath(caller, path string) string {
	if a.IsPrivate(callerFunction(caller)) {
		return userPath
	}
	if !strings.ContainsFunc(path, isDigitOrUpper) {
		return path
	}
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if identifies(segment) {
			segments[i] = idSegment
		}
	}
	return strings.Join(segments, "/")
}

// identifies tells a segment that names one thing from one that names a kind
// of thing: hex with 0x, a long number, or a long string with digits or
// mixed case, which is what tokens, UUIDs and hashes look like and words do not.
func identifies(segment string) bool {
	var digits, upper, lower int
	for _, c := range segment {
		switch {
		case c >= '0' && c <= '9':
			digits++
		case c >= 'A' && c <= 'Z':
			upper++
		case c >= 'a' && c <= 'z':
			lower++
		}
	}
	switch {
	case len(segment) >= 4 && (strings.HasPrefix(segment, "0x") || strings.HasPrefix(segment, "0X")):
		return true
	case len(segment) >= 6 && digits == len(segment):
		return true
	case len(segment) >= 16 && digits > 0:
		return true
	case len(segment) >= 20 && upper > 0 && lower > 0:
		return true
	}
	return false
}

func isDigitOrUpper(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')
}
