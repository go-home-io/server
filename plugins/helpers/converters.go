package helpers

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

// Internal property type.
type propertyType int

const (
	propFloat propertyType = iota
	propColor
	propString
	propStringSlice
	propEnum
	propBool
	propPercent
	propInt
)

// Converts device property to its type.
func getPropertyType(p enums.Property) propertyType {
	switch p {
	case enums.PropColor:
		return propColor
	case enums.PropScenes:
		return propStringSlice
	case enums.PropSensorType, enums.PropVacStatus:
		return propEnum
	case enums.PropPicture, enums.PropUser:
		return propString
	case enums.PropOn, enums.PropClick, enums.PropDoubleClick, enums.PropPress:
		return propBool
	case enums.PropBrightness, enums.PropBatteryLevel, enums.PropFanSpeed:
		return propPercent
	case enums.PropDuration, enums.PropDistance:
		return propInt
	}

	return propFloat
}

// Fixing properties.
func propertyFix(x interface{}, p enums.Property,
	f func(interface{}, interface{}) (interface{}, error)) (interface{}, error) {
	if nil == x {
		return x, nil
	}

	switch getPropertyType(p) {
	case propColor:
		return convertProperty(x, &common.Color{})
	case propStringSlice, propEnum, propString:
		return x, nil
	case propBool:
		r, ok := x.(bool)
		if !ok {
			return nil, errors.New("error converting bool")
		}
		return r, nil
	case propPercent:
		return f(x, &common.Percent{})
	case propInt:
		return f(x, &common.Int{})
	default:
		return f(x, &common.Float{})
	}

}

// PropertyFixYaml help with issues of unknown configs' data structure.
// For example default
// field	interface{}
// with data
// color:
//   r : 120
// 	 g : 120
//   b : 120
// will be un-marshaled as map[interface{}] interface{}.
// Which prevents normal deep compare.
func PropertyFixYaml(x interface{}, p enums.Property) (interface{}, error) {
	return propertyFix(x, p, convertValueProperty)
}

// UnmarshalProperty returns type used by property from it's map[{interface}]interface{}
// or interface{} representation distributed through FanOut channel.
func UnmarshalProperty(x interface{}, p enums.Property) (interface{}, error) {
	return propertyFix(x, p, convertProperty)
}

// PlainProperty converts common.* data into plain properties to use with mappers.
func PlainProperty(x interface{}, p enums.Property) interface{} {
	if nil == x {
		return x
	}

	switch getPropertyType(p) {
	case propColor:
		c := x.(common.Color)
		return fmt.Sprintf("r:%d,g:%d,b:%d", c.R, c.G, c.B)
	case propStringSlice, propEnum, propString, propBool:
		return x
	case propPercent:
		return x.(common.Percent).Value
	case propInt:
		return x.(common.Int).Value
	default:
		return x.(common.Float).Value
	}
}

// PlainValueProperty converts value-based property to simple value.
func PlainValueProperty(x interface{}, p enums.Property) interface{} {
	if nil == x {
		return x
	}

	switch getPropertyType(p) {
	case propPercent:
		return x.(common.Percent).Value
	case propInt:
		return x.(common.Int).Value
	case propFloat:
		return x.(common.Float).Value
	default:
		return x
	}
}

// PropertyDeepEqual uses some extended rules for different common types.
// For example we don't care about scenes updates, so it's always true.
func PropertyDeepEqual(x, y interface{}, p enums.Property) bool {
	switch p {
	case enums.PropPicture:
		// Picture is pre-processed.
		return false
	case enums.PropScenes, enums.PropSensorType:
		// No updates for scenes
		return true
	default:
		return cmp.Equal(x, y)
	}
}

// CommandPropertyFixYaml fixes properties similar to PropertyFixYaml method.
func CommandPropertyFixYaml(x interface{}, c enums.Command) (interface{}, error) {
	if nil == x {
		return x, nil
	}

	switch c {
	case enums.CmdOn, enums.CmdOff, enums.CmdToggle, enums.CmdFindMe, enums.CmdDock, enums.CmdPause:
		return nil, nil
	case enums.CmdSetBrightness, enums.CmdSetFanSpeed:
		return convertValueProperty(x, &common.Percent{})
	case enums.CmdSetTransitionTime:
		return convertValueProperty(x, &common.Int{})
	case enums.CmdSetColor:
		return convertProperty(x, &common.Color{})
	}

	return x, nil
}

// PropertyFixNum fixes float64 values after templating.
func PropertyFixNum(x interface{}, p enums.Property) interface{} {
	if nil == x {
		return x
	}

	switch p {
	case enums.PropBatteryLevel, enums.PropBrightness, enums.PropFanSpeed:
		return uint8(x.(float64))
	case enums.PropTransitionTime:
		return uint16(x.(float64))
	case enums.PropDuration, enums.PropDistance:
		return int(x.(float64))
	}

	return x
}

// Converts to default value-based property.
func convertValueProperty(from, to interface{}) (interface{}, error) {
	wrap := map[string]interface{}{"value": from}
	return convertProperty(wrap, to)
}

// Converts property to target type.
func convertProperty(from, to interface{}) (interface{}, error) {
	data, err := yaml.Marshal(from)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, to)
	return reflect.ValueOf(to).Elem().Interface(), err
}
