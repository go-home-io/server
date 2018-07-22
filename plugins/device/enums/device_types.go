//go:generate enumer -type=DeviceType -transform=kebab -trimprefix=Dev -json -text -yaml

package enums

// DeviceType describes enum with known device types.
type DeviceType int

const (
	// DevUnknown describes unknown device type.
	DevUnknown DeviceType = iota
	// DevHub describes hub device type.
	DevHub
	// DevLight describes lights device type.
	DevLight
	// DevSwitch describes switch device type.
	DevSwitch
	// DevSensor describes sensor device type.
	DevSensor
)

// SliceContainsDeviceType is a helper Slice.contains.
func SliceContainsDeviceType(s []DeviceType, e DeviceType) bool {
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
