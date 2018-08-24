//go:generate enumer -type=Property -transform=snake -trimprefix=Prop -json -text -yaml

package enums

// Property describes enum with known devices' properties.
type Property int

const (
	// PropOn describes On/Off status of the device.
	PropOn Property = iota
	// PropColor describes color of the device.
	PropColor
	// PropNumDevices describes number of devices per hub.
	PropNumDevices
	// PropTransitionTime describes transition time of the device.
	PropTransitionTime
	// PropBrightness describes brightness of the device.
	PropBrightness
	// PropScenes describes list of scenes available for the device.
	PropScenes
	// PropPower describes device consumption power.
	PropPower
	// PropTemperature describes temperature.
	PropTemperature
	// PropBatteryLevel describes device battery level.
	PropBatteryLevel
	// PropSunrise describes sunrise time.
	PropSunrise
	// PropSunset describes sunset time.
	PropSunset
	// PropHumidity describes humidity.
	PropHumidity
	// PropPressure describes pressure.
	PropPressure
	// PropVisibility describes visibility.
	PropVisibility
	// PropWindDirection describes wind direction.
	PropWindDirection
	// PropWindSpeed describes wind speed.
	PropWindSpeed
	// PropClick describes single click button.
	PropClick
	// PropDoubleClick describes double click button.
	PropDoubleClick
	// PropPress describes long press button.
	PropPress
	// PropSensorType describes possible sensor type.
	PropSensorType
)

// AllowedProperties contains set of all possible allowed properties per device type.
var AllowedProperties = map[DeviceType][]Property{
	DevHub:    {PropNumDevices},
	DevLight:  {PropOn, PropColor, PropTransitionTime, PropBrightness, PropScenes},
	DevSwitch: {PropOn, PropPower},
	DevSensor: {PropSensorType, PropOn, PropBatteryLevel, PropPower, PropTemperature, PropHumidity,
		PropClick, PropDoubleClick, PropPress},
	DevWeather: {PropTemperature, PropSunrise, PropSunset, PropHumidity, PropPressure,
		PropVisibility, PropWindDirection, PropWindSpeed},
}

// SliceContainsProperty checks whether slice contains certain property.
func SliceContainsProperty(s []Property, e Property) bool {
	if nil == s {
		return false
	}
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// GetPropertyName returns actual property name.
func (i Property) GetPropertyName() string {
	return transformCommandOrProperty(i.String(), "_")
}

// IsPropertyAllowed checks whether property is allowed to certain device type.
func (i Property) IsPropertyAllowed(deviceType DeviceType) bool {
	slice, ok := AllowedProperties[deviceType]
	if !ok {
		return false
	}

	return SliceContainsProperty(slice, i)
}
