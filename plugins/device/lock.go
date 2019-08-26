package device

import "reflect"

// ILock defines lock plugin interface.
type ILock interface {
	IDevice
	Load() (*LockState, error)
	Update() (*LockState, error)
	On() error
	Off() error
	Toggle() error
}

// LockState returns information about known lock.
type LockState struct {
	GenericDeviceState

	On           bool   `json:"on"`
	BatteryLevel uint8  `json:"battery_level"`
}

// TypeLock is a syntax sugar around ILock type.
var TypeLock = reflect.TypeOf((*ILock)(nil)).Elem()
