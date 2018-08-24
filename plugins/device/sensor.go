package device

import (
	"reflect"

	"github.com/go-home-io/server/plugins/device/enums"
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
	Power        float64          `json:"power"`
	Temperature  float64          `json:"temperature"`
	Humidity     float64          `json:"humidity"`
	BatteryLevel uint8            `json:"battery_level"`
	On           bool             `json:"on"`
	Press        bool             `json:"press"`
	Click        bool             `json:"click"`
	DoubleClick  bool             `json:"double_click"`
}

// TypeSensor is a syntax sugar around ISensor type.
var TypeSensor = reflect.TypeOf((*ISensor)(nil)).Elem()
