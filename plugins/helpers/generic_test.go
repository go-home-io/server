package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/plugins/device/enums"
)

// Tests ID -> name conversion.
func TestGetNameFromID(t *testing.T) {
	data := []struct {
		in  string
		out string
	}{
		{
			in:  "hue.light.w_r1_fan_3",
			out: "w r1 fan 3",
		},
		{
			in:  "test._device",
			out: " device",
		},
		{
			in:  "другой.девайс_в_кухне",
			out: "девайс в кухне",
		},
	}

	for _, v := range data {
		assert.Equal(t, v.out, GetNameFromID(v.in), v.in)
	}
}

// Tests contains helper.
func TestSliceContainsString(t *testing.T) {
	in := []string{"$", "@", "testFunc", "#!_=", "123", "другая строка"}

	for _, v := range in {
		assert.True(t, SliceContainsString(in, v), "equal %s", v)
		assert.False(t, SliceContainsString(in, v+v), "not equal %s", v)
	}
}

// Tests slices equality.
func TestEqualSlices(t *testing.T) {
	data := []struct {
		s1 []string
		s2 []string
	}{
		{
			s1: []string{"1", "2", "3"},
			s2: []string{"1", "2", "3"},
		},
		{
			s1: []string{"1", "2", "3"},
			s2: []string{"3", "2", "1"},
		},
		{
			s1: []string{"1", "2", "1"},
			s2: []string{"1", "1", "2"},
		},
	}

	for i, v := range data {
		assert.True(t, SliceEqualsString(v.s1, v.s2), "%d", i)
	}
}

// Tests that slices are not equal.
func TestUnEqualSlices(t *testing.T) {
	data := []struct {
		s1 []string
		s2 []string
	}{
		{
			s1: []string{"1", "2", "3"},
			s2: []string{"1", "2"},
		},
		{
			s1: []string{"2", "2", "3"},
			s2: []string{"3", "2", "1"},
		},
		{
			s1: []string{"2", "2", "1"},
			s2: []string{"1", "1", "2"},
		},
	}

	for i, v := range data {
		assert.False(t, SliceEqualsString(v.s1, v.s2), "%d", i)
	}
}

// Tests property deep equal.
func TestDeepEquals(t *testing.T) {
	data := []struct {
		prop enums.Property
		p1   interface{}
		p2   interface{}
		gold bool
	}{
		{
			prop: enums.PropColor,
			p1:   map[string]interface{}{"r": 1, "g": 2, "b": 3},
			p2:   map[string]interface{}{"r": 1, "b": 3, "g": 2},
			gold: true,
		},
		{
			prop: enums.PropPicture,
			p1:   "t",
			p2:   "t",
			gold: false,
		},
		{
			prop: enums.PropSensorType,
			p1:   enums.SenButton,
			p2:   enums.SenGeneric,
			gold: true,
		},
		{
			prop: enums.PropHumidity,
			p1:   10.01,
			p2:   10.02,
			gold: false,
		},
	}

	for _, v := range data {
		assert.Equal(t, v.gold, PropertyDeepEqual(v.p1, v.p2, v.prop), v.prop.String())
	}
}

// Tests debug stringify.
func TestDebugObjectStringify(t *testing.T) {
	test := &struct {
		p1 string
		p2 int
	}{
		p1: "1",
		p2: 2,
	}

	res := DebugObjectStringify(test)

	assert.Equal(t, res, "&struct { p1 string; p2 int }{\n  p1: \"1\",\n  p2: 2,\n}")
}
