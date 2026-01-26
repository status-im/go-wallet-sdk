package manager

import (
	"fmt"
)

// StaticInitialListProvider returns an InitialListProvider from a map of initial lists.
func StaticInitialListProvider(initialLists map[string][]byte) InitialListProvider {
	return func(id string) ([]byte, error) {
		b, ok := initialLists[id]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrNotFoundInInitialLists, id)
		}
		return b, nil
	}
}

// InitialListIDsFromMap returns a list of initial list IDs from a map of initial lists.
func InitialListIDsFromMap(initialLists map[string][]byte) []string {
	ids := make([]string, 0, len(initialLists))
	for id := range initialLists {
		ids = append(ids, id)
	}
	return ids
}
