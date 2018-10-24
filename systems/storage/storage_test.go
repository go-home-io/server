package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
	"go-home.io/x/server/plugins/storage"
)

// Fake plugin.
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

// Tests that empty storage is not causing panic.
func TestNewEmptyStorageProvider(t *testing.T) {
	defer func() {
		assert.Nil(t, recover(), "panic")
	}()

	p := NewEmptyStorageProvider()
	p.History("test")
	p.Heartbeat("test")
	p.State(nil)
}

// Tests that failed plugin is not causing panic.
func TestNewStorageProviderError(t *testing.T) {
	defer func() {
		assert.Nil(t, recover(), "panic")
	}()

	ctor := &ConstructStorage{
		PluginLogger: mocks.FakeNewLogger(nil),
		RawConfig: []byte(`
storeHeartbeat: true`),
		Loader:   mocks.FakeNewPluginLoader(nil),
		Provider: "test",
		Secret:   mocks.FakeNewSecretStore(nil, true),
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
		PluginLogger: mocks.FakeNewLogger(nil),
		RawConfig:    []byte(""),
		Loader:       mocks.FakeNewPluginLoader(pl),
		Provider:     "test",
		Secret:       mocks.FakeNewSecretStore(nil, true),
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
	assert.Equal(t, 0, len(r), "wrong history")
	time.Sleep(1 * time.Second)
	r = p.History("test")
	assert.NotNil(t, r, "wrong history with correct ID")
	_, ok := r[enums.PropOn][4]
	require.True(t, ok, "wrong invoke")
	assert.Equal(t, 1, len(r), "wrong invoke count")
}

// Tests correct heartbeat invocation.
func TestHeartbeat(t *testing.T) {
	pl := &fakePlugin{}
	ctor := &ConstructStorage{
		PluginLogger: mocks.FakeNewLogger(nil),
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
	assert.Equal(t, 2, pl.invokes)
}

// Tests proper device exclusion.
func TestExcluding(t *testing.T) {
	pl := &fakePlugin{}
	ctor := &ConstructStorage{
		PluginLogger: mocks.FakeNewLogger(nil),
		RawConfig: []byte(`
exclude:
  - test?`),
		Loader:   mocks.FakeNewPluginLoader(pl),
		Provider: "test",
		Secret:   mocks.FakeNewSecretStore(nil, true),
	}

	p := NewStorageProvider(ctor)

	update := &common.MsgDeviceUpdate{
		State:     map[enums.Property]interface{}{enums.PropOn: true},
		ID:        "test1",
		FirstSeen: true,
		Name:      "test",
		Type:      enums.DevWeather,
	}

	p.State(update)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 0, pl.invokes)
}

// Tests proper device ignore.
func TestIgnores(t *testing.T) {
	pl := &fakePlugin{}
	ctor := &ConstructStorage{
		PluginLogger: mocks.FakeNewLogger(nil),
		RawConfig:    []byte(""),
		Loader:       mocks.FakeNewPluginLoader(pl),
		Provider:     "test",
		Secret:       mocks.FakeNewSecretStore(nil, true),
	}

	p := NewStorageProvider(ctor)

	update := &common.MsgDeviceUpdate{
		State:     map[enums.Property]interface{}{enums.PropOn: true},
		ID:        "test1",
		FirstSeen: true,
		Name:      "test",
		Type:      enums.DevCamera,
	}

	p.State(update)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 0, pl.invokes)
}

// Tests proper device inclusion.
func TestIncluding(t *testing.T) {
	pl := &fakePlugin{}
	ctor := &ConstructStorage{
		PluginLogger: mocks.FakeNewLogger(nil),
		RawConfig: []byte(`
include:
  - test?`),
		Loader:   mocks.FakeNewPluginLoader(pl),
		Provider: "test",
		Secret:   mocks.FakeNewSecretStore(nil, true),
	}

	p := NewStorageProvider(ctor)

	update := &common.MsgDeviceUpdate{
		State:     map[enums.Property]interface{}{enums.PropOn: true},
		ID:        "test1",
		FirstSeen: true,
		Name:      "test",
		Type:      enums.DevCamera,
	}

	p.State(update)
	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, pl.invokes)
}
