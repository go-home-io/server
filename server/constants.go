//go:generate enumer -type=entityStatus -transform=snake -trimprefix=entity -json -text -yaml

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

// entityStatus describes enum with entity load status.
type entityStatus int

const (
	// entityAssignmentFailed describes assignment failed status.
	entityAssignmentFailed entityStatus = iota
	// entityAssigned describes success assignment status.
	entityAssigned
	// entityWrongWorker describes notification from a wrong worker status.
	entityWrongWorker
	// entityLoaded describes success load status.
	entityLoaded
	// entityLoadFailed describes error while loading status.
	entityLoadFailed
)
