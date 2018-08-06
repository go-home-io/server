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
	GetDevice(string) *KnownDevice
	PushMasterDeviceUpdate(*MasterDeviceUpdate)
}

// MasterDeviceUpdate contains data required for pushing update for device running on master.
// Usually it's a group.
type MasterDeviceUpdate struct {
	ID       string
	Name     string
	State    map[string]interface{}
	Commands []string
	Type     enums.DeviceType
}

// KnownDevice contains data about known device.
type KnownDevice struct {
	Worker   string
	Commands []string
}
