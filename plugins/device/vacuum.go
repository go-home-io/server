package device

import (
	"reflect"

	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
)

// IVacuum defines vacuum device type.
type IVacuum interface {
	IDevice
	Load() (*VacuumState, error)
	Update() (*VacuumState, error)
	On() error
	Off() error
	Pause() error
	Dock() error
	FindMe() error
	SetFanSpeed(common.Percent) error
}

// VacuumState describes vacuum state.
type VacuumState struct {
	GenericDeviceState

	VacStatus    enums.VacStatus `json:"vac_status"`
	Duration     int             `json:"duration"`
	BatteryLevel uint8           `json:"battery_level"`
	FanSpeed     uint8           `json:"fan_speed"`
	Area         float64         `json:"area"`
}

// TypeVacuum is a syntax sugar around IVacuum type.
var TypeVacuum = reflect.TypeOf((*IVacuum)(nil)).Elem()
