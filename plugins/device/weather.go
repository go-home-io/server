package device

import (
	"reflect"
)

// IWeather defines weather device.
type IWeather interface {
	IDevice
	Load() (*WeatherState, error)
	Update() (*WeatherState, error)
}

// WeatherState describes weather device state
type WeatherState struct {
	GenericDeviceState

	Humidity      float64 `json:"humidity"`
	Pressure      float64 `json:"pressure"`
	Visibility    float64 `json:"visibility"`
	WindDirection float64 `json:"wind_direction"`
	WindSpeed     float64 `json:"wind_speed"`
	Temperature   float64 `json:"temperature"`
	Sunrise       string  `json:"sunrise"`
	Sunset        string  `json:"sunset"`
	Description   string  `json:"description"`
}

// TypeWeather is a syntax sugar around IWeather type.
var TypeWeather = reflect.TypeOf((*IWeather)(nil)).Elem()
