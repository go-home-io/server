//+build !release

package mocks

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
)

type fakeSettings struct {
	isWorker bool
	logger   common.ILoggerProvider
	cron     providers.ICronProvider
	bus      providers.IBusProvider
	devices  []providers.RawDevice
}

func (f *fakeSettings) SystemLogger() common.ILoggerProvider {
	return f.logger
}

func (f *fakeSettings) Secrets() common.ISecretProvider {
	return nil
}

func (f *fakeSettings) PluginLogger(system systems.SystemType, provider string) common.ILoggerProvider {
	return f.logger
}

func (f *fakeSettings) ServiceBus() providers.IBusProvider {
	return f.bus
}

func (f *fakeSettings) NodeID() string {
	return "go-home-tests"
}

func (f *fakeSettings) Cron() providers.ICronProvider {
	return f.cron
}

func (f *fakeSettings) PluginLoader() providers.IPluginLoaderProvider {
	return nil
}

func (f *fakeSettings) Validator() providers.IValidatorProvider {
	return nil
}

func (f *fakeSettings) WorkerSettings() *providers.WorkerSettings {
	return nil
}

func (f *fakeSettings) MasterSettings() *providers.MasterSettings {
	return &providers.MasterSettings{
		Port:         9999,
		DelayedStart: 1,
	}
}

func (f *fakeSettings) IsWorker() bool {
	return f.isWorker
}

func (f *fakeSettings) DevicesConfig() []providers.RawDevice {
	return f.devices
}

func FakeNewSettings(sbPublish func(string, ...interface{}), isWorker bool,
	devices []providers.RawDevice, logCallback func(string)) *fakeSettings {
	return &fakeSettings{
		isWorker: isWorker,
		bus:      FakeNewServiceBus(sbPublish),
		logger:   FakeNewLogger(logCallback),
		cron:     FakeNewCron(),
		devices:  devices,
	}
}
