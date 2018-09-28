package device

import "github.com/go-home-io/server/plugins/device/enums"

// IProcessor defines device post-processor.
type IProcessor interface {
	IsPropertyGood(enums.Property, interface{}) (bool, interface{})
}

// Constructs a new device processor if required for the device.
func newDeviceProcessor(deviceType enums.DeviceType, rawConfig string) IProcessor {
	switch deviceType {
	case enums.DevCamera:
		return newCameraProcessor(rawConfig)
	}

	return nil
}
