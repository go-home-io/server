package helpers

import (
	"strings"

	"github.com/go-home-io/server/plugins/device/enums"
)

// List of properties, which needs to be converted.
var convertRequired = []enums.Property{
	enums.PropTemperature, enums.PropWindSpeed,
	enums.PropVisibility, enums.PropPressure,
	enums.PropArea,
}

// UOMConvertString converts properties from one system to another.
func UOMConvertString(value float64, property enums.Property, currentUOM string, desiredUOM enums.UOM) float64 {
	current := getUOM(property, currentUOM)
	return UOMConvert(value, property, current, desiredUOM)
}

// UOMConvertInterface checks whether property conversion is required and do the conversion.
func UOMConvertInterface(value interface{}, property enums.Property,
	currentUOM enums.UOM, desiredUOM enums.UOM) interface{} {
	if !enums.SliceContainsProperty(convertRequired, property) {
		return value
	}

	val, ok := value.(float64)
	if !ok {
		return value
	}

	return UOMConvert(val, property, currentUOM, desiredUOM)
}

// UOMConvert converts properties from one system to another.
func UOMConvert(value float64, property enums.Property, currentUOM enums.UOM, desiredUOM enums.UOM) float64 {
	if desiredUOM == currentUOM {
		return value
	}

	if currentUOM == enums.UOMMetric {
		return convertMetricToImperial(value, property)
	}

	return convertImperialToMetric(value, property)
}

// Converts imperial to metric
func convertImperialToMetric(value float64, property enums.Property) float64 {
	switch property {
	case enums.PropTemperature:
		return (value - 32.0) / 1.8
	case enums.PropWindSpeed, enums.PropVisibility:
		return value / 1.609344
	case enums.PropPressure:
		return value * 33.864
	case enums.PropArea:
		return value / 10.7639
	}

	return value
}

// Converts metric to imperial.
func convertMetricToImperial(value float64, property enums.Property) float64 {
	switch property {
	case enums.PropTemperature:
		return value*1.8 + 32.0
	case enums.PropWindSpeed, enums.PropVisibility:
		return 1.609344 * value
	case enums.PropPressure:
		return value / 33.864
	case enums.PropArea:
		return value * 10.7639
	}

	return value
}

// Returns current UOM.
func getUOM(property enums.Property, current string) enums.UOM {
	current = strings.ToLower(current)
	isImp := false
	switch property {
	case enums.PropTemperature:
		isImp = "f" == current
	case enums.PropWindSpeed:
		isImp = "mph" == current
	case enums.PropPressure:
		isImp = "in" == current
	case enums.PropVisibility:
		isImp = "mi" == current
	}

	if isImp {
		return enums.UOMImperial
	}

	return enums.UOMMetric
}
