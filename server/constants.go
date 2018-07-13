package server

// URLTokens describes enum with known API tokens.
type URLTokens string

const (
	// URLDeviceID describes device ID URL param.
	URLDeviceID URLTokens = "deviceID"
	// URLCommandName describes device command name URL param.
	URLCommandName URLTokens = "commandName"
)
