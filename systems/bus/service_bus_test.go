package bus

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/bus"
	"github.com/stretchr/testify/assert"
)

// Fake bus plugin.
type fakePlugin struct {
	err   error
	sub   bool
	unSub bool
	pub   bool
	ping  bool
}

func (f *fakePlugin) Init(*bus.InitDataServiceBus) error {
	return f.err
}

func (f *fakePlugin) Subscribe(channel string, queue chan bus.RawMessage) error {
	f.sub = true
	return nil
}

func (f *fakePlugin) Unsubscribe(channel string) {
	f.unSub = true
}

func (f *fakePlugin) Publish(channel string, messages ...interface{}) {
	f.pub = true
}

func (f *fakePlugin) Ping() error {
	f.ping = true
	return nil
}

func (f *fakePlugin) reset() {
	f.sub = false
	f.unSub = false
	f.pub = false
	f.ping = false
}

// Tests proper error handling.
func TestErrorLoading(t *testing.T) {
	ctor := &ConstructBus{
		Loader: mocks.FakeNewPluginLoader(nil),
		Logger: mocks.FakeNewLogger(nil),
	}

	p, err := NewServiceBusProvider(ctor)
	assert.Error(t, err)
	assert.Nil(t, p)
}

// Tests correct methods invocation.
//noinspection GoUnhandledErrorResult
func TestServiceBus(t *testing.T) {
	f := &fakePlugin{}
	ctor := &ConstructBus{
		Loader: mocks.FakeNewPluginLoader(f),
		Logger: mocks.FakeNewLogger(nil),
	}

	p, err := NewServiceBusProvider(ctor)
	require.NoError(t, err)
	require.NotNil(t, p)

	f.reset()
	p.Ping()
	assert.True(t, f.ping, "ping")

	f.reset()
	p.Publish(bus.ChDeviceUpdates, nil)
	assert.True(t, f.pub, "publish")

	f.reset()
	p.PublishStr("test", nil)
	assert.True(t, f.pub, "publish string")

	f.reset()
	p.PublishToWorker("test", nil)
	assert.True(t, f.pub, "publish to worker")

	f.reset()
	p.Subscribe(bus.ChDiscovery, nil)
	assert.True(t, f.sub, "subscribe")

	f.reset()
	p.SubscribeStr("test", nil)
	assert.True(t, f.sub, "subscribe string")

	f.reset()
	p.SubscribeToWorker("test", nil)
	assert.True(t, f.sub, "subscribe to worker")

	f.reset()
	p.Unsubscribe("test")
	assert.True(t, f.unSub, "unsubscribe")
}
