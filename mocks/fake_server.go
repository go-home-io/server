package mocks

import (
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/gobwas/glob"
)

type fakeServer struct {
	callback func()
}

func (*fakeServer) Start() {
}

func (f *fakeServer) InternalCommandInvokeDeviceCommand(deviceRegexp glob.Glob, cmd enums.Command,
	data map[string]interface{}) {
	if nil != f.callback {
		f.callback()
	}
}

func FakeNewServer(callback func()) providers.IServerProvider {
	return &fakeServer{
		callback: callback,
	}
}
