package device

import (
	"errors"
	"reflect"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
)

// ConstructDevice has data required for a new device loader.
type ConstructDevice struct {
	DeviceName string
	DeviceType enums.DeviceType
	ConfigName string
	RawConfig  string
	Settings   providers.ISettingsProvider

	StatusUpdatesChan chan *UpdateEvent
	DiscoveryChan     chan *NewDeviceDiscoveredEvent
}

// LoadDevice validates device type and loads requested plugin.
func LoadDevice(ctor *ConstructDevice) ([]IDeviceWrapperProvider, error) {
	if ctor.DeviceType == enums.DevHub {
		return loadHub(ctor)
	}

	pluginLogger := ctor.Settings.PluginLogger(systems.SysDevice, ctor.DeviceName)

	wrappers := make([]IDeviceWrapperProvider, 1)

	loadData := &device.InitDataDevice{
		Logger:                pluginLogger,
		Secret:                ctor.Settings.Secrets(),
		DeviceDiscoveredChan:  make(chan *device.DiscoveredDevices, 3),
		DeviceStateUpdateChan: make(chan *device.StateUpdateData, 10),
	}

	expectedType, err := getExpectedType(ctor.DeviceType)
	if err != nil {
		pluginLogger.Error("Failed to load device plugin", err,
			common.LogDeviceTypeToken, ctor.DeviceName)
		return nil, err
	}

	pluginLoadRequest := &providers.PluginLoadRequest{
		InitData:       loadData,
		RawConfig:      []byte(ctor.RawConfig),
		PluginProvider: ctor.DeviceName,
		SystemType:     systems.SysDevice,
		ExpectedType:   expectedType,
	}

	i, err := ctor.Settings.PluginLoader().LoadPlugin(pluginLoadRequest)
	if err != nil {
		pluginLogger.Error("Failed to load device plugin", err,
			common.LogDeviceTypeToken, ctor.DeviceName)
		return nil, err
	}

	deviceState, err := loadDevice(i, ctor.DeviceType)
	if err != nil {
		pluginLogger.Error("Failed to load device plugin", err,
			common.LogDeviceTypeToken, ctor.DeviceName)
		return nil, err
	}

	deviceCtor := &wrapperConstruct{
		DeviceType:        ctor.DeviceType,
		DeviceInterface:   i,
		IsRootDevice:      true,
		Cron:              ctor.Settings.Cron(),
		DeviceConfigName:  ctor.ConfigName,
		DeviceState:       deviceState,
		LoadData:          loadData,
		Logger:            pluginLogger,
		Secret:            ctor.Settings.Secrets(),
		WorkerID:          ctor.Settings.NodeID(),
		Validator:         ctor.Settings.Validator(),
		DiscoveryChan:     ctor.DiscoveryChan,
		StatusUpdatesChan: ctor.StatusUpdatesChan,
	}

	wrappers[0] = NewDeviceWrapper(deviceCtor)

	return wrappers, nil
}

// Returns device interface, depends on it's type.
func getExpectedType(deviceType enums.DeviceType) (reflect.Type, error) {
	switch deviceType {
	case enums.DevLight:
		return device.TypeLight, nil
	}

	return nil, errors.New("unknown device type")
}

// Executes load method of the device implementation.
func loadDevice(deviceInterface interface{}, deviceType enums.DeviceType) (interface{}, error) {
	switch deviceType {
	case enums.DevLight:
		return deviceInterface.(device.ILight).Load()
	}

	return nil, errors.New("unknown device type")
}
