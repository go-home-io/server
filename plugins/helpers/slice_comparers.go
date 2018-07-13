// Package helpers contains various helpers which can be re-used by plugin.
package helpers

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
