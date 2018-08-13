package storage

import (
	"testing"
	"time"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/storage"
)

// Tests that empty storage is not causing panic.
func TestNewEmptyStorageProvider(t *testing.T) {
	p := NewEmptyStorageProvider()
	p.History("test")
	p.Heartbeat("test")
	p.State(nil)
}

type fakePlugin struct {
	invokes int
}

func (f *fakePlugin) Init(*storage.InitDataStorage) error {
	return nil
}

func (f *fakePlugin) Heartbeat(string) {
	f.invokes++
}

func (f *fakePlugin) State(string, map[string]interface{}) {
	f.invokes++
}

func (f *fakePlugin) History(ID string, hrs int) map[string]map[int64]interface{} {
	f.invokes++

	if "test" == ID {
		return map[string]map[int64]interface{}{"on": {int64(f.invokes): nil}, "test": {int64(f.invokes): nil}}
	}

	return nil
}

// Tests that failed plugin is not causing panic.
func TestNewStorageProviderError(t *testing.T) {
	ctor := &ConstructStorage{
		Logger:    mocks.FakeNewLogger(nil),
		RawConfig: []byte(`
storeHeartbeat: true`),
		Loader:    mocks.FakeNewPluginLoader(nil),
		Provider:  "test",
		Secret:    mocks.FakeNewSecretStore(nil, true),
	}

	p := NewStorageProvider(ctor)
	p.History("test")
	p.Heartbeat("test")
	p.State(nil)
}

// Test correct invokes.
func TestNewStorageProvider(t *testing.T) {
	pl := &fakePlugin{}
	ctor := &ConstructStorage{
		Logger:    mocks.FakeNewLogger(nil),
		RawConfig: []byte(""),
		Loader:    mocks.FakeNewPluginLoader(pl),
		Provider:  "test",
		Secret:    mocks.FakeNewSecretStore(nil, true),
	}

	p := NewStorageProvider(ctor)

	update := &common.MsgDeviceUpdate{
		State:     map[enums.Property]interface{}{enums.PropOn: true},
		ID:        "test",
		FirstSeen: true,
		Name:      "test",
		Type:      enums.DevWeather,
	}

	p.Heartbeat("test")
	p.State(update)
	p.State(update)
	r := p.History("not_test")
	if 0 != len(r) {
		t.Error("Wrong history")
		t.Fail()
	}

	time.Sleep(1 * time.Second)

	r = p.History("test")
	if nil == r {
		t.Error("Wrong history with correct ID")
		t.Fail()
	}

	_, ok := r[enums.PropOn][4]

	if !ok || 1 != len(r) {
		t.Error("Wrong invokes count")
		t.Fail()
	}
}

// Tests correct heartbeat invocation.
func TestHeartbeat(t *testing.T) {
	pl := &fakePlugin{}
	ctor := &ConstructStorage{
		Logger: mocks.FakeNewLogger(nil),
		RawConfig: []byte(`
storeHeartbeat: true`),
		Loader:   mocks.FakeNewPluginLoader(pl),
		Provider: "test",
		Secret:   mocks.FakeNewSecretStore(nil, true),
	}

	p := NewStorageProvider(ctor)
	p.Heartbeat("test")
	p.Heartbeat("test")

	time.Sleep(1 * time.Second)

	if 2 != pl.invokes {
		t.Fail()
	}
}
