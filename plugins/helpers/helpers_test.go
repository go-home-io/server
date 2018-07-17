package helpers

import (
	"testing"
	"github.com/go-home-io/server/plugins/device/enums"
)

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

func TestDeepEquals(t *testing.T) {
	x := map[string]interface{}{"r": 1, "g": 2, "b": 3}
	y := map[string]interface{}{"r": 1, "g": 2, "b": 3}

	if !PropertyDeepEqual(x, y, enums.PropColor) {
		t.Fail()
	}
}
