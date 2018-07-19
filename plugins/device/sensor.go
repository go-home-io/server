package device

import "reflect"

// ISensor defines sensor plugin interface.
type ISensor interface {
	IDevice
	Load() (*SensorState, error)
	Update() (*SensorState, error)
}

// SensorState returns information about known sensor.
type SensorState struct {
	Power        float64 `json:"power"`
	Temperature  float64 `json:"temperature"`
	BatteryLevel uint8   `json:"battery_level"`
	On           bool    `json:"on"`
}

// TypeSensor is a syntax sugar around ISensor type.
var TypeSensor = reflect.TypeOf((*ISensor)(nil)).Elem()
