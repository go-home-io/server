package device

import (
	"reflect"

	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/systems/logger"
)

// ConstructDevice has data required for a new device loader.
type ConstructDevice struct {
	DeviceName string
	DeviceType enums.DeviceType
	ConfigName string
	RawConfig  string
	Settings   providers.ISettingsProvider
	UOM        enums.UOM

	StatusUpdatesChan chan *UpdateEvent
	DiscoveryChan     chan *NewDeviceDiscoveredEvent
}

// LoadDevice validates device type and loads requested plugin.
// nolint: dupl
func LoadDevice(ctor *ConstructDevice) ([]IDeviceWrapperProvider, error) {
	logCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.Settings.PluginLogger(),
		Provider:     ctor.DeviceName,
		System:       systems.SysDevice.String(),
		ExtraFields: map[string]string{
			common.LogNameToken:       ctor.ConfigName,
			common.LogDeviceTypeToken: ctor.DeviceType.String(),
		},
	}

	log := logger.NewPluginLogger(logCtor)

	if ctor.DeviceType == enums.DevHub {
		return loadHub(ctor, log)
	}

	wrappers := make([]IDeviceWrapperProvider, 1)

	loadData := &device.InitDataDevice{
		Logger:                log,
		Secret:                ctor.Settings.Secrets(),
		UOM:                   ctor.UOM,
		DeviceDiscoveredChan:  make(chan *device.DiscoveredDevices, 3),
		DeviceStateUpdateChan: make(chan *device.StateUpdateData, 10),
	}

	expectedType, err := getExpectedType(ctor.DeviceType)
	if err != nil {
		log.Error("Failed to load device plugin", err)
		return nil, errors.Wrap(err, "unknown type")
	}

	pluginLoadRequest := &providers.PluginLoadRequest{
		InitData:           loadData,
		RawConfig:          []byte(ctor.RawConfig),
		PluginProvider:     ctor.DeviceName,
		SystemType:         systems.SysDevice,
		ExpectedType:       expectedType,
		DownloadTimeoutSec: 5,
	}

	i, err := ctor.Settings.PluginLoader().LoadPlugin(pluginLoadRequest)
	if err != nil {
		log.Error("Failed to load device plugin", err)
		return nil, errors.Wrap(err, "plugin load failed")
	}

	deviceState, err := loadDevice(i, ctor.DeviceType)
	if err != nil {
		log.Error("Failed to load device plugin", err)
		return nil, errors.Wrap(err, "plugin init failed")
	}

	deviceCtor := &wrapperConstruct{
		DeviceType:        ctor.DeviceType,
		DeviceInterface:   i,
		IsRootDevice:      true,
		DeviceConfigName:  ctor.ConfigName,
		DeviceProvider:    ctor.DeviceName,
		DeviceState:       deviceState,
		LoadData:          loadData,
		Logger:            log,
		SystemLogger:      ctor.Settings.PluginLogger(),
		Secret:            ctor.Settings.Secrets(),
		WorkerID:          ctor.Settings.NodeID(),
		Validator:         ctor.Settings.Validator(),
		DiscoveryChan:     ctor.DiscoveryChan,
		StatusUpdatesChan: ctor.StatusUpdatesChan,
		UOM:               ctor.UOM,
		processor:         newDeviceProcessor(ctor.DeviceType, ctor.RawConfig),
		RawConfig:         ctor.RawConfig,
	}

	wrappers[0] = NewDeviceWrapper(deviceCtor)

	return wrappers, nil
}

// Returns device interface, depends on it's type.
func getExpectedType(deviceType enums.DeviceType) (reflect.Type, error) {
	switch deviceType {
	case enums.DevLight:
		return device.TypeLight, nil
	case enums.DevSwitch:
		return device.TypeSwitch, nil
	case enums.DevSensor:
		return device.TypeSensor, nil
	case enums.DevWeather:
		return device.TypeWeather, nil
	case enums.DevVacuum:
		return device.TypeVacuum, nil
	case enums.DevCamera:
		return device.TypeCamera, nil
	case enums.DevLock:
		return device.TypeLock, nil
	}

	return nil, &ErrUnknownDeviceType{}
}

// Executes load method of the device implementation.
func loadDevice(deviceInterface interface{}, deviceType enums.DeviceType) (interface{}, error) {
	switch deviceType {
	case enums.DevLight:
		return deviceInterface.(device.ILight).Load()
	case enums.DevSwitch:
		return deviceInterface.(device.ISwitch).Load()
	case enums.DevSensor:
		return deviceInterface.(device.ISensor).Load()
	case enums.DevWeather:
		return deviceInterface.(device.IWeather).Load()
	case enums.DevVacuum:
		return deviceInterface.(device.IVacuum).Load()
	case enums.DevCamera:
		return deviceInterface.(device.ICamera).Load()
	case enums.DevLock:
		return deviceInterface.(device.ILock).Load()
	}

	return nil, &ErrUnknownDeviceType{}
}
