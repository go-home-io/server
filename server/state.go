package server

import (
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/settings"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/utils"
	"github.com/gobwas/glob"
)

// IServerStateProvider defines server state logic.
type IServerStateProvider interface {
	Discovery(msg *bus.DiscoveryMessage)
	Update(msg *bus.DeviceUpdateMessage)
	GetAllDevices() []*knownDevice
	GetDevice(string) *knownDevice
}

// Worker properties.
type knownWorker struct {
	ID               string
	LastSeen         int64
	WorkerProperties map[string]string
	Devices          []*bus.DeviceAssignment
	MaxDevices       int
}

// Connected workers' state.
type serverState struct {
	Settings providers.ISettingsProvider
	Logger   common.ILoggerProvider

	KnownWorkers map[string]*knownWorker
	KnownDevices map[string]*knownDevice

	workerMutex *sync.Mutex
	deviceMutex *sync.Mutex

	fanOut providers.IInternalFanOutProvider
}

// Constructs a new server state.
func newServerState(settings providers.ISettingsProvider) *serverState {
	s := serverState{
		KnownWorkers: make(map[string]*knownWorker),
		KnownDevices: make(map[string]*knownDevice),
		Settings:     settings,
		Logger:       settings.SystemLogger(),

		workerMutex: &sync.Mutex{},
		deviceMutex: &sync.Mutex{},

		fanOut: settings.FanOud(),
	}

	_, err := settings.Cron().AddFunc("@every 15s", s.checkStaleWorkers)
	if err != nil {
		panic("Failed to start staled workers job")
	}
	return &s
}

// Discovery processes incoming discovery message.
func (s *serverState) Discovery(msg *bus.DiscoveryMessage) {
	var wk *knownWorker
	var reBalanceNeeded bool
	var newWorkerID string
	syncProperties := true

	if w, ok := s.KnownWorkers[msg.NodeID]; ok {
		wk = w
		if s.compareProperties(msg) {
			if msg.IsFirstStart {
				s.Logger.Info("Received discovery from a known worker with no changes, re-sending device data",
					common.LogWorkerToken, msg.NodeID, common.LogSystemToken, logSystem)
				s.Settings.ServiceBus().PublishToWorker(msg.NodeID, bus.NewDeviceAssignmentMessage(wk.Devices))
			} else {
				s.Logger.Debug("Received discovery from a known worker",
					common.LogWorkerToken, msg.NodeID, common.LogSystemToken, logSystem)
			}
			syncProperties = false
			reBalanceNeeded = false

		} else {
			s.Logger.Info("Received discovery from a known worker with changes in properties, re-balance needed",
				common.LogWorkerToken, msg.NodeID, common.LogSystemToken, logSystem)
			reBalanceNeeded = true
		}

	} else {
		s.Logger.Info("Received discovery from a new worker",
			common.LogWorkerToken, msg.NodeID, common.LogSystemToken, logSystem)
		wk = &knownWorker{
			ID:      msg.NodeID,
			Devices: make([]*bus.DeviceAssignment, 0),
		}

		s.KnownWorkers[msg.NodeID] = wk
		reBalanceNeeded = true
		newWorkerID = wk.ID
	}
	wk.LastSeen = utils.TimeNow()
	wk.MaxDevices = msg.MaxDevices

	if syncProperties {
		wk.WorkerProperties = make(map[string]string, len(msg.Properties)+1)
		for k, v := range msg.Properties {
			wk.WorkerProperties[strings.ToLower(k)] = v
		}
		wk.WorkerProperties[settings.ConfigSelectorName] = msg.NodeID
	}

	if reBalanceNeeded {
		go s.reBalance(newWorkerID)
	}
}

// Update processes incoming device update message.
func (s *serverState) Update(msg *bus.DeviceUpdateMessage) {
	s.Logger.Debug("Received update for the device", common.LogDeviceTypeToken, msg.DeviceType.String(),
		common.LogSystemToken, logSystem, common.LogDeviceNameToken, msg.DeviceID)
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()
	firstOccurrence := false

	var dv *knownDevice
	if d, ok := s.KnownDevices[msg.DeviceID]; ok {
		dv = d
	} else {
		firstOccurrence = true
		dv = &knownDevice{
			Type:     msg.DeviceType,
			Commands: make([]string, len(msg.Commands)),
			State:    make(map[string]interface{}),
			ID:       msg.DeviceID,
			Name:     msg.DeviceName,
		}

		copy(dv.Commands, msg.Commands)
		s.KnownDevices[msg.DeviceID] = dv
	}

	dv.LastSeen = utils.TimeNow()
	dv.Worker = msg.WorkerID

	s.processDeviceStateUpdate(dv, msg.State, firstOccurrence)
}

// GetAllDevices returns list of all known devices.
func (s *serverState) GetAllDevices() []*knownDevice {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()

	devices := make([]*knownDevice, 0)
	for _, v := range s.KnownDevices {
		devices = append(devices, v)
	}
	return devices
}

// GetDevice returns device bu ID.
func (s *serverState) GetDevice(deviceID string) *knownDevice {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()
	return s.KnownDevices[deviceID]
}

func (s *serverState) processDeviceStateUpdate(dv *knownDevice, newState map[string]interface{}, firstOccurrence bool) {
	msg := &common.MsgDeviceUpdate{
		ID:        dv.ID,
		State:     make(map[enums.Property]interface{}),
		FirstSeen: firstOccurrence,
		Type:      dv.Type,
	}
	for k, v := range newState {
		prop, err := enums.PropertyString(k)
		if err != nil {
			continue
		}

		t, err := helpers.PropertyFixYaml(v, prop)
		if nil != err {
			continue
		}

		if !firstOccurrence {
			prev, ok := dv.State[k]
			if ok && !helpers.PropertyDeepEqual(prev, v, prop) {
				msg.State[prop] = t
			}
		} else {
			msg.State[prop] = t
		}

		dv.State[k] = v
	}

	if 0 != len(msg.State) {
		s.fanOut.ChannelInDeviceUpdates() <- msg
	}
}

// Compares received properties to already known state.
// It's enough to check length and iterate through known worker properties
// since it covers all possible changes.
func (s *serverState) compareProperties(msg *bus.DiscoveryMessage) bool {
	if len(s.KnownWorkers[msg.NodeID].WorkerProperties) != len(msg.Properties)+1 {
		return false
	}

	for k, v := range s.KnownWorkers[msg.NodeID].WorkerProperties {
		if k == settings.ConfigSelectorName {
			continue
		}
		if w, ok := msg.Properties[k]; !ok || v != w {
			return false
		}
	}

	return true
}

// Re-balancing devices between workers.
func (s *serverState) reBalance(newWorkerID string) {
	s.Logger.Debug("Starting re-balancing", common.LogSystemToken, logSystem)
	s.workerMutex.Lock()
	defer s.workerMutex.Unlock()

	sort.Slice(s.Settings.DevicesConfig(), func(i, j int) bool {
		return len(s.Settings.DevicesConfig()[i].Selector.Selectors) > len(s.Settings.DevicesConfig()[j].Selector.Selectors)
	})

	distributed := make(map[string][]*bus.DeviceAssignment)

	for _, wk := range s.KnownWorkers {
		distributed[wk.ID] = make([]*bus.DeviceAssignment, 0)
	}

	for _, d := range s.Settings.DevicesConfig() {
		candidates := s.pickWorker(d)
		if 0 == len(candidates) {
			s.Logger.Warn("Failed to select a worker for the device", common.LogSystemToken, logSystem,
				common.LogDeviceTypeToken, d.Plugin, "name", d.Selector.Name)
			continue
		}

		best := ""
		min := 10000
		for _, c := range candidates {
			assigned := len(distributed[c])
			if assigned < min && assigned < s.KnownWorkers[c].MaxDevices {
				min = len(distributed[c])
				best = c
			}
		}

		if "" == best {
			s.Logger.Warn("Failed to select a worker: too many devices", common.LogSystemToken, logSystem,
				common.LogDeviceTypeToken, d.Plugin, "name", d.Selector.Name)
			continue
		}

		distributed[best] = append(distributed[best], &bus.DeviceAssignment{
			Plugin: d.Plugin,
			Config: d.StrConfig,
			Type:   d.DeviceType,
			Name:   d.Name,
			IsAPI:  d.IsAPI,
		})
	}

	for n, d := range distributed {
		if n != newWorkerID && s.isWorkerHasSameDevicesAlready(n, d) {
			s.Logger.Debug("Worker already has same set of devices, not sending an update",
				"worker", n, common.LogSystemToken, logSystem)
			continue
		}

		if n == newWorkerID && 0 == len(d) {
			continue
		}

		s.Settings.ServiceBus().PublishToWorker(n, bus.NewDeviceAssignmentMessage(d))
		s.KnownWorkers[n].Devices = make([]*bus.DeviceAssignment, len(d))
		copy(s.KnownWorkers[n].Devices, d)
	}

	s.Logger.Debug("Finished re-balancing", common.LogSystemToken, logSystem)
}

// Validates whether worker has all this devices already, so we don't need
// to sed devices assignment message. Helps to avoid unnecessary updated.
func (s *serverState) isWorkerHasSameDevicesAlready(workerID string, proposedDevices []*bus.DeviceAssignment) bool {
	existing, ok := s.KnownWorkers[workerID]
	if !ok {
		return false
	}

	if len(existing.Devices) != len(proposedDevices) {
		return false
	}

	check := func(one []*bus.DeviceAssignment, two []*bus.DeviceAssignment) bool {
		for _, oneX := range one {
			isFound := false
			for _, twoX := range two {
				isFound = reflect.DeepEqual(oneX, twoX)
				if isFound {
					break
				}
			}

			if !isFound {
				return false
			}
		}

		return true
	}

	return check(existing.Devices, proposedDevices) && check(proposedDevices, existing.Devices)
}

// Selecting workers for a device.
func (s *serverState) pickWorker(device *providers.RawDevice) []string {
	nonMet := make([]string, 0)
	for selK, sel := range device.Selector.Selectors {
		key := strings.ToLower(selK)
		r, err := glob.Compile(sel)
		if err != nil {
			s.Logger.Warn("Device selector misconfiguration",
				common.LogSystemToken, logSystem, common.LogDeviceTypeToken, device.Plugin, "selector", sel)
			continue
		}

		for _, wk := range s.KnownWorkers {
			if helpers.SliceContainsString(nonMet, wk.ID) {
				continue
			}

			var met = false
			for k, v := range wk.WorkerProperties {
				if key != k {
					continue
				}
				met = r.Match(v)
				break
			}

			if !met {
				nonMet = append(nonMet, wk.ID)
			}
		}
	}

	candidates := make([]string, 0)
	for _, wk := range s.KnownWorkers {
		if !helpers.SliceContainsString(nonMet, wk.ID) {
			candidates = append(candidates, wk.ID)
		}
	}

	return candidates
}

// Periodic validation whether all workers are sending pings in time.
// Triggers re-balance if workers didn't send anything for the past 2 minutes.
func (s *serverState) checkStaleWorkers() {
	s.workerMutex.Lock()

	now := utils.TimeNow()
	toDelete := make([]string, 0)

	for name, v := range s.KnownWorkers {
		if now-v.LastSeen > 2*60 {
			toDelete = append(toDelete, name)
		}
	}

	for _, name := range toDelete {
		s.Logger.Info("Removing stale worker", common.LogSystemToken, logSystem,
			common.LogWorkerToken, name)
		delete(s.KnownWorkers, name)
	}

	s.workerMutex.Unlock()
	if len(toDelete) > 0 {
		s.reBalance("")
	}
}
