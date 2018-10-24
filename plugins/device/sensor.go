package device

import (
	"reflect"

	"go-home.io/x/server/plugins/device/enums"
)

// ISensor defines sensor plugin interface.
type ISensor interface {
	IDevice
	Load() (*SensorState, error)
	Update() (*SensorState, error)
}

// SensorState returns information about known sensor.
type SensorState struct {
	SensorType   enums.SensorType `json:"sensor_type"`
	User         string           `json:"user"`
	Power        float64          `json:"power"`
	Temperature  float64          `json:"temperature"`
	Humidity     float64          `json:"humidity"`
	Pressure     float64          `json:"pressure"`
	BatteryLevel uint8            `json:"battery_level"`
	On           bool             `json:"on"`
	Press        bool             `json:"press"`
	Click        bool             `json:"click"`
	DoubleClick  bool             `json:"double_click"`
}

// TypeSensor is a syntax sugar around ISensor type.
var TypeSensor = reflect.TypeOf((*ISensor)(nil)).Elem()
