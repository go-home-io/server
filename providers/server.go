package providers

import (
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/gobwas/glob"
)

// IServerProvider defines server interface which is required by
// some other internal systems.
type IServerProvider interface {
	Start()
	InternalCommandInvokeDeviceCommand(deviceRegexp glob.Glob, cmd enums.Command, data map[string]interface{})
}
