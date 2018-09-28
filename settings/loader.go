// Package settings is responsible for parsing yaml-based configuration.
package settings

import (
	"bytes"
	"io"
	"strings"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
	"github.com/go-home-io/server/systems/bus"
	"github.com/go-home-io/server/systems/config"
	"github.com/go-home-io/server/systems/fanout"
	"github.com/go-home-io/server/systems/logger"
	"github.com/go-home-io/server/systems/secret"
	"github.com/go-home-io/server/systems/security"
	"github.com/go-home-io/server/systems/storage"
	"github.com/go-home-io/server/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	// Logger system.
	logSystem = "settings"
)

const (
	// Describes config record for server.
	configGoHomeMaster = "master"
	// Describes config record for worker.
	configGoHomeWorker = "worker"
	// ConfigSelectorName describes selector name field.
	ConfigSelectorName = "name"
)

// StartUpOptions defines arguments allowed by the system.
type StartUpOptions struct {
	PluginsFolder string `short:"p" long:"plugins" description:"Plugins location."`
	IsWorker      bool   `short:"w" long:"worker" description:"Flag indicating working instance."`

	Config map[string]string `short:"c" long:"config" description:"Config files provider. Defaults to local FS."`
	Secret map[string]string `short:"s" long:"secret" description:"Secrets provider. Defaults to local FS."`
}

// Defines loaded provider record.
type rawProvider struct {
	System   string
	Provider string
	Config   []byte
}

// System settings.
type settingsProvider struct {
	logger       common.ILoggerProvider
	bus          providers.IBusProvider
	nodeID       string
	cron         providers.ICronProvider
	pluginLoader providers.IPluginLoaderProvider
	validator    providers.IValidatorProvider
	secrets      common.ISecretProvider
	storage      providers.IStorageProvider

	wSettings *providers.WorkerSettings
	mSettings *providers.MasterSettings
	isWorker  bool

	devicesConfig []*providers.RawDevice

	rawRoles         []*providers.SecRole
	rawUsersProvider *rawProvider
	securityProvider providers.ISecurityProvider

	fanOut       providers.IInternalFanOutProvider
	triggers     []*providers.RawMasterComponent
	extendedAPIs []*providers.RawMasterComponent
	groups       []*providers.RawMasterComponent
}

// Load system configuration.
func Load(options *StartUpOptions) providers.ISettingsProvider {
	settings := settingsProvider{
		isWorker:      options.IsWorker,
		devicesConfig: make([]*providers.RawDevice, 0),
		logger:        logger.NewConsoleLogger(),
		rawRoles:      make([]*providers.SecRole, 0),
		triggers:      make([]*providers.RawMasterComponent, 0),
		extendedAPIs:  make([]*providers.RawMasterComponent, 0),
		fanOut:        fanout.NewFanOut(),
		groups:        make([]*providers.RawMasterComponent, 0),
	}

	settings.validator = utils.NewValidator(settings.logger)

	pluginsCtor := &utils.ConstructPluginLoader{
		PluginsFolder: options.PluginsFolder,
		Validator:     settings.validator,
	}
	settings.pluginLoader = utils.NewPluginLoader(pluginsCtor)

	secretsConstruct := &secret.ConstructSecret{
		Logger:  settings.logger,
		Options: options.Secret,
		Loader:  settings.pluginLoader,
	}
	settings.secrets = secret.NewSecretProvider(secretsConstruct)

	tplCtor := &constructTemplate{
		Logger:  settings.logger,
		Secrets: settings.secrets,
	}
	templateProvider := newTemplateProvider(tplCtor)

	allProviders := make([]*rawProvider, 0)

	cfgConstruct := &config.ConstructConfig{
		Logger:  settings.logger,
		Options: options.Config,
		Loader:  settings.pluginLoader,
		Secret:  settings.secrets,
	}
	configProvider := config.NewConfigProvider(cfgConstruct)

	dataChan := configProvider.Load()
	if nil == dataChan {
		settings.logger.Fatal("Didn't get any configuration", errors.New("config provider returned nothing"))
		return nil
	}

	for fileData := range dataChan {
		allProviders = append(allProviders, settings.loadFile(fileData, templateProvider)...)
	}

	allProviders = settings.loadDevicesAndGoHomeDefinitions(allProviders)
	allProviders = settings.loadLoggerProvider(allProviders)

	for _, v := range allProviders {
		settings.parseProvider(v)
	}

	secConstruct := &security.ConstructSecurityProvider{
		Roles:        settings.rawRoles,
		Secret:       settings.secrets,
		Loader:       settings.pluginLoader,
		Logger:       settings.logger,
		UserProvider: "",
	}

	if nil != settings.rawUsersProvider {
		secConstruct.UserProvider = settings.rawUsersProvider.Provider
		secConstruct.UserRawConfig = settings.rawUsersProvider.Config
	}

	settings.securityProvider = security.NewSecurityProvider(secConstruct)

	settings.validate()
	return &settings
}

// Validates whether all necessary settings are present.
func (s *settingsProvider) validate() {
	if s.bus == nil {
		panic("Service bus is not configured")
	}

	s.cron = utils.NewCron()
	_, err := s.cron.AddFunc("@every 10s", func() {
		s.logger.Flush()
	})

	if err != nil {
		panic("Failed to register logger flushing")
	}

	if s.isWorker {
		if nil == s.wSettings {
			s.logger.Warn("Worker settings are not defined, using the default ones",
				common.LogSystemToken, logSystem)
			s.wSettings = &providers.WorkerSettings{
				MaxDevices: 99,
			}
		}
	} else {
		if nil == s.mSettings {
			s.logger.Warn("Master settings are not defined, using the default ones",
				common.LogSystemToken, logSystem)
			s.mSettings = &providers.MasterSettings{
				Port: 8080,
			}
		}

		if nil == s.storage {
			s.logger.Warn("Storage provider is not defined", common.LogSystemToken, logSystem)
			s.storage = storage.NewEmptyStorageProvider()
		}
	}
}

// Processes single yaml file.
func (s *settingsProvider) loadFile(fileData []byte, templateProvider ITemplateProvider) []*rawProvider {

	fileData = templateProvider.Process(fileData)

	provs := make([]*rawProvider, 0)
	decoder := yaml.NewDecoder(bytes.NewReader(fileData))
	for {
		var value map[string]interface{}
		err := decoder.Decode(&value)
		if err == io.EOF {
			break
		}

		if err != nil {
			s.logger.Error("Failed to parse config file", err, common.LogSystemToken, logSystem)
			continue
		}

		componentType := ""
		componentProvider := ""

		if cs, ok := value["system"].(string); ok {
			componentType = strings.ToLower(cs)
		}

		if ct, ok := value["provider"].(string); ok {
			componentProvider = strings.ToLower(ct)
		}

		if componentType == "" || componentProvider == "" {
			s.logger.Warn("Failed to parse a record in the config file: system or provider is not defined",
				common.LogSystemToken, logSystem)
			continue
		}

		byteData, err := yaml.Marshal(value)
		if err != nil {
			s.logger.Error("Failed to parse config file", err, common.LogSystemToken, componentType,
				common.LogProviderToken, componentProvider)
			continue
		}

		provs = append(provs, &rawProvider{
			Provider: strings.ToLower(componentProvider),
			System:   strings.ToLower(componentType),
			Config:   byteData,
		})
	}

	return provs
}

// Loads config for master/worker nodes and raw devices.
func (s *settingsProvider) loadDevicesAndGoHomeDefinitions(provs []*rawProvider) []*rawProvider {
	providersLeft := make([]*rawProvider, 0)

	for _, v := range provs {
		sys, err := systems.SystemTypeString(v.System)
		if err != nil {
			providersLeft = append(providersLeft, v)
			continue
		}

		switch sys {
		case systems.SysGoHome:
			s.loadGoHomeDefinition(v)
		case systems.SysDevice:
			err = s.processDeviceProvider(v)
		default:
			providersLeft = append(providersLeft, v)
		}

		if err != nil {
			s.logger.Error("Failed to load provider config", err, common.LogProviderToken, v.Provider,
				common.LogSystemToken, v.System)
		}
	}

	return providersLeft
}

// Loads worker or server configuration.
func (s *settingsProvider) loadGoHomeDefinition(provider *rawProvider) {
	if s.isWorker && provider.Provider == configGoHomeWorker {
		set := &providers.WorkerSettings{}
		if err := yaml.Unmarshal(provider.Config, &set); err != nil {
			panic("Failed to unmarshal worker config")
		}

		if !s.validator.Validate(set) {
			panic("Incorrect worker settings")
		}

		s.wSettings = set
		s.nodeID = s.wSettings.Name

	} else if !s.isWorker && provider.Provider == configGoHomeMaster {
		set := &providers.MasterSettings{
			Locations: make([]*providers.RawMasterComponent, 0),
		}
		if err := yaml.Unmarshal(provider.Config, &set); err != nil {
			panic("Failed to unmarshal server config")
		}

		if !s.validator.Validate(set) {
			panic("Incorrect master settings")
		}

		s.mSettings = set
		s.nodeID = "master"
	}
}

// Processes config records related to devices.
func (s *settingsProvider) processDeviceProvider(provider *rawProvider) error {
	d, err := s.loadDeviceProvider(provider)
	if nil == d {
		return err
	}

	s.devicesConfig = append(s.devicesConfig, d)
	return nil
}

// Loads devices.
func (s *settingsProvider) loadDeviceProvider(provider *rawProvider) (*providers.RawDevice, error) {
	if s.isWorker {
		return nil, nil
	}

	selector := providers.RawDeviceSelector{}
	if err := yaml.Unmarshal(provider.Config, &selector); err != nil {
		return nil, err
	}

	if selector.Name == "" {
		s.logger.Warn("Generating random name since it's not configured. "+
			"History will not be preserver between master restarts", common.LogDeviceTypeToken, provider.Provider,
			common.LogSystemToken, provider.System)
		selector.Name = namesgenerator.GetRandomName(0)
	}

	selector.Name = strings.ToLower(selector.Name)

	if provider.Provider == enums.DevGroup.String() {
		s.groups = append(s.groups, &providers.RawMasterComponent{
			Provider:  provider.Provider,
			Name:      selector.Name,
			RawConfig: provider.Config,
		})

		return nil, nil
	}

	deviceType := utils.VerifyDeviceProvider(provider.Provider)
	if deviceType == enums.DevUnknown && provider.System != systems.SysAPI.String() {
		s.logger.Warn("Ignoring device since type is unknown", common.LogDeviceTypeToken, provider.Provider,
			common.LogSystemToken, provider.System)
		return nil, nil
	}

	dup := false
	for _, e := range s.devicesConfig {
		if e.Selector.Name == selector.Name {
			s.logger.Warn("Ignoring device since name is duplicated", common.LogDeviceTypeToken, provider.Provider,
				common.LogSystemToken, provider.System, common.LogDeviceNameToken, selector.Name)
			dup = true
			break
		}
	}

	if dup {
		return nil, nil
	}

	d := &providers.RawDevice{
		Plugin:     provider.Provider,
		DeviceType: deviceType,
		Selector:   &selector,
		StrConfig:  string(provider.Config),
		Name:       selector.Name,
		IsAPI:      deviceType == enums.DevUnknown,
	}

	return d, nil
}

// Loads logger configuration.
func (s *settingsProvider) loadLoggerProvider(provs []*rawProvider) []*rawProvider {
	index := -1
	for i, v := range provs {
		if v.System != systems.SysLogger.String() {
			continue
		}

		index = i
		ctor := &logger.ConstructLogger{
			RawConfig:  v.Config,
			LoggerType: v.Provider,
			Loader:     s.pluginLoader,
			NodeID:     s.nodeID,
			Secret:     s.secrets,
		}
		log, err := logger.NewLoggerProvider(ctor)
		if err != nil {
			s.logger.Error("Failed to load logger", err, common.LogProviderToken, v.Provider)
			continue
		}

		s.logger = log

		validatorLogger := &logger.ConstructPluginLogger{
			SystemLogger: s.logger,
			Provider:     "go-home",
			System:       "validator",
		}

		s.validator.SetLogger(logger.NewPluginLogger(validatorLogger))
		s.secrets.(providers.IInternalSecret).UpdateLogger(s.logger)
	}

	if -1 != index {
		return append(provs[:index], provs[index+1:]...)
	}
	return provs
}

// Processes single provider config.
func (s *settingsProvider) parseProvider(provider *rawProvider) {
	s.logger.Debug("Processing config", common.LogProviderToken, provider.Provider,
		common.LogSystemToken, provider.System)

	sys, err := systems.SystemTypeString(provider.System)
	if err != nil {
		s.logger.Warn("Unknown provider's system", common.LogProviderToken, provider.Provider,
			common.LogSystemToken, provider.System)
		return
	}
	switch sys {
	case systems.SysBus:
		if nil != s.bus {
			s.logger.Warn("Duplicated service bus", common.LogProviderToken, provider.Provider,
				common.LogSystemToken, provider.System)
			return
		}

		ctor := &bus.ConstructBus{
			RawConfig: provider.Config,
			Provider:  provider.Provider,
			Logger:    s.PluginLogger(systems.SysBus, provider.Provider),
			Loader:    s.pluginLoader,
			NodeID:    s.nodeID,
			Secret:    s.secrets,
		}
		s.bus, err = bus.NewServiceBusProvider(ctor)
		if err != nil {
			s.bus = nil
		}
	case systems.SysSecurity:
		s.processSecurity(provider)
	case systems.SysTrigger:
		cmp := s.getMasterComponents(provider)
		if nil == cmp {
			return
		}
		s.triggers = append(s.triggers, cmp)
	case systems.SysAPI:
		s.loadAPI(provider)
	case systems.SysUI:
		s.loadUIProviders(provider)
	case systems.SysStorage:
		s.processStorage(provider)
	}

	if err != nil {
		s.logger.Error("Failed to load plugin", err, common.LogProviderToken, provider.Provider,
			common.LogSystemToken, provider.System)
	}
}

// Processes storage provider.
func (s *settingsProvider) processStorage(provider *rawProvider) {
	if s.isWorker {
		return
	}

	if s.storage != nil {
		s.logger.Warn("Duplicated storage provider",
			common.LogProviderToken, provider.Provider, common.LogSystemToken, provider.System)
		return
	}

	ctor := &storage.ConstructStorage{
		Logger:    s.logger,
		Secret:    s.secrets,
		Provider:  provider.Provider,
		RawConfig: provider.Config,
		Loader:    s.pluginLoader,
	}

	s.storage = storage.NewStorageProvider(ctor)
}

// Processes security groups.
func (s *settingsProvider) processSecurity(provider *rawProvider) {
	if s.isWorker {
		return
	}

	parts := strings.Split(provider.Provider, "/")
	if 2 == len(parts) && "user" == parts[0] {
		if nil != s.rawUsersProvider {
			s.logger.Warn("Duplicated user provider",
				common.LogProviderToken, provider.Provider, common.LogSystemToken, provider.System)
			return
		}

		provider.Provider = parts[1]
		s.rawUsersProvider = provider
		return
	}

	if provider.Provider == "role" {
		group := &providers.SecRole{}
		err := yaml.Unmarshal(provider.Config, &group)
		if err != nil {
			s.logger.Error("Failed to unmarshal security group", err,
				common.LogProviderToken, provider.Provider, common.LogSystemToken, provider.System)
			return
		}

		if !s.validator.Validate(group) {
			s.logger.Error("Failed to validate security group", err,
				common.LogProviderToken, provider.Provider, common.LogSystemToken, provider.System)
			return
		}

		s.rawRoles = append(s.rawRoles, group)
		return
	}

	s.logger.Warn("Unknown security record",
		common.LogProviderToken, provider.Provider, common.LogSystemToken, provider.System)
}

// Loads extended API.
func (s *settingsProvider) loadAPI(provider *rawProvider) {
	cmp := s.getMasterComponents(provider)
	if nil == cmp {
		return
	}

	d, err := s.loadDeviceProvider(provider)
	if err != nil || nil == d {
		return
	}

	for _, v := range s.extendedAPIs {
		if v.Name == d.Name {
			s.logger.Warn("Duplicated API name, ignoring",
				common.LogProviderToken, provider.Provider, "name", d.Name)
			return
		}
	}

	if len(d.Selector.Selectors) > 0 {
		s.devicesConfig = append(s.devicesConfig, d)
	}

	cmp.Name = d.Name
	s.extendedAPIs = append(s.extendedAPIs, cmp)
}

// Loads master component.
func (s *settingsProvider) getMasterComponents(provider *rawProvider) *providers.RawMasterComponent {
	if s.isWorker {
		return nil
	}

	cmp := &providers.RawMasterComponent{
		RawConfig: provider.Config,
		Provider:  provider.Provider,
	}

	if "" == cmp.Name {
		cmp.Name = cmp.Provider
	}

	return cmp
}

// Loads UI configuration.
func (s *settingsProvider) loadUIProviders(provider *rawProvider) {
	if s.isWorker {
		return
	}
	c := &providers.RawDeviceSelector{}
	err := yaml.Unmarshal(provider.Config, c)
	if err != nil {
		s.logger.Error("Failed to unmarshal UI provider", err)
		return
	}

	if "" == c.Name {
		s.logger.Warn("Skipping UI provider since name is nul", common.LogProviderToken, provider.Provider)
		return
	}

	switch provider.Provider {
	case "location":
		s.mSettings.Locations = append(s.mSettings.Locations, &providers.RawMasterComponent{
			Provider:  provider.Provider,
			Name:      c.Name,
			RawConfig: provider.Config,
		})
	}
}
