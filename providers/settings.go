package providers

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/systems"
)

// ISettingsProvider defines settings loader provider logic.
type ISettingsProvider interface {
	SystemLogger() common.ILoggerProvider
	PluginLogger(system systems.SystemType, provider string) common.ILoggerProvider
	ServiceBus() IBusProvider
	NodeID() string
	Cron() ICronProvider
	PluginLoader() IPluginLoaderProvider
	Validator() IValidatorProvider
	WorkerSettings() *WorkerSettings
	MasterSettings() *MasterSettings
	IsWorker() bool
	DevicesConfig() []RawDevice
	Secrets() common.ISecretProvider
}

// RawDeviceSelector has data required for understanding
// which worker should be picked for the device.
type RawDeviceSelector struct {
	Name      string            `yaml:"name"`
	Selectors map[string]string `yaml:"workerSelectors"`
}

// RawDevice has data describing data about device,
// loaded from config files.
type RawDevice struct {
	Plugin     string
	DeviceType enums.DeviceType
	Selector   *RawDeviceSelector
	StrConfig  string
	Name       string
}

// MasterSettings has configured data for master node.
type MasterSettings struct {
	Port         int `yaml:"port" validate:"required,port" default:"8000"`
	DelayedStart int `yaml:"delayedStart" validate:"gte=0"`
}

// WorkerSettings has configured data for worker node.
type WorkerSettings struct {
	Name       string            `yaml:"name"`
	Properties map[string]string `yaml:"properties"`
	MaxDevices int               `yaml:"maxDevices" validate:"gte=0,lte=1000" default:"99"`
}
