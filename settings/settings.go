package settings

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
	"github.com/go-home-io/server/systems/logger"
)

// SystemLogger returns default system logger.
func (s *settingsProvider) SystemLogger() common.ILoggerProvider {
	return s.logger
}

func (s *settingsProvider) Secrets() common.ISecretProvider {
	return s.secrets
}

// PluginLogger returns logger specifically for plugin provider.
func (s *settingsProvider) PluginLogger(system systems.SystemType, provider string) common.ILoggerProvider {
	ctor := &logger.ConstructPluginLogger{
		SystemLogger: s.logger,
		Provider:     provider,
		System:       system.String(),
	}
	return logger.NewPluginLogger(ctor)
}

// ServiceBus returns service bus provider.
func (s *settingsProvider) ServiceBus() providers.IBusProvider {
	return s.bus
}

// NodeID returns current instance node ID.
func (s *settingsProvider) NodeID() string {
	return s.nodeID
}

// Cron returns system's cron provider.
func (s *settingsProvider) Cron() providers.ICronProvider {
	return s.cron
}

// PluginLoader returns plugin loader provider.
func (s *settingsProvider) PluginLoader() providers.IPluginLoaderProvider {
	return s.pluginLoader
}

// Validator returns yaml validator provider.
func (s *settingsProvider) Validator() providers.IValidatorProvider {
	return s.validator
}

// WorkerSettings returns worker settings.
func (s *settingsProvider) WorkerSettings() *providers.WorkerSettings {
	return s.wSettings
}

// MasterSettings returns master settings.
func (s *settingsProvider) MasterSettings() *providers.MasterSettings {
	return s.mSettings
}

// IsWorker returns flag indicating whether this instance is a worker.
func (s *settingsProvider) IsWorker() bool {
	return s.isWorker
}

// DevicesConfig returns raw devices configs.
func (s *settingsProvider) DevicesConfig() []*providers.RawDevice {
	return s.devicesConfig
}

func (s *settingsProvider) Security() providers.ISecurityProvider {
	return s.securityProvider
}

func (s *settingsProvider) Triggers() []*providers.RawMasterComponent {
	return s.triggers
}

func (s *settingsProvider) FanOut() providers.IInternalFanOutProvider {
	return s.fanOut
}

func (s *settingsProvider) ExtendedAPIs() []*providers.RawMasterComponent {
	return s.extendedAPIs
}

func (s *settingsProvider) Groups() []*providers.RawMasterComponent {
	return s.groups
}
