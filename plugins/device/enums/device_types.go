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
)
