package server

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/settings"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/utils"
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

// Known devices, received from workers.
type knownDevice struct {
	ID       string                 `json:"id"`
	Worker   string                 `json:"worker"`
	Type     enums.DeviceType       `json:"type"`
	State    map[string]interface{} `json:"state"`
	LastSeen int64                  `json:"last_seen"`
	Commands []string               `json:"commands"`
}

// Connected workers' state.
type serverState struct {
	Settings providers.ISettingsProvider
	Logger   common.ILoggerProvider

	KnownWorkers map[string]*knownWorker
	KnownDevices map[string]*knownDevice

	workerMutex *sync.Mutex
	deviceMutex *sync.Mutex
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
		go s.reBalance()
	}
}

// Update processes incoming device update message.
func (s *serverState) Update(msg *bus.DeviceUpdateMessage) {
	s.Logger.Debug("Received update for the device", common.LogDeviceTypeToken, msg.DeviceType.String(),
		common.LogSystemToken, logSystem, common.LogDeviceNameToken, msg.DeviceID)

	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()

	var dv *knownDevice
	if d, ok := s.KnownDevices[msg.DeviceID]; ok {
		dv = d
	} else {
		dv = &knownDevice{
			Type:     msg.DeviceType,
			Commands: make([]string, len(msg.Commands)),
			State:    make(map[string]interface{}),
			ID:       msg.DeviceID,
		}

		copy(dv.Commands, msg.Commands)
		s.KnownDevices[msg.DeviceID] = dv
	}

	dv.LastSeen = utils.TimeNow()
	dv.Worker = msg.WorkerID
	for k, v := range msg.State {
		dv.State[k] = v
	}
}

// GetAllDevices returns list of all known devices.
func (s *serverState) GetAllDevices() []*knownDevice {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()

	allowedDevices := make([]*knownDevice, 0)
	for _, v := range s.KnownDevices {
		allowedDevices = append(allowedDevices, v)
	}
	return allowedDevices
}

// GetDevice returns device bu ID.
func (s *serverState) GetDevice(deviceID string) *knownDevice {
	s.deviceMutex.Lock()
	defer s.deviceMutex.Unlock()
	return s.KnownDevices[deviceID]
}

// Compares received properties to already known state.
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
func (s *serverState) reBalance() {
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
		candidates := s.pickWorker(&d)
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
		})
	}

	for n, d := range distributed {
		s.Settings.ServiceBus().PublishToWorker(n, bus.NewDeviceAssignmentMessage(d))
		s.KnownWorkers[n].Devices = make([]*bus.DeviceAssignment, len(d))
		copy(s.KnownWorkers[n].Devices, d)
	}

	s.Logger.Debug("Finished re-balancing", common.LogSystemToken, logSystem)
}

// Selecting workers for a device.
func (s *serverState) pickWorker(device *providers.RawDevice) []string {
	nonMet := make([]string, 0)
	for selK, sel := range device.Selector.Selectors {
		key := strings.ToLower(selK)
		r, err := regexp.Compile(sel)
		if err != nil {
			s.Logger.Warn("Device selector misconfiguration",
				common.LogSystemToken, logSystem, common.LogDeviceTypeToken, device.Plugin, "selector", sel)
			continue
		}

		for _, wk := range s.KnownWorkers {
			if utils.SliceContainsString(nonMet, wk.ID) {
				continue
			}

			var met = false
			for k, v := range wk.WorkerProperties {
				if key != k {
					continue
				}
				met = r.MatchString(v)
				break
			}

			if !met {
				nonMet = append(nonMet, wk.ID)
			}
		}
	}

	candidates := make([]string, 0)
	for _, wk := range s.KnownWorkers {
		if !utils.SliceContainsString(nonMet, wk.ID) {
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
		s.reBalance()
	}
}
