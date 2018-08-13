package logger

import "testing"

// Tests proper fields allocation.
func TestCorrectFields(t *testing.T) {
	r := withFields("f1", "f1", "f2", "f2")
	if 2 != len(r) {
		t.Fail()
	}

	r = withFields("f1", "f1", "f2", "f2", "f3")
	if 2 != len(r) {
		t.Fail()
	}
}
