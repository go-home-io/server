package providers

import (
	"github.com/gobwas/glob"
	"go-home.io/x/server/plugins/device/enums"
)

// IServerProvider defines server interface which is required by
// some other internal systems.
type IServerProvider interface {
	Start()
	InternalCommandInvokeDeviceCommand(glob.Glob, enums.Command, map[string]interface{})
	SendNotificationCommand(glob.Glob, string)
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
	Type     enums.DeviceType
	Commands []string
}
