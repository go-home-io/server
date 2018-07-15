package helpers

import "testing"

func TestEqualSlices(t *testing.T) {
	if !SliceEqualsString(
		[]string{"1", "2", "3"},
		[]string{"1", "2", "3"},
	) {
		t.Fail()
	}

	if !SliceEqualsString(
		[]string{"1", "2", "3"},
		[]string{"3", "2", "1"},
	) {
		t.Fail()
	}


	if !SliceEqualsString(
		[]string{"1", "2", "1"},
		[]string{"1", "1", "2"},
	) {
		t.Fail()
	}
}


func TestUnEqualSlices(t *testing.T) {
	if SliceEqualsString(
		[]string{"1", "2", "3"},
		[]string{"1", "2"},
	) {
		t.Fail()
	}

	if SliceEqualsString(
		[]string{"2", "2", "3"},
		[]string{"3", "2", "1"},
	) {
		t.Fail()
	}


	if SliceEqualsString(
		[]string{"2", "2", "1"},
		[]string{"1", "1", "2"},
	) {
		t.Fail()
	}
}
