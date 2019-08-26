package device

import "reflect"

// ISwitch defines switch plugin interface.
type ISwitch interface {
	IDevice
	Load() (*SwitchState, error)
	On() error
	Off() error
	Toggle() error
	Update() (*SwitchState, error)
}

// SwitchState returns information about known switch.
type SwitchState struct {
	GenericDeviceState

	On    bool    `json:"on"`
	Power float64 `json:"power"`
}

// TypeSwitch is a syntax sugar around ISwitch type.
var TypeSwitch = reflect.TypeOf((*ISwitch)(nil)).Elem()
