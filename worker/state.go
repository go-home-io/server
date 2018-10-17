package worker

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"

	busPlugin "github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems/api"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/systems/device"
	"github.com/go-home-io/server/utils"
)

// IWorkerStateProvider state abstraction.
type IWorkerStateProvider interface {
	// Processing device assignment message.
	DevicesAssignmentMessage(*bus.DeviceAssignmentMessage)
	// Processing device command message.
	DevicesCommandMessage(*bus.DeviceCommandMessage)
}

var (
	// Timeout before terminating device load.
	deviceLoadTimeout = 50 * time.Second
	// Timeout before unload fail panic.
	deviceUnloadTimeout = 5 * time.Second
)

// Worker state definition.
type workerState struct {
	Settings providers.ISettingsProvider
	Logger   common.ILoggerProvider

	mutex     *sync.Mutex
	dictMutex *sync.Mutex

	devices      map[string]device.IDeviceWrapperProvider
	extendedAPIs map[string]providers.IExtendedAPIProvider

	statusUpdatesChan chan *device.UpdateEvent
	discoveryChan     chan *device.NewDeviceDiscoveredEvent

	lastAssignment     []string
	lastAssignmentTime int64
	failedDevices      *bus.DeviceAssignmentMessage
}

// Creating a new worker state object.
func newWorkerState(settings providers.ISettingsProvider) *workerState {
	w := workerState{
		Settings: settings,
		Logger:   settings.SystemLogger(),

		mutex:     &sync.Mutex{},
		dictMutex: &sync.Mutex{},

		lastAssignment: make([]string, 0),

		devices:           make(map[string]device.IDeviceWrapperProvider),
		extendedAPIs:      make(map[string]providers.IExtendedAPIProvider),
		discoveryChan:     make(chan *device.NewDeviceDiscoveredEvent, 5),
		statusUpdatesChan: make(chan *device.UpdateEvent, 30),
	}

	_, err := settings.Cron().AddFunc("@every 15s", w.checkStaleMaster)
	if err != nil {
		w.Logger.Fatal("Failed to start staled workers job", err)
	}

	go w.start()
	return &w
}

// DevicesAssignmentMessage processes a device assignment message, received from server.
// Worker should stop existing device-listening processes and re-load a new set.
func (w *workerState) DevicesAssignmentMessage(msg *bus.DeviceAssignmentMessage) {
	w.lastAssignmentTime = utils.TimeNow()
	w.mutex.Lock()
	tmpSum := make([]string, 0)
	for _, v := range msg.Devices {
		t := md5.Sum([]byte(v.Config)) // nolint: gosec
		tmpSum = append(tmpSum, hex.EncodeToString(t[:]))
	}

	if helpers.SliceEqualsString(w.lastAssignment, tmpSum) {
		w.mutex.Unlock()
		w.Logger.Debug("Received device assignment is the same", common.LogSystemToken, logSystem)
		return
	}

	w.lastAssignment = make([]string, len(tmpSum))
	copy(w.lastAssignment, tmpSum)
	w.mutex.Unlock()

	w.Logger.Info("Received device assignment message", common.LogSystemToken, logSystem)
	w.unloadDevices()
	go w.loadDevices(msg)
}

// DevicesCommandMessage processes a new device command message, received from server.
func (w *workerState) DevicesCommandMessage(msg *bus.DeviceCommandMessage) {
	w.Logger.Debug("Received device command message", common.LogSystemToken, logSystem,
		common.LogIDToken, msg.DeviceID, common.LogDeviceCommandToken, msg.Command.String())

	wrapper, ok := w.devices[msg.DeviceID]
	if !ok {
		w.Logger.Warn("Failed to find device on this worker", common.LogSystemToken, logSystem,
			common.LogIDToken, msg.DeviceID,
			common.LogDeviceCommandToken, msg.Command.String())
		return
	}

	wrapper.InvokeCommand(msg.Command, msg.Payload)
}

// Periodic checks to determine whether master is active.
func (w *workerState) checkStaleMaster() {
	w.mutex.Lock()
	if 0 != len(w.lastAssignment) && utils.IsLongTimeNoSee(w.lastAssignmentTime) {
		w.Logger.Warn("Didn't get anything from master for a long time. Unloading devices",
			common.LogSystemToken, logSystem)
		w.lastAssignment = make([]string, 0)
		go w.unloadDevices()
	}

	w.mutex.Unlock()
}

// Starting worker state internal processes.
func (w *workerState) start() {
	// We generally don't want to overlap with discovery messages.
	// Not that we care much, but discovery can bring a new assignment message
	// so it doesn't make much sense to re-try at the same time.
	_, err := w.Settings.Cron().AddFunc("@every 1m13s", w.retryLoad)
	if err != nil {
		w.Logger.Error("Failed to register retry cron", err, common.LogSystemToken, logSystem)
	}

	for {
		select {
		case update := <-w.statusUpdatesChan:
			w.mutex.Lock()
			wrapper, ok := w.devices[update.ID]
			w.mutex.Unlock()
			if !ok {
				w.Logger.Warn("Received unknown device update", common.LogSystemToken, logSystem,
					common.LogIDToken, update.ID)
				break
			}

			w.mutex.Lock()
			go w.Settings.ServiceBus().Publish(busPlugin.ChDeviceUpdates, wrapper.GetUpdateMessage())
			w.mutex.Unlock()
		case discover := <-w.discoveryChan:
			w.mutex.Lock()
			id := discover.Provider.ID()
			if _, ok := w.devices[id]; ok {
				w.Logger.Warn("Received duplicate discovery for the same device",
					common.LogSystemToken, logSystem, common.LogIDToken, id)
			}

			w.devices[id] = discover.Provider
			w.mutex.Unlock()
		}
	}
}

// Loading a new set of devices.
func (w *workerState) loadDevices(msg *bus.DeviceAssignmentMessage) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if nil == msg {
		return
	}

	wg := sync.WaitGroup{}

	devices := make([]*bus.DeviceAssignment, len(msg.Devices))
	copy(devices, msg.Devices)

	wg.Add(len(devices))

	failed := &bus.DeviceAssignmentMessage{
		Devices: make([]*bus.DeviceAssignment, 0),
	}

	for _, a := range devices {
		a.LoadFinished = false
		a.CancelLoading = false

		if a.IsAPI {
			ctor := &api.ConstructAPI{
				Name:       a.Name,
				Provider:   a.Plugin,
				IsServer:   false,
				Logger:     w.Settings.PluginLogger(),
				Loader:     w.Settings.PluginLoader(),
				RawConfig:  []byte(a.Config),
				ServiceBus: w.Settings.ServiceBus(),
				Secret:     w.Settings.Secrets(),
				Validator:  w.Settings.Validator(),
			}

			go w.tryAPIAssignmentLoad(a, ctor, &wg, failed)
			continue
		}

		ctor := &device.ConstructDevice{
			DiscoveryChan:     w.discoveryChan,
			StatusUpdatesChan: w.statusUpdatesChan,
			Settings:          w.Settings,
			ConfigName:        a.Name,
			RawConfig:         a.Config,
			DeviceName:        a.Plugin,
			DeviceType:        a.Type,
			UOM:               msg.UOM,
		}

		go w.tryDeviceAssignmentLoad(a, ctor, &wg, failed)
	}

	if !waitWithTimeout(&wg, deviceLoadTimeout) {
		w.Logger.Warn("Got timeout while waiting for devices load")
		w.lastAssignment = make([]string, 0)
		analyzeFailedToLoadDevices(devices, failed)
	}

	if len(failed.Devices) > 0 {
		w.failedDevices = failed
		w.Logger.Warn("Failed to reload some device, will retry later", common.LogSystemToken, logSystem)
	} else {
		w.failedDevices = nil
	}

	w.Logger.Info("Done re-loading devices", common.LogSystemToken, logSystem)
}

// Tries to load API providers.
func (w *workerState) tryAPIAssignmentLoad(a *bus.DeviceAssignment, ctor *api.ConstructAPI,
	wg *sync.WaitGroup, failed *bus.DeviceAssignmentMessage) {
	defer wg.Done()

	wrapper, err := api.NewExtendedAPIProvider(ctor)
	a.LoadFinished = true

	if err == nil && a.CancelLoading {
		go w.tryUnload(wrapper)
		return
	}

	if a.CancelLoading {
		return
	}

	if err != nil {
		failed.Devices = append(failed.Devices, a)
		return
	}

	w.dictMutex.Lock()
	defer w.dictMutex.Unlock()
	_, ok := w.extendedAPIs[wrapper.ID()]
	if ok {
		w.Logger.Warn("Duplicated load for API, unloading", common.LogSystemToken, logSystem)
		go w.tryUnload(wrapper)
		return
	}

	w.extendedAPIs[wrapper.ID()] = wrapper
}

// Tries to unload provider.
func (w *workerState) tryUnload(p ...providers.ILoadedProvider) {
	wg := sync.WaitGroup{}
	wg.Add(len(p))
	for _, v := range p {
		go func(w *sync.WaitGroup, pr providers.ILoadedProvider) {
			defer wg.Done()
			v.Unload()
		}(&wg, v)
	}

	if !waitWithTimeout(&wg, deviceUnloadTimeout) {
		w.Logger.Fatal("Failed to unload provider, have to terminate", &ErrUnloadFailed{})
	}
}

// Tries to load device providers.
func (w *workerState) tryDeviceAssignmentLoad(a *bus.DeviceAssignment, ctor *device.ConstructDevice,
	wg *sync.WaitGroup, failed *bus.DeviceAssignmentMessage) {
	defer wg.Done()
	wrappers, err := device.LoadDevice(ctor)
	a.LoadFinished = true

	if err == nil && a.CancelLoading {
		for _, v := range wrappers {
			go w.tryUnload(v)
		}
	}

	if a.CancelLoading {
		return
	}

	if err != nil {
		failed.Devices = append(failed.Devices, a)
		return
	}

	alreadyPresent := false

	w.dictMutex.Lock()
	defer w.dictMutex.Unlock()
	for _, v := range wrappers {
		_, ok := w.devices[v.ID()]
		if ok {
			alreadyPresent = true
			break
		}
	}

	if alreadyPresent {
		w.Logger.Warn("Duplicated load for devices, unloading", common.LogSystemToken, logSystem)
		for _, v := range wrappers {
			go w.tryUnload(v)
		}
		return
	}

	for _, v := range wrappers {
		w.devices[v.ID()] = v
		go w.Settings.ServiceBus().Publish(busPlugin.ChDeviceUpdates, v.GetUpdateMessage())
	}
}

// Analyzes failed to load due to timeout devices.
func analyzeFailedToLoadDevices(devices []*bus.DeviceAssignment, failed *bus.DeviceAssignmentMessage) {
	for _, a := range devices {
		if !a.LoadFinished {
			failed.Devices = append(failed.Devices, a)
			a.CancelLoading = true
		}
	}
}

// Waits with timeout.
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}

// Retries loading failed devices.
func (w *workerState) retryLoad() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if nil != w.failedDevices && len(w.failedDevices.Devices) > 0 {
		go w.loadDevices(w.failedDevices)
	}
}

// Unloading an old set of devices.
func (w *workerState) unloadDevices() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.dictMutex.Lock()
	defer w.dictMutex.Unlock()

	w.Logger.Debug("Unloading devices", common.LogSystemToken, logSystem)
	w.failedDevices = nil
	for k, v := range w.devices {
		go w.tryUnload(v)
		delete(w.devices, k)
	}

	for k, v := range w.extendedAPIs {
		go w.tryUnload(v)
		delete(w.extendedAPIs, k)
	}

	w.Logger.Debug("Done un-loading", common.LogSystemToken, logSystem)
}
