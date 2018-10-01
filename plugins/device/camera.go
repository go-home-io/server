package device

import "reflect"

// ICamera defines camera plugin interface.
type ICamera interface {
	IDevice
	Load() (*CameraState, error)
	Update() (*CameraState, error)
	TakePicture() error
}

// CameraState contains information about current camera's state.
type CameraState struct {
	Picture  string `json:"picture"`
	Distance int    `json:"distance"`
}

// TypeCamera is a syntax sugar around ICamera type.
var TypeCamera = reflect.TypeOf((*ICamera)(nil)).Elem()
