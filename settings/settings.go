package settings

import (
	"time"

	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
)

// SystemLogger returns default system logger.
func (s *settingsProvider) SystemLogger() common.ILoggerProvider {
	return s.logger
}

func (s *settingsProvider) Secrets() common.ISecretProvider {
	return s.secrets
}

// PluginLogger returns logger specifically for plugin provider.
func (s *settingsProvider) PluginLogger() common.ILoggerProvider {
	return s.pluginLogger
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

// IsWorker returns a flag indicating whether this instance is a worker.
func (s *settingsProvider) IsWorker() bool {
	return s.isWorker
}

// DevicesConfig returns raw devices configs.
func (s *settingsProvider) DevicesConfig() []*providers.RawDevice {
	return s.devicesConfig
}

// Security returns a security provider.
func (s *settingsProvider) Security() providers.ISecurityProvider {
	return s.securityProvider
}

// Triggers returns a list of known triggers.
func (s *settingsProvider) Triggers() []*providers.RawMasterComponent {
	return s.triggers
}

// FanOut returns fan out channel.
func (s *settingsProvider) FanOut() providers.IInternalFanOutProvider {
	return s.fanOut
}

// ExtendedAPIs returns a list of known APIs.
func (s *settingsProvider) ExtendedAPIs() []*providers.RawMasterComponent {
	return s.extendedAPIs
}

// Groups returns a list of known groups.
func (s *settingsProvider) Groups() []*providers.RawMasterComponent {
	return s.groups
}

// Storage returns a storage provider.
func (s *settingsProvider) Storage() providers.IStorageProvider {
	return s.storage
}

// Timezone returns configured timezone.
func (s *settingsProvider) Timezone() *time.Location {
	return s.timezone
}
