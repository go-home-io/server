package device

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
)

// Loads hub device.
// Hub is different from other devices, since it can operate multiple different devices.
func loadHub(ctor *ConstructDevice) ([]IDeviceWrapperProvider, error) {
	wrappers := make([]IDeviceWrapperProvider, 0)
	pluginLogger := ctor.Settings.PluginLogger(systems.SysDevice, ctor.DeviceName)

	loadData := &device.InitDataDevice{
		Logger:                pluginLogger,
		Secret:                ctor.Settings.Secrets(),
		DeviceDiscoveredChan:  make(chan *device.DiscoveredDevices, 3),
		DeviceStateUpdateChan: make(chan *device.StateUpdateData, 10),
	}

	pluginLoadRequest := &providers.PluginLoadRequest{
		InitData:       loadData,
		RawConfig:      []byte(ctor.RawConfig),
		PluginProvider: ctor.DeviceName,
		SystemType:     systems.SysDevice,
		ExpectedType:   device.TypeHub,
	}
	i, err := ctor.Settings.PluginLoader().LoadPlugin(pluginLoadRequest)
	if err != nil {
		pluginLogger.Error("Failed to load hub plugin", err)
		return nil, err
	}

	hub := i.(device.IHub)

	hubResults, err := hub.Load()
	if err != nil {
		pluginLogger.Error("Failed to load hub devices", err)
		return nil, err
	}

	hubCtor := &wrapperConstruct{
		DeviceType:        enums.DevHub,
		DeviceInterface:   hub,
		IsRootDevice:      true,
		Cron:              ctor.Settings.Cron(),
		DeviceConfigName:  ctor.ConfigName,
		DeviceState:       hubResults.State,
		LoadData:          loadData,
		Logger:            pluginLogger,
		Secret:            ctor.Settings.Secrets(),
		WorkerID:          ctor.Settings.NodeID(),
		Validator:         ctor.Settings.Validator(),
		DiscoveryChan:     ctor.DiscoveryChan,
		StatusUpdatesChan: ctor.StatusUpdatesChan,
	}

	hubWrapper := NewDeviceWrapper(hubCtor)
	wrappers = append(wrappers, hubWrapper)

	for _, v := range hubResults.Devices {
		subLoadData := &device.InitDataDevice{
			Logger:                pluginLogger,
			Secret:                ctor.Settings.Secrets(),
			DeviceDiscoveredChan:  loadData.DeviceDiscoveredChan,
			DeviceStateUpdateChan: make(chan *device.StateUpdateData, 10),
		}

		dev, ok := v.Interface.(device.IDevice)
		if !ok {
			pluginLogger.Warn("One of the loaded devices is not implementing IDevice interface",
				common.LogDeviceNameToken, hubWrapper.GetID())
			continue
		}

		dev.Init(subLoadData)

		spawnedCtor := &wrapperConstruct{
			DeviceType:        v.Type,
			DeviceInterface:   v.Interface,
			IsRootDevice:      false,
			Cron:              ctor.Settings.Cron(),
			DeviceConfigName:  ctor.ConfigName,
			DeviceState:       v.State,
			LoadData:          subLoadData,
			Logger:            pluginLogger,
			Secret:            ctor.Settings.Secrets(),
			WorkerID:          ctor.Settings.NodeID(),
			Validator:         ctor.Settings.Validator(),
			DiscoveryChan:     ctor.DiscoveryChan,
			StatusUpdatesChan: ctor.StatusUpdatesChan,
		}

		w := NewDeviceWrapper(spawnedCtor)
		if nil == w {
			continue
		}
		wrappers = append(wrappers, w)
	}

	return wrappers, nil
}
