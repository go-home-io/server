//+build !release

package mocks

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
)

type IFakeSettings interface {
	AddLoader(returnOj interface{})
	AddSBCallback(func(...interface{}))
}

type fakeSettings struct {
	isWorker bool
	logger   common.ILoggerProvider
	cron     providers.ICronProvider
	bus      providers.IBusProvider
	devices  []*providers.RawDevice
	security providers.ISecurityProvider
	fanOut   providers.IInternalFanOutProvider
	storage  providers.IStorageProvider
	loader   providers.IPluginLoaderProvider
}

func (f *fakeSettings) Storage() providers.IStorageProvider {
	if nil != f.storage {
		return f.storage
	}

	return FakeNewStorage()
}

func (f *fakeSettings) Groups() []*providers.RawMasterComponent {
	return nil
}

func (f *fakeSettings) ExtendedAPIs() []*providers.RawMasterComponent {
	return nil
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
	return f.loader
}

func (f *fakeSettings) Validator() providers.IValidatorProvider {
	return nil
}

func (f *fakeSettings) WorkerSettings() *providers.WorkerSettings {
	return &providers.WorkerSettings{}
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

func (f *fakeSettings) DevicesConfig() []*providers.RawDevice {
	return f.devices
}

func (f *fakeSettings) Security() providers.ISecurityProvider {
	return f.security
}

func (f *fakeSettings) Triggers() []*providers.RawMasterComponent {
	return []*providers.RawMasterComponent{}
}

func (f *fakeSettings) FanOut() providers.IInternalFanOutProvider {
	return f.fanOut
}

func (f *fakeSettings) AddLoader(returnOj interface{}) {
	f.loader = FakeNewPluginLoader(returnOj)
}

func (f *fakeSettings) AddSBCallback(cb func(...interface{})) {
	f.bus.(*fakeServiceBus).publishCallback = cb
}

func FakeNewSettings(sbPublish func(string, ...interface{}), isWorker bool,
	devices []*providers.RawDevice, logCallback func(string)) providers.ISettingsProvider {
	return &fakeSettings{
		isWorker: isWorker,
		bus:      FakeNewServiceBus(sbPublish),
		logger:   FakeNewLogger(logCallback),
		cron:     FakeNewCron(),
		devices:  devices,
		fanOut:   FakeNewFanOut(),
	}
}

func FakeNewSettingsWithUserStorage(sec providers.ISecurityProvider) *fakeSettings {
	return &fakeSettings{
		security: sec,
	}
}
