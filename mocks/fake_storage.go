//+build !release

package mocks

import (
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
)

type fakeStorage struct {
}

func (*fakeStorage) Heartbeat(string) {
}

func (*fakeStorage) State(*common.MsgDeviceUpdate) {
}

func (*fakeStorage) History(string) map[enums.Property]map[int64]interface{} {
	return nil
}

// FakeNewStorage creates a new fake storage provider.
func FakeNewStorage() *fakeStorage {
	return &fakeStorage{}
}
