//+build !release

package mocks

import (
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
)

type fakeFanOut struct {
	inDeviceUpdates  chan *common.MsgDeviceUpdate
	outDeviceUpdates map[int64]chan *common.MsgDeviceUpdate

	inTriggerUpdates  chan string
	outTriggerUpdates map[int64]chan string
}

func (f *fakeFanOut) SubscribeDeviceUpdates() (int64, chan *common.MsgDeviceUpdate) {
	return 1, f.inDeviceUpdates
}

func (f *fakeFanOut) UnSubscribeDeviceUpdates(int64) {
}

func (f *fakeFanOut) ChannelInDeviceUpdates() chan *common.MsgDeviceUpdate {
	return f.inDeviceUpdates
}

func (f *fakeFanOut) SubscribeTriggerUpdates() (int64, chan string) {
	return 1, f.inTriggerUpdates
}

func (f *fakeFanOut) UnSubscribeTriggerUpdates(int64) {
}

func (f *fakeFanOut) ChannelInTriggerUpdates() chan string {
	return f.inTriggerUpdates
}

func FakeNewFanOut() providers.IInternalFanOutProvider {
	return &fakeFanOut{
		inTriggerUpdates:  make(chan string, 10),
		outTriggerUpdates: make(map[int64]chan string),
		inDeviceUpdates:   make(chan *common.MsgDeviceUpdate, 10),
		outDeviceUpdates:  make(map[int64]chan *common.MsgDeviceUpdate),
	}
}
