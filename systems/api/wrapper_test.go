package api

import (
	"testing"
	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/api"
	"github.com/go-home-io/server/utils"
	"github.com/fortytw2/leaktest"
	"time"
	"github.com/go-home-io/server/providers"
	"encoding/json"
	"github.com/go-home-io/server/plugins/bus"
	"github.com/gobwas/glob"
	"github.com/go-home-io/server/plugins/device/enums"
)

type fakePlugin struct {
	com api.IExtendedAPICommunicator
	q   chan []byte
}

func (f *fakePlugin) FakeInit(prov api.IExtendedAPICommunicator) {
	f.q = make(chan []byte)
	f.com = prov
}

func (f *fakePlugin) Init(data *api.InitDataAPI) error {
	return nil
}

func (*fakePlugin) Routes() []string {
	return []string{}
}

func (*fakePlugin) Unload() {
}

func (f *fakePlugin) Subscribe() {
	f.com.Subscribe(f.q)
}

type fakeMessage struct {
	published int64
}

func (f *fakeMessage) SetSendTime(t int64) {
	f.published = t
}

func (f *fakePlugin) HasMessages() bool {
	afterCh := time.After(1 * time.Second)
	for {
		select {
		case _, ok := <-f.q:
			return ok
		case <-afterCh:
			return false
		}
	}
}

func getProvider(pl *fakePlugin, bus providers.IBusProvider, srvCallback func()) (providers.IExtendedAPIProvider, error) {
	ctor := &ConstructAPI{
		Server:     mocks.FakeNewServer(srvCallback),
		Name:       "test",
		Logger:     mocks.FakeNewLogger(nil),
		Loader:     mocks.FakeNewPluginLoader(pl),
		FanOut:     mocks.FakeNewFanOut(),
		ServiceBus: bus,
		IsServer:   true,
		Validator:  utils.NewValidator(mocks.FakeNewLogger(nil)),
		Secret:     mocks.FakeNewSecretStore(nil, true),
		Provider:   "test",
		RawConfig:  []byte(""),
	}

	p, err := NewExtendedAPIProvider(ctor)
	pl.FakeInit(p.(api.IExtendedAPICommunicator))
	return p, err

}

func getMessage(message api.IExtendedAPIMessage) bus.RawMessage {
	d, _ := json.Marshal(message)
	return bus.RawMessage{
		Body: d,
	}
}

// Tests that unload .
func TestCorrectUnload(t *testing.T) {
	defer leaktest.CheckTimeout(t, 1*time.Second)()
	wr, err := getProvider(&fakePlugin{}, mocks.FakeNewServiceBus(nil), nil)
	if err != nil {
		t.FailNow()
	}
	wr.Unload()
	time.Sleep(1 * time.Second)
}

// Tests that messages are delivered properly.
func TestMessagesDelivery(t *testing.T) {
	p := &fakePlugin{}
	b := mocks.FakeNewServiceBus(nil)

	_, err := getProvider(p, b, nil)
	p.Subscribe()
	if err != nil {
		t.FailNow()
	}

	b.FakePublish("", getMessage(&api.ExtendedAPIMessage{SendTime: utils.TimeNow() - bus.MsgTTLSeconds - 1}))
	if p.HasMessages() {
		t.FailNow()
	}

	b.FakePublish("", bus.RawMessage{Body: nil})
	if p.HasMessages() {
		t.FailNow()
	}

	b.FakePublish("", getMessage(&api.ExtendedAPIMessage{SendTime: utils.TimeNow()}))
	if !p.HasMessages() {
		t.FailNow()
	}
}

// Test wrapper methods.
func TestMethods(t *testing.T) {
	ctor := &ConstructAPI{
		Name:      "test",
		Loader:    mocks.FakeNewPluginLoader(nil),
		Logger:    mocks.FakeNewLogger(nil),
		IsServer:  true,
		Provider:  "test",
		RawConfig: []byte(""),
	}

	_, err := NewExtendedAPIProvider(ctor)
	if err == nil {
		t.FailNow()
	}

	called := false
	srvCalled := false
	wr, err := getProvider(&fakePlugin{}, mocks.FakeNewServiceBusRegular(func(i ...interface{}) {
		called = true
	}), func() {
		srvCalled = true
	})

	msg := &fakeMessage{}
	wr.(api.IExtendedAPICommunicator).Publish(msg)

	if !called || 0 == msg.published {
		t.FailNow()
	}

	wr.(api.IExtendedAPICommunicator).InvokeDeviceCommand(glob.MustCompile("**"), enums.CmdOn, nil)
	if !srvCalled {
		t.FailNow()
	}

	ctor = &ConstructAPI{
		Name:      "test",
		Logger:    mocks.FakeNewLogger(nil),
		Loader:    mocks.FakeNewPluginLoader(&fakePlugin{}),
		IsServer:  false,
		Validator: utils.NewValidator(mocks.FakeNewLogger(nil)),
		Secret:    mocks.FakeNewSecretStore(nil, true),
		Provider:  "test",
		RawConfig: []byte(""),
	}
	wr.(api.IExtendedAPICommunicator).InvokeDeviceCommand(glob.MustCompile("**"), enums.CmdOn, nil)
	if "test" != wr.ID() {
		t.FailNow()
	}
}
