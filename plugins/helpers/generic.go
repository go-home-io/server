package helpers

import (
	"strings"

	"github.com/sanity-io/litter"
)

// GetNameFromID converts entity ID to readable name.
// Used if name is not overwritten.
func GetNameFromID(entityID string) string {
	parts := strings.Split(entityID, ".")
	for ii := len(parts) - 1; ii > 0; ii-- {
		r := strings.Replace(parts[ii], "_", " ", -1)
		if "" == r {
			continue
		}

		return r
	}

	return "N/A"
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

// DebugObjectStringify is a debug function which returns prettified string representation
// of an object.
func DebugObjectStringify(i interface{}) string {
	litter.Config.HidePrivateFields = false
	return litter.Sdump(i)
}
