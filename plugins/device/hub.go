package device

import "reflect"

// IHub defines hub plugin interface.
type IHub interface {
	IDevice
	Load() (*HubLoadResult, error)
	Update() (*HubLoadResult, error)
}

// HubLoadResult returns information about known devices for the hub.
type HubLoadResult struct {
	State   *HubState
	Devices []*DiscoveredDevices
}

// HubState contains device state data.
type HubState struct {
	NumDevices int `json:"num_devices"`
}

// TypeHub is a syntax sugar around IHub type.
var TypeHub = reflect.TypeOf((*IHub)(nil)).Elem()
