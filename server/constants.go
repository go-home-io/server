package server

// muxKeys describes enum with known API tokens.
type muxKeys string

const (
	// urlDeviceID describes device ID URL param.
	urlDeviceID muxKeys = "deviceID"
	// urlCommandName describes device command name URL param.
	urlCommandName muxKeys = "commandName"
	// ctxtUserName describes user in the context.
	ctxtUserName muxKeys = "user"
	// routeAPI describes base api prefix.
	routeAPI = "/api/v1"
)
