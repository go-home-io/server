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
)

// AllowedProperties contains set of all possible allowed properties per device type.
var AllowedProperties = map[DeviceType][]Property{
	DevHub:   {PropNumDevices},
	DevLight: {PropOn, PropColor, PropTransitionTime, PropBrightness, PropScenes},
}

// SliceContainsProperty checks whether slice contains certain property.
func SliceContainsProperty(s []Property, e Property) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
