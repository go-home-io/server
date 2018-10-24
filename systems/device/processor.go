package device

import "go-home.io/x/server/plugins/device/enums"

// IProcessor defines device post-processor.
type IProcessor interface {
	IsExtraProperty(enums.Property) bool
	GetExtraSupportPropertiesSpec() []enums.Property
	IsPropertyGood(enums.Property, interface{}) (bool, map[enums.Property]interface{})
}

// Constructs a new device processor if required for the device.
func newDeviceProcessor(deviceType enums.DeviceType, rawConfig string) IProcessor {
	switch deviceType {
	case enums.DevCamera:
		return newCameraProcessor(rawConfig)
	}

	return nil
}
