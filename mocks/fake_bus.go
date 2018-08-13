//+build !release

package mocks

import (
	"errors"

	"github.com/go-home-io/server/plugins/bus"
)

// Service bus provider.
type fakeServiceBus struct {
	publishToWorkerCallback func(string, ...interface{})
	publishCallback         func(...interface{})

	sub chan bus.RawMessage
}

func (s *fakeServiceBus) SubscribeStr(channel string, queue chan bus.RawMessage) error {
	s.sub = queue
	return nil
}

func (s *fakeServiceBus) PublishStr(channel string, messages ...interface{}) {
	if nil != s.publishCallback {
		s.publishCallback(messages...)
	}
}

// Subscribing to the channel.
func (s *fakeServiceBus) Subscribe(channel bus.ChannelName, queue chan bus.RawMessage) error {
	s.sub = queue
	return nil
}

// Syntax sugar around worker channels.
func (s *fakeServiceBus) SubscribeToWorker(workerName string, queue chan bus.RawMessage) error {
	s.sub = queue
	return nil
}

// Un-subscribing from the channel.
func (s *fakeServiceBus) Unsubscribe(channel string) {
}

// Publishing a new message.
func (s *fakeServiceBus) Publish(channel bus.ChannelName, messages ...interface{}) {
	if nil != s.publishCallback {
		s.publishCallback(messages...)
	}
}

// Syntax sugar around worker channels.
func (s *fakeServiceBus) PublishToWorker(workerName string, messages ...interface{}) {
	if nil != s.publishToWorkerCallback {
		s.publishToWorkerCallback(workerName, messages...)
	}
}

// Internal ping.
func (s *fakeServiceBus) Ping() error {
	return nil
}

func (s *fakeServiceBus) FakePublish(name string, msg bus.RawMessage) error {
	if nil == s.sub {
		return errors.New("no subs")
	}

	s.sub <- msg
	return nil
}

// Creating a fake service bus.
func FakeNewServiceBus(publish func(string, ...interface{})) *fakeServiceBus {
	return &fakeServiceBus{
		publishToWorkerCallback: publish,
	}
}

func FakeNewServiceBusRegular(publish func(...interface{})) *fakeServiceBus {
	return &fakeServiceBus{
		publishCallback: publish,
	}
}
