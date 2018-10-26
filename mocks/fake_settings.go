//+build !release

package mocks

import (
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
)

//IFakeSettings adds additional capabilities to a fake server provider.
type IFakeSettings interface {
	AddLoader(returnOj interface{})
	AddSBCallback(func(...interface{}))
	AddMasterComponents(groups, externalAPI, triggers []*providers.RawMasterComponent)
	AddMasterSettings(*providers.MasterSettings)
}

type fakeSettings struct {
	isWorker       bool
	logger         common.ILoggerProvider
	cron           providers.ICronProvider
	bus            providers.IBusProvider
	devices        []*providers.RawDevice
	security       providers.ISecurityProvider
	fanOut         providers.IInternalFanOutProvider
	storage        providers.IStorageProvider
	loader         providers.IPluginLoaderProvider
	groups         []*providers.RawMasterComponent
	externalAPI    []*providers.RawMasterComponent
	triggers       []*providers.RawMasterComponent
	masterSettings *providers.MasterSettings
}

func (f *fakeSettings) Storage() providers.IStorageProvider {
	if nil != f.storage {
		return f.storage
	}

	return FakeNewStorage()
}

func (f *fakeSettings) Groups() []*providers.RawMasterComponent {
	return f.groups
}

func (f *fakeSettings) ExtendedAPIs() []*providers.RawMasterComponent {
	return f.externalAPI
}

func (f *fakeSettings) SystemLogger() common.ILoggerProvider {
	return f.logger
}

func (f *fakeSettings) Secrets() common.ISecretProvider {
	return nil
}

func (f *fakeSettings) PluginLogger() common.ILoggerProvider {
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
	return FakeNewValidator(true)
}

func (f *fakeSettings) WorkerSettings() *providers.WorkerSettings {
	return &providers.WorkerSettings{}
}

func (f *fakeSettings) MasterSettings() *providers.MasterSettings {
	if nil != f.masterSettings {
		return f.masterSettings
	}

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
	if nil != f.triggers {
		return f.triggers
	}
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

func (f *fakeSettings) AddMasterComponents(groups, externalAPI, triggers []*providers.RawMasterComponent) {
	f.groups = groups
	f.externalAPI = externalAPI
	f.triggers = triggers
}

func (f *fakeSettings) AddMasterSettings(m *providers.MasterSettings) {
	f.masterSettings = m
}

// FakeNewSettings creates a new fake settings provider.
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

// FakeNewSettingsWithUserStorage creates a new fake setting provider with custom users' storage.
func FakeNewSettingsWithUserStorage(sec providers.ISecurityProvider) *fakeSettings {
	return &fakeSettings{
		security: sec,
	}
}
