package device

// ErrUnknownDeviceType defines an unknown device type error.
type ErrUnknownDeviceType struct {
}

// Error formats output.
func (e *ErrUnknownDeviceType) Error() string {
	return "unknown device type"
}

// ErrNoDataFromPlugin defines an empty response from plugin error.
type ErrNoDataFromPlugin struct {
}

// Error formats output.
func (*ErrNoDataFromPlugin) Error() string {
	return "plugin didn't return any data"
}
