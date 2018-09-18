package bus

import (
	"testing"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/bus"
)

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

// Tests proper error handling.
func TestErrorLoading(t *testing.T) {
	ctor := &ConstructBus{
		Loader: mocks.FakeNewPluginLoader(nil),
		Logger: mocks.FakeNewLogger(nil),
	}

	p, err := NewServiceBusProvider(ctor)
	if p != nil || err == nil {
		t.Fail()
	}
}

// Tests correct methods invocation.
func TestServiceBus(t *testing.T) {
	f := &fakePlugin{}
	ctor := &ConstructBus{
		Loader: mocks.FakeNewPluginLoader(f),
		Logger: mocks.FakeNewLogger(nil),
	}

	p, err := NewServiceBusProvider(ctor)
	if p == nil || err != nil {
		t.FailNow()
	}

	p.Ping()
	if !f.ping {
		t.Error("Ping failed")
		t.Fail()
	}

	p.Publish(bus.ChDeviceUpdates, nil)
	if !f.pub {
		t.Error("Publish failed")
		t.Fail()
	}

	f.pub = false
	p.PublishStr("test", nil)
	if !f.pub {
		t.Error("PublishStr failed")
		t.Fail()
	}

	f.pub = false
	p.PublishToWorker("test", nil)
	if !f.pub {
		t.Error("PublishToWorker failed")
		t.Fail()
	}

	p.Subscribe(bus.ChDiscovery, nil)
	if !f.sub {
		t.Error("Subscribe failed")
		t.Fail()
	}

	f.sub = false
	p.SubscribeStr("test", nil)
	if !f.sub {
		t.Error("SubscribeStr failed")
		t.Fail()
	}

	f.sub = false
	p.SubscribeToWorker("test", nil)
	if !f.sub {
		t.Error("SubscribeToWorker failed")
		t.Fail()
	}

	p.Unsubscribe("test")
	if !f.unSub {
		t.Error("Unsubscribe failed")
		t.Fail()
	}
}
