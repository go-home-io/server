package helpers

import (
	"reflect"
	"testing"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
)

type testData struct {
	input interface{}
	gold  interface{}
	prop  enums.Property
	cmd   enums.Command
}

// Tests properties conversion.
func TestProperties(t *testing.T) {
	data := []testData{
		{
			input: map[interface{}]interface{}{"r": 10, "g": 20, "b": 30},
			gold:  common.Color{R: 10, G: 20, B: 30},
			prop:  enums.PropColor,
			cmd:   enums.CmdSetColor,
		},
		{
			input: true,
			gold:  true,
			prop:  enums.PropOn,
			cmd:   -1,
		},
		{
			input: 10,
			gold:  common.Percent{Value: 10},
			prop:  enums.PropBrightness,
			cmd:   enums.CmdSetBrightness,
		},
		{
			input: "test_user",
			gold:  "test_user",
			prop:  enums.PropUser,
			cmd:   -1,
		},
		{
			input: []string{"s1", "s2"},
			gold:  []string{"s1", "s2"},
			prop:  enums.PropScenes,
		},
	}

	for _, v := range data {
		p, err := PropertyFixYaml(v.input, v.prop)
		if err != nil {
			t.Error("failed fix property " + v.prop.String())
			t.FailNow()
		}

		if !PropertyDeepEqual(p, v.gold, v.prop) {
			t.Error("failed deep equal property " + v.prop.String())
			t.FailNow()
		}

		if -1 == v.cmd {
			continue
		}

		p, err = CommandPropertyFixYaml(v.input, v.cmd)

		if err != nil {
			t.Error("failed fix command " + v.cmd.String())
			t.Fail()
		}

		if !PropertyDeepEqual(p, v.gold, v.prop) {
			t.Error("failed deep equal command " + v.prop.String())
			t.Fail()
		}
	}
}

// Tests unmarshal properties.
func TestUnmarshalProperty(t *testing.T) {
	in := map[enums.Property]interface{}{
		enums.PropOn:         true,
		enums.PropBrightness: map[interface{}]interface{}{"value": 88},
		enums.PropColor:      map[string]interface{}{"r": 129, "g": 10, "b": 23},
		enums.PropHumidity:   map[interface{}]interface{}{"value": 88},
	}

	out := map[enums.Property]interface{}{
		enums.PropOn:         true,
		enums.PropBrightness: common.Percent{Value: 88},
		enums.PropColor:      common.Color{R: 129, G: 10, B: 23},
		enums.PropHumidity:   common.Float{Value: 88},
	}

	for i, v := range in {
		p, err := UnmarshalProperty(v, i)
		if err != nil {
			t.Error("Unmarshal failed on " + i.String())
			t.Fail()
		}

		if !PropertyDeepEqual(p, out[i], i) {
			t.Error("Equal failed on " + i.String())
			t.Fail()
		}
	}
}

// Tests property converters.
func TestPlainProperty(t *testing.T) {
	data := []struct {
		in   interface{}
		prop enums.Property
		out  interface{}
	}{
		{
			in:   enums.VacDocked,
			prop: enums.PropVacStatus,
			out:  enums.VacDocked,
		},
		{
			in:   true,
			prop: enums.PropOn,
			out:  true,
		},
		{
			in:   common.Color{R: 10, G: 20, B: 30},
			prop: enums.PropColor,
			out:  "r:10,g:20,b:30",
		},
		{
			in:   common.Percent{Value: 10},
			prop: enums.PropBatteryLevel,
			out:  uint8(10),
		},
		{
			in:   common.Int{Value: 20},
			prop: enums.PropDuration,
			out:  20,
		},
		{
			in:   common.Float{Value: 30},
			prop: enums.PropTemperature,
			out:  30.0,
		},
	}

	for _, v := range data {
		out := PlainProperty(v.in, v.prop)
		if !reflect.DeepEqual(out, v.out) {
			t.Error("Failed " + v.prop.String())
			t.Fail()
		}
	}
}

// Tests property converters.
func TestPlainValueProperty(t *testing.T) {
	data := []struct {
		in   interface{}
		prop enums.Property
		out  interface{}
	}{
		{
			in:   enums.VacDocked,
			prop: enums.PropVacStatus,
			out:  enums.VacDocked,
		},
		{
			in:   true,
			prop: enums.PropOn,
			out:  true,
		},
		{
			in:   common.Color{R: 10, G: 20, B: 30},
			prop: enums.PropColor,
			out:  common.Color{R: 10, G: 20, B: 30},
		},
		{
			in:   common.Percent{Value: 10},
			prop: enums.PropBatteryLevel,
			out:  uint8(10),
		},
		{
			in:   common.Int{Value: 20},
			prop: enums.PropDuration,
			out:  20,
		},
		{
			in:   common.Float{Value: 30},
			prop: enums.PropTemperature,
			out:  30.0,
		},
	}

	for _, v := range data {
		out := PlainValueProperty(v.in, v.prop)
		if !reflect.DeepEqual(out, v.out) {
			t.Error("Failed " + v.prop.String())
			t.Fail()
		}
	}
}

// Tests number fix properties.
func TestPropertyFixNum(t *testing.T) {
	in := []struct {
		prop enums.Property
		val  interface{}
		out  interface{}
	}{
		{
			prop: enums.PropBatteryLevel,
			val:  float64(10),
			out:  uint8(10),
		},
		{
			prop: enums.PropTransitionTime,
			val:  float64(10),
			out:  uint16(10),
		},
		{
			prop: enums.PropDuration,
			val:  float64(10),
			out:  int(10),
		},
		{
			prop: enums.PropArea,
			val:  float64(10),
			out:  float64(10),
		},
	}

	for _, v := range in {
		r := PropertyFixNum(v.val, v.prop)
		if !reflect.DeepEqual(r, v.out) {
			t.Error("Failed on " + v.prop.String())
			t.Fail()
		}
	}
}
