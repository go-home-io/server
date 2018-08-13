package storage

import (
	"reflect"
	"testing"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
)

// Test conversions.
func TestConverters(t *testing.T) {
	data := []struct {
		In       interface{}
		Expected interface{}
		Property enums.Property
		TwoWay   bool
	}{
		{
			In: common.Color{
				R: 10,
				G: 20,
				B: 30,
			},
			Expected: `{"r":10,"g":20,"b":30}`,
			Property: enums.PropColor,
			TwoWay:   true,
		},
		{
			In:       common.Float{Value: 75},
			Expected: 75.0,
			Property: enums.PropHumidity,
			TwoWay:   false,
		},
		{
			In:       common.Percent{Value: 75},
			Expected: uint8(75),
			Property: enums.PropBrightness,
			TwoWay:   false,
		},
		{
			In:       []string{"1", "2"},
			Expected: nil,
			Property: enums.PropScenes,
			TwoWay:   false,
		},
		{
			In:       true,
			Expected: true,
			Property: enums.PropOn,
			TwoWay:   true,
		},
	}

	for _, v := range data {
		c, err := PropertySave(v.Property, v.In)
		if err != nil || !reflect.DeepEqual(c, v.Expected) {
			t.Error("Failed to save " + v.Property.String())
			t.Fail()
		}

		o, err := PropertyLoad(v.Property, c)
		if err != nil || (v.TwoWay && !helpers.PropertyDeepEqual(v.In, o, v.Property)) {
			t.Error("Failed to load " + v.Property.String())
			t.Fail()
		}
	}
}
