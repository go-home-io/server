package storage

import (
	"sync"

	"github.com/gobwas/glob"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/storage"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/systems/logger"
	"gopkg.in/yaml.v2"
)

// Storage provider.
type provider struct {
	sync.Mutex
	plugin   storage.IStorage
	logger   common.ILoggerProvider
	settings *settings
}

// Provider settings.
type settings struct {
	StoreHeartbeat bool     `yaml:"storeHeartbeat"`
	Exclude        []string `yaml:"exclude"`
	Include        []string `yaml:"include"`

	excludeExp []glob.Glob
	includeExp []glob.Glob
}

// ConstructStorage has data required for a new storage provider.
type ConstructStorage struct {
	PluginLogger common.ILoggerProvider
	Secret       common.ISecretProvider
	Loader       providers.IPluginLoaderProvider
	RawConfig    []byte
	Provider     string
}

// NewEmptyStorageProvider returns an empty storage provider.
// It is used if no configuration was supplied.
func NewEmptyStorageProvider() providers.IStorageProvider {
	return &provider{settings: &settings{}}
}

// NewStorageProvider returns a new storage provider.
func NewStorageProvider(ctor *ConstructStorage) providers.IStorageProvider {
	loggerCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.PluginLogger,
		Provider:     ctor.Provider,
		System:       systems.SysStorage.String(),
	}
	log := logger.NewPluginLogger(loggerCtor)
	log.Debug("Loading storage provider")

	settings := &settings{}
	err := yaml.Unmarshal(ctor.RawConfig, settings)
	if err != nil {
		log.Error("Failed to read settings, will not store heartbeat", err)
	}

	settings.excludeExp = make([]glob.Glob, 0)
	settings.includeExp = make([]glob.Glob, 0)

	for _, v := range settings.Exclude {
		g, err := glob.Compile(v)
		if err != nil {
			log.Warn("Failed to compile exclude expression")
			continue
		}
		settings.excludeExp = append(settings.excludeExp, g)
	}

	for _, v := range settings.Include {
		g, err := glob.Compile(v)
		if err != nil {
			log.Warn("Failed to compile include expression")
			continue
		}
		settings.includeExp = append(settings.excludeExp, g) // nolint: gocritic
	}

	pluginLoadRequest := &providers.PluginLoadRequest{
		InitData: &storage.InitDataStorage{
			Logger: log,
			Secret: ctor.Secret,
		},
		RawConfig:      ctor.RawConfig,
		PluginProvider: ctor.Provider,
		SystemType:     systems.SysStorage,
		ExpectedType:   storage.TypeStorage,
	}

	prov := &provider{
		logger:   log,
		settings: settings,
	}

	pluginInterface, err := ctor.Loader.LoadPlugin(pluginLoadRequest)
	if err != nil {
		log.Error("Failed to load storage provider. No data will be persisted", err)
		return prov
	}

	prov.plugin = pluginInterface.(storage.IStorage)
	return prov
}

// State stores a new state entry.
func (s *provider) State(msg *common.MsgDeviceUpdate) {
	go s.processDeviceUpdate(msg)
}

// Heartbeat stores a new heartbeat entry if configured.
func (s *provider) Heartbeat(deviceID string) {
	if !s.settings.StoreHeartbeat {
		return
	}

	go s.processHeartbeat(deviceID)
}

// History returns device state history for the past 24 hrs.
func (s *provider) History(deviceID string) map[enums.Property]map[int64]interface{} {
	s.Lock()
	defer s.Unlock()
	if nil == s.plugin {
		return nil
	}

	d := s.plugin.History(deviceID, 24)
	result := make(map[enums.Property]map[int64]interface{})
	if nil == d {
		return result
	}

	for k, v := range d {
		prop, err := enums.PropertyString(k)
		if err != nil {
			s.logger.Warn("Unknown device property", common.LogIDToken, deviceID,
				common.LogDevicePropertyToken, k)
			continue
		}

		result[prop] = make(map[int64]interface{})
		for t, val := range v {
			f, err := PropertyLoad(prop, val)
			if err != nil {
				s.logger.Error("Failed to unmarshal device property", err, common.LogIDToken, deviceID,
					common.LogDevicePropertyToken, k)
				continue
			}

			result[prop][t] = f
		}
	}

	return result
}

// Processes device update message.
func (s *provider) processDeviceUpdate(msg *common.MsgDeviceUpdate) {
	s.Lock()
	defer s.Unlock()
	if nil == s.plugin || !s.needToSave(msg.Type, msg.ID) {
		return
	}

	data := make(map[string]interface{})

	for k, v := range msg.State {
		d, err := PropertySave(k, v)

		if err != nil {
			s.logger.Error("Failed to prepare property before state saving", err,
				common.LogIDToken, msg.ID, common.LogDevicePropertyToken, k.String())
			continue
		}

		if d == nil {
			continue
		}

		data[k.String()] = d
	}

	s.plugin.State(msg.ID, data)
}

// Processes device heartbeat event.
func (s *provider) processHeartbeat(deviceID string) {
	s.Lock()
	defer s.Unlock()

	if nil == s.plugin {
		return
	}

	s.plugin.Heartbeat(deviceID)
}

func (s *provider) needToSave(deviceType enums.DeviceType, deviceID string) bool {
	switch deviceType {
	case enums.DevCamera:
		return s.needToSaveExcluded(deviceID)
	default:
		return s.needToSaveIncluded(deviceID)
	}
}

func (s *provider) needToSaveExcluded(deviceID string) bool {
	for _, v := range s.settings.includeExp {
		if v.Match(deviceID) {
			return true
		}
	}

	return false
}

func (s *provider) needToSaveIncluded(deviceID string) bool {
	for _, v := range s.settings.excludeExp {
		if v.Match(deviceID) {
			return false
		}
	}

	return true
}
