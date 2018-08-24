//+build !release

package mocks

import (
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
)

type fakeGroups struct {
	groupID  string
	devices  []string
	callback func()
}

func (f *fakeGroups) ID() string {
	return f.groupID
}

func (f *fakeGroups) Devices() []string {
	return f.devices
}

func (f *fakeGroups) Icon() string {
	return ""
}

func (f *fakeGroups) InvokeCommand(enums.Command, map[string]interface{}) {
	if nil != f.callback {
		f.callback()
	}
}

func FakeNewGroupProvider(groupID string, devices []string, callback func()) providers.IGroupProvider {
	return &fakeGroups{
		devices:  devices,
		groupID:  groupID,
		callback: callback,
	}
}

func FakeNewLocationProvider(groupID string, devices []string, callback func()) providers.ILocationProvider {
	return &fakeGroups{
		devices:  devices,
		groupID:  groupID,
		callback: callback,
	}
}
