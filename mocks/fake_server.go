//+build !release

package mocks

import (
	"github.com/gobwas/glob"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/providers"
)

// IFakeServer adds additional capabilities to a fake server.
type IFakeServer interface {
	AddDevice(device *providers.KnownDevice)
}

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

// FakeNewServer creates a new fake server.
func FakeNewServer(callback func()) IFakeServer {
	return &fakeServer{
		callback: callback,
	}
}
