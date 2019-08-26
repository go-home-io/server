package helpers

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
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
			cmd: -1,
		},
		{
			input: map[string]interface{}{"title": "test", "params": map[string]string{"p1": "P"}},
			gold:  common.Input{Title: "test", Params: map[string]string{"p1": "P"}},
			prop:  enums.PropInput,
			cmd:   enums.CmdInput,
		},
	}

	for _, v := range data {
		p, err := PropertyFixYaml(v.input, v.prop)
		require.NoError(t, err, "fix yaml %s", v.prop.String())
		require.True(t, PropertyDeepEqual(p, v.gold, v.prop), "equal %s", v.prop.String())

		if -1 == v.cmd {
			continue
		}

		p, err = CommandPropertyFixYaml(v.input, v.cmd)
		assert.NoError(t, err, "cmd fix yaml %s", v.cmd.String())
		assert.True(t, PropertyDeepEqual(p, v.gold, v.prop), "cmd equal %s", v.prop.String())
	}
}

// Tests unmarshal properties.
func TestUnmarshalProperty(t *testing.T) {
	data := []struct {
		prop enums.Property
		in   interface{}
		out  interface{}
	}{
		{
			prop: enums.PropOn,
			in:   true,
			out:  true,
		},
		{
			prop: enums.PropBrightness,
			in:   map[interface{}]interface{}{"value": 88},
			out:  common.Percent{Value: 88},
		},
		{
			prop: enums.PropColor,
			in:   map[string]interface{}{"r": 129, "g": 10, "b": 23},
			out:  common.Color{R: 129, G: 10, B: 23},
		},
		{
			prop: enums.PropHumidity,
			in:   map[interface{}]interface{}{"value": 88},
			out:  common.Float{Value: 88},
		},
		{
			prop: enums.PropInput,
			in:   map[string]interface{}{"title": "test", "params": map[string]string{"p1": "P"}},
			out:  common.Input{Title: "test", Params: map[string]string{"p1": "P"}},
		},
	}

	for _, v := range data {
		p, err := UnmarshalProperty(v.in, v.prop)
		assert.NoError(t, err, "unmarshal %s", v.prop.String())
		assert.True(t, PropertyDeepEqual(p, v.out, v.prop), "equal %s", v.prop.String())
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
		assert.True(t, reflect.DeepEqual(out, v.out), v.prop.String())
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
		assert.True(t, reflect.DeepEqual(out, v.out), v.prop.String())
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
		assert.True(t, reflect.DeepEqual(r, v.out), v.prop.String())
	}
}
