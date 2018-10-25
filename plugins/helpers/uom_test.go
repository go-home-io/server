package helpers

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/plugins/device/enums"
)

// Tests conversion.
func TestUOMConvert(t *testing.T) {
	for _, v := range convertRequired {
		out := UOMConvert(100.0, v, enums.UOMMetric, enums.UOMImperial)
		outI := UOMConvertInterface(100.0, v, enums.UOMMetric, enums.UOMImperial)

		in := UOMConvert(out, v, enums.UOMImperial, enums.UOMMetric)
		inI := UOMConvertInterface(outI, v, enums.UOMImperial, enums.UOMMetric)

		assert.False(t, math.Abs(in-100) > 0.2, v.String())
		assert.False(t, math.Abs(inI.(float64)-100) > 0.2, v.String())
		assert.Equal(t, outI.(float64), out, "in %s", v.String())
		assert.Equal(t, inI.(float64), in, "out %s", v.String())
	}
}

// Tests strings conversion.
func TestUOMConvertString(t *testing.T) {
	str := map[enums.Property]map[enums.UOM]string{
		enums.PropTemperature: {enums.UOMImperial: "f", enums.UOMMetric: "c"},
		enums.PropWindSpeed:   {enums.UOMImperial: "mph", enums.UOMMetric: "kmh"},
		enums.PropVisibility:  {enums.UOMImperial: "mi", enums.UOMMetric: "km"},
		enums.PropPressure:    {enums.UOMImperial: "inHg", enums.UOMMetric: "mbar"},
	}
	for _, v := range convertRequired {
		s, ok := str[v]
		if !ok {
			continue
		}

		out := UOMConvertString(100.0, v, s[enums.UOMImperial], enums.UOMMetric)
		in := UOMConvertString(out, v, s[enums.UOMMetric], enums.UOMImperial)

		assert.False(t, math.Abs(in-100) > 0.2, "in %s", v.String())
		assert.False(t, math.Abs(out-in) < 0.2, "out %s", v.String())
	}
}
