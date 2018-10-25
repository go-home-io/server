//+build !release

package mocks

import (
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/providers"
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

// FakeNewGroupProvider creates a new fake group provider.
func FakeNewGroupProvider(groupID string, devices []string, callback func()) providers.IGroupProvider {
	return &fakeGroups{
		devices:  devices,
		groupID:  groupID,
		callback: callback,
	}
}

// FakeNewLocationProvider creates a new fake location provider.
func FakeNewLocationProvider(groupID string, devices []string, callback func()) providers.ILocationProvider {
	return &fakeGroups{
		devices:  devices,
		groupID:  groupID,
		callback: callback,
	}
}
