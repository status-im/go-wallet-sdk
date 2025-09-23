package types

import (
	"fmt"
)

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

func (r *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", r.Major, r.Minor, r.Patch)
}

// TokenList represents a token list.
type TokenList struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Timestamp        string                 `json:"timestamp"`        // time when the list was last updated
	FetchedTimestamp string                 `json:"fetchedTimestamp"` // time when the list was fetched
	Source           string                 `json:"source"`
	Version          Version                `json:"version"`
	Tags             map[string]interface{} `json:"tags"`
	LogoURI          string                 `json:"logoUri"`
	Keywords         []string               `json:"keywords"`
	Tokens           []*Token               `json:"tokens"`
}
