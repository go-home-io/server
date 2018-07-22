package helpers

import (
	"testing"
	"github.com/go-home-io/server/plugins/device/enums"
)

// Tests ID -> name conversion.
func TestGetNameFromID(t *testing.T) {
	in := []string{"hue.light.w_r1_fan_3", "test._device", "другой.девайс_в_кухне"}
	out := []string{"w r1 fan 3", " device", "девайс в кухне"}
	for i, v := range in {
		o := GetNameFromID(v)
		if o != out[i] {
			t.Fail()
		}
	}
}

// Tests contains helper.
func TestSliceContainsString(t *testing.T) {
	in := []string{"$", "@", "testFunc", "#!_=", "123", "другая строка"}

	for _, v := range in {
		if !SliceContainsString(in, v) {
			t.Fail()
		}

		if SliceContainsString(in, v+v) {
			t.Fail()
		}
	}
}

// Tests slices equality.
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

// Tests that slices are not equal.
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
