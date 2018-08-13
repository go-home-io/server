//+build !release

package mocks

import (
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/gobwas/glob"
)

type fakeServer struct {
	callback func()
	device   *providers.KnownDevice
}

func (f *fakeServer) GetDevice(string) *providers.KnownDevice {
	return f.device
}

func (f *fakeServer) PushMasterDeviceUpdate(*providers.MasterDeviceUpdate) {
}

func (f *fakeServer) Start() {
}

func (f *fakeServer) InternalCommandInvokeDeviceCommand(deviceRegexp glob.Glob, cmd enums.Command,
	data map[string]interface{}) {
	if nil != f.callback {
		f.callback()
	}
}

func (f *fakeServer) AddDevice(device *providers.KnownDevice) {
	f.device = device
}

func FakeNewServer(callback func()) *fakeServer {
	return &fakeServer{
		callback: callback,
	}
}
