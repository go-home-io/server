// Package bus contains service bus abstractions.
package bus

import (
	"fmt"

	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/bus"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
)

const (
	// Logs representation.
	logSystem = "service_bus"
)

// ConstructBus holds values required for a new service bus provider.
type ConstructBus struct {
	Provider  string
	Loader    providers.IPluginLoaderProvider
	RawConfig []byte
	Logger    common.ILoggerProvider
	NodeID    string
	Secret    common.ISecretProvider
}

// Service bus provider.
type provider struct {
	bus bus.IServiceBus
}

// NewServiceBusProvider constructs a new service bus provider.
func NewServiceBusProvider(ctor *ConstructBus) (providers.IBusProvider, error) {
	p := provider{}

	pluginLoadRequest := &providers.PluginLoadRequest{
		ExpectedType:   bus.TypeServiceBus,
		SystemType:     systems.SysBus,
		PluginProvider: ctor.Provider,
		RawConfig:      ctor.RawConfig,
		InitData: &bus.InitDataServiceBus{
			NodeID: ctor.NodeID,
			Logger: ctor.Logger,
			Secret: ctor.Secret,
		},
	}

	i, err := ctor.Loader.LoadPlugin(pluginLoadRequest)
	if err != nil {
		return nil, errors.Wrap(err, "plugin load failed")
	}

	p.bus = i.(bus.IServiceBus)
	return &p, nil
}

// Subscribe allows to subscribe to the incoming messages.
func (s *provider) Subscribe(channel bus.ChannelName, queue chan bus.RawMessage) error {
	return s.SubscribeStr(channel.String(), queue)
}

func (s *provider) SubscribeStr(channel string, queue chan bus.RawMessage) error {
	return s.bus.Subscribe(channel, queue)
}

// SubscribeToWorker is a syntax sugar around worker channels.
func (s *provider) SubscribeToWorker(workerName string, queue chan bus.RawMessage) error {
	return s.SubscribeStr(fmt.Sprintf(bus.ChWorkerFormat, workerName), queue)
}

// Unsubscribe removes bus subscription.
func (s *provider) Unsubscribe(channel string) {
	s.bus.Unsubscribe(channel)
}

// Publish allows to send a new message.
func (s *provider) Publish(channel bus.ChannelName, messages ...interface{}) {
	s.PublishStr(channel.String(), messages...)
}

func (s *provider) PublishStr(channel string, messages ...interface{}) {
	s.bus.Publish(channel, messages...)
}

// PublishToWorker is a syntax sugar around worker channels.
func (s *provider) PublishToWorker(workerName string, messages ...interface{}) {
	s.PublishStr(fmt.Sprintf(bus.ChWorkerFormat, workerName), messages...)
}

// Ping allows to validate whether service bus is available.
func (s *provider) Ping() error {
	return s.bus.Ping()
}
