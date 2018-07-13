//+build !release

package mocks

import (
	"github.com/go-home-io/server/plugins/bus"
)

// Service bus provider.
type fakeServiceBus struct {
	publishToWorker func(string, ...interface{})
}

// Subscribing to the channel.
func (s *fakeServiceBus) Subscribe(channel bus.ChannelName, queue chan bus.RawMessage) error {
	return nil
}

// Syntax sugar around worker channels.
func (s *fakeServiceBus) SubscribeToWorker(workerName string, queue chan bus.RawMessage) error {
	return nil
}

// Un-subscribing from the channel.
func (s *fakeServiceBus) Unsubscribe(channel string) {
}

// Publishing a new message.
func (s *fakeServiceBus) Publish(channel bus.ChannelName, messages ...interface{}) {
}

// Syntax sugar around worker channels.
func (s *fakeServiceBus) PublishToWorker(workerName string, messages ...interface{}) {
	if nil != s.publishToWorker {
		s.publishToWorker(workerName, messages...)
	}
}

// Internal ping.
func (s *fakeServiceBus) Ping() error {
	return nil
}

// Creating a fake service bus.
func FakeNewServiceBus(publish func(string, ...interface{})) *fakeServiceBus {
	return &fakeServiceBus{
		publishToWorker: publish,
	}
}
