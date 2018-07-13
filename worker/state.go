package worker

import (
	"sync"

	busPlugin "github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/systems/device"
)

// IWorkerStateProvider state abstraction.
type IWorkerStateProvider interface {
	// Processing device assignment message.
	DevicesAssignmentMessage(*bus.DeviceAssignmentMessage)
	// Processing device command message.
	DevicesCommandMessage(*bus.DeviceCommandMessage)
}

// Worker state definition.
type workerState struct {
	Settings providers.ISettingsProvider
	Logger   common.ILoggerProvider

	mutex *sync.Mutex

	devices map[string]device.IDeviceWrapperProvider

	statusUpdatesChan chan *device.UpdateEvent
	discoveryChan     chan *device.NewDeviceDiscoveredEvent
}

// Creating a new worker state object.
func newWorkerState(settings providers.ISettingsProvider) *workerState {
	w := workerState{
		Settings: settings,
		Logger:   settings.SystemLogger(),

		mutex: &sync.Mutex{},

		devices:           make(map[string]device.IDeviceWrapperProvider),
		discoveryChan:     make(chan *device.NewDeviceDiscoveredEvent, 5),
		statusUpdatesChan: make(chan *device.UpdateEvent, 30),
	}

	go w.start()
	return &w
}

// DevicesAssignmentMessage processes a device assignment message, received from server.
// Worker should stop existing device-listening processes and re-load a new set.
func (w *workerState) DevicesAssignmentMessage(msg *bus.DeviceAssignmentMessage) {
	w.Logger.Info("Received device assignment message", common.LogSystemToken, logSystem)
	go w.loadDevices(msg)
}

// DevicesCommandMessage processes a new device command message, received from server.
func (w *workerState) DevicesCommandMessage(msg *bus.DeviceCommandMessage) {
	w.Logger.Debug("Received device command message", common.LogSystemToken, logSystem,
		common.LogDeviceNameToken, msg.DeviceID, common.LogDeviceCommandToken, msg.Command.String())

	wrapper, ok := w.devices[msg.DeviceID]
	if !ok {
		w.Logger.Warn("Failed to find device on this worker", common.LogSystemToken, logSystem,
			common.LogDeviceNameToken, msg.DeviceID,
			common.LogDeviceCommandToken, msg.Command.String())
		return
	}

	wrapper.InvokeCommand(msg.Command, msg.Payload)
}

// Starting worker state internal processes.
func (w *workerState) start() {
	for {
		select {
		case update := <-w.statusUpdatesChan:
			wrapper, ok := w.devices[update.ID]
			if !ok {
				w.Logger.Warn("Received unknown device update", common.LogSystemToken, logSystem,
					common.LogDeviceNameToken, update.ID)
				break
			}

			go w.Settings.ServiceBus().Publish(busPlugin.ChDeviceUpdates, wrapper.GetUpdateMessage())
		case discover := <-w.discoveryChan:
			id := discover.Provider.GetID()
			if _, ok := w.devices[id]; ok {
				w.Logger.Warn("Received duplicate discovery for the same device",
					common.LogSystemToken, logSystem, common.LogDeviceNameToken, id)
			}

			w.devices[id] = discover.Provider
		}
	}
}

// Loading a new set of devices.
func (w *workerState) loadDevices(msg *bus.DeviceAssignmentMessage) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.unloadDevices()

	wg := sync.WaitGroup{}
	wg.Add(len(msg.Devices))

	for _, a := range msg.Devices {
		ctor := &device.ConstructDevice{
			DiscoveryChan:     w.discoveryChan,
			StatusUpdatesChan: w.statusUpdatesChan,
			Settings:          w.Settings,
			ConfigName:        a.Name,
			RawConfig:         a.Config,
			DeviceName:        a.Plugin,
			DeviceType:        a.Type,
		}

		go func() {
			defer wg.Done()
			wrappers, err := device.LoadDevice(ctor)
			if err != nil {
				return
			}

			for _, v := range wrappers {
				w.devices[v.GetID()] = v
				go w.Settings.ServiceBus().Publish(busPlugin.ChDeviceUpdates, v.GetUpdateMessage())
			}
		}()
	}

	wg.Wait()

	w.Logger.Info("Done re-loading devices", common.LogSystemToken, logSystem)
}

// Unloading an old set of devices.
func (w *workerState) unloadDevices() {
	w.Logger.Debug("Unloading devices", common.LogSystemToken, logSystem)
	for _, v := range w.devices {
		v.Unload()
	}
}
