package helpers

import (
	"math"
	"testing"

	"github.com/go-home-io/server/plugins/device/enums"
)

// Tests conversion.
func TestUOMConvert(t *testing.T) {

	for _, v := range convertRequired {
		out := UOMConvert(100.0, v, enums.UOMMetric, enums.UOMImperial)
		outI := UOMConvertInterface(100.0, v, enums.UOMMetric, enums.UOMImperial)

		in := UOMConvert(out, v, enums.UOMImperial, enums.UOMMetric)
		inI := UOMConvertInterface(outI, v, enums.UOMImperial, enums.UOMMetric)

		if math.Abs(in-100) > 0.2 || math.Abs(inI.(float64)-100) > 0.2 {
			t.Error("Failed " + v.String())
			t.Fail()
		}

		if out != outI.(float64) || in != inI.(float64) {
			t.Error("Interface failed " + v.String())
		}
	}
}

// Tests strings conversion.
func TestUOMConvertString(t *testing.T) {
	str := map[enums.Property]map[enums.UOM]string{
		enums.PropTemperature: {enums.UOMImperial: "f", enums.UOMMetric: "c"},
		enums.PropWindSpeed:   {enums.UOMImperial: "mph", enums.UOMMetric: "kmh"},
		enums.PropVisibility:  {enums.UOMImperial: "mi", enums.UOMMetric: "km"},
		enums.PropPressure:    {enums.UOMImperial: "in", enums.UOMMetric: "bar"},
	}
	for _, v := range convertRequired {
		s, ok := str[v]
		if !ok {
			continue
		}

		out := UOMConvertString(100.0, v, s[enums.UOMImperial], enums.UOMMetric)
		in := UOMConvertString(out, v, s[enums.UOMMetric], enums.UOMImperial)

		if math.Abs(in-100) > 0.2 || math.Abs(out-in) < 0.2 {
			t.Error("String failed " + v.String())
			t.Fail()
		}
	}
}
