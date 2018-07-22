package helpers

import (
	"strings"
)

// GetNameFromID converts entity ID to readable name.
// Used if name is not overwritten.
func GetNameFromID(entityID string) string {
	parts := strings.Split(entityID, ".")
	return strings.Replace(parts[len(parts)-1], "_", " ", -1)
}

// SliceEqualsString checks whether string slices are equal.
// Order is ignored.
func SliceEqualsString(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	diff := make(map[string]int, len(x))
	for _, ix := range x {
		diff[ix]++
	}
	for _, iy := range y {
		if _, ok := diff[iy]; !ok {
			return false
		}
		diff[iy]--
		if diff[iy] == 0 {
			delete(diff, iy)
		}
	}

	return len(diff) == 0
}

// SliceContainsString slice.contains implementation for strings.
func SliceContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
