// Package providers contains interfaces for internal system providers.
package providers

import "go-home.io/x/server/plugins/bus"

// IBusProvider defines service bus provider logic.
type IBusProvider interface {
	Subscribe(channel bus.ChannelName, queue chan bus.RawMessage) error
	SubscribeStr(channel string, queue chan bus.RawMessage) error
	SubscribeToWorker(workerName string, queue chan bus.RawMessage) error
	Unsubscribe(channel string)
	Publish(channel bus.ChannelName, messages ...interface{})
	PublishStr(channel string, messages ...interface{})
	PublishToWorker(workerName string, messages ...interface{})
	Ping() error
}
