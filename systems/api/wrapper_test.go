package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/api"
	"github.com/go-home-io/server/plugins/bus"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/utils"
	"github.com/gobwas/glob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fake plugin patch.
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

//noinspection GoUnhandledErrorResult
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

func getProvider(pl *fakePlugin, bus providers.IBusProvider, srvCallback func(),
	t *testing.T) (providers.IExtendedAPIProvider, error) {
	ctor := &ConstructAPI{
		Server:     mocks.FakeNewServer(srvCallback).(providers.IServerProvider),
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
	require.NoError(t, err, "provider")
	return p, err
}

func getMessage(message api.IExtendedAPIMessage) bus.RawMessage {
	d, _ := json.Marshal(message)
	return bus.RawMessage{
		Body: d,
	}
}

// Tests that unload.
func TestCorrectUnload(t *testing.T) {
	defer leaktest.CheckTimeout(t, 1*time.Second)()
	wr, err := getProvider(&fakePlugin{}, mocks.FakeNewServiceBus(nil), nil, t)
	require.NoError(t, err)
	wr.Unload()
	time.Sleep(1 * time.Second)
}

// Tests that messages are delivered properly.
//noinspection GoUnhandledErrorResult
func TestMessagesDelivery(t *testing.T) {
	p := &fakePlugin{}
	b := mocks.FakeNewServiceBus(nil)

	_, err := getProvider(p, b, nil, t)
	p.Subscribe()
	require.NoError(t, err)

	data := []struct {
		msg  bus.RawMessage
		gold bool
	}{
		{
			msg:  getMessage(&api.ExtendedAPIMessage{SendTime: utils.TimeNow() - bus.MsgTTLSeconds - 1}),
			gold: false,
		},
		{
			msg:  bus.RawMessage{Body: nil},
			gold: false,
		},
		{
			msg:  getMessage(&api.ExtendedAPIMessage{SendTime: utils.TimeNow()}),
			gold: true,
		},
	}

	for _, v := range data {
		b.FakePublish("", v.msg)
		assert.Equal(t, v.gold, p.HasMessages(), string(v.msg.Body))
	}
}

// Tests constructors.
func TestCtor(t *testing.T) {
	ctor := &ConstructAPI{
		Name:      "test",
		Loader:    mocks.FakeNewPluginLoader(nil),
		Logger:    mocks.FakeNewLogger(nil),
		IsServer:  true,
		Provider:  "test",
		RawConfig: []byte(""),
	}

	_, err := NewExtendedAPIProvider(ctor)
	assert.Error(t, err, "server")

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
	_, err = NewExtendedAPIProvider(ctor)
	assert.NoError(t, err, "worker")
}

// Test wrapper methods.
func TestMethods(t *testing.T) {
	called := false
	srvCalled := false
	wr, _ := getProvider(&fakePlugin{}, mocks.FakeNewServiceBusRegular(func(i ...interface{}) {
		called = true
	}), func() {
		srvCalled = true
	}, t)

	msg := &fakeMessage{}
	wr.(api.IExtendedAPICommunicator).Publish(msg)
	assert.True(t, called, "fake message not called")
	assert.NotEqual(t, 0, msg.published, "fake message not published")

	wr.(api.IExtendedAPICommunicator).InvokeDeviceCommand(glob.MustCompile("**"), enums.CmdOn, nil)
	assert.True(t, srvCalled, "server not called")

	wr.(api.IExtendedAPICommunicator).InvokeDeviceCommand(glob.MustCompile("**"), enums.CmdOn, nil)
	assert.Equal(t, "test", wr.ID(), "wrong ID")
}
