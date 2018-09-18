package device

import (
	"reflect"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
)

// IVacuum defines vacuum device type.
type IVacuum interface {
	IDevice
	Load() (*VacuumState, error)
	On() error
	Off() error
	Pause() error
	Dock() error
	FindMe() error
	SetFanSpeed(common.Percent) error
	Update() (*VacuumState, error)
}

// VacuumState describes vacuum state.
type VacuumState struct {
	VacStatus    enums.VacStatus `json:"vac_status"`
	BatteryLevel uint8           `json:"battery_level"`
	Area         float64         `json:"area"`
	Duration     int             `json:"duration"`
	FanSpeed     uint8           `json:"fan_speed"`
}

// TypeVacuum is a syntax sugar around IVacuum type.
var TypeVacuum = reflect.TypeOf((*IVacuum)(nil)).Elem()
