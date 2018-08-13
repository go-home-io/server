//+build !release

package mocks

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
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

func FakeNewStorage() *fakeStorage {
	return &fakeStorage{}
}
