// Package device contains devices plugin definitions.
package device

import (
	"time"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
)

// IDevice defines generic device plugin interface.
type IDevice interface {
	Init(*InitDataDevice) error
	Unload()
	GetName() string
	GetSpec() *Spec
}

// Spec contains information about the device.
type Spec struct {
	UpdatePeriod           time.Duration
	SupportedCommands      []enums.Command
	SupportedProperties    []enums.Property
	PostCommandDeferUpdate time.Duration
}

// StateUpdateData contains updated state of the device.
type StateUpdateData struct {
	State interface{}
}

// DiscoveredDevices contains information of a newly discovered devices.
type DiscoveredDevices struct {
	Type      enums.DeviceType
	Interface interface{}
	State     interface{}
}

// InitDataDevice has data required for initializing a new device.
type InitDataDevice struct {
	Logger common.ILoggerProvider
	Secret common.ISecretProvider
	UOM    enums.UOM

	DeviceStateUpdateChan chan *StateUpdateData
	DeviceDiscoveredChan  chan *DiscoveredDevices
}
