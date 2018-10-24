package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/config"
)

// Fake config plugin.
type fakeConfig struct {
	loadCalled bool
}

func (*fakeConfig) Init(*config.InitDataConfig) error {
	return nil
}

func (f *fakeConfig) Load() chan []byte {
	f.loadCalled = true
	return make(chan []byte)
}

// Tests proper loading.
func TestNewConfigProvider(t *testing.T) {
	f := &fakeConfig{}
	ctor := &ConstructConfig{
		Secret:       mocks.FakeNewSecretStore(nil, true),
		PluginLogger: mocks.FakeNewLogger(nil),
		Options:      map[string]string{common.LogProviderToken: "test"},
		Loader:       mocks.FakeNewPluginLoader(f),
	}

	c := NewConfigProvider(ctor)
	c.Load()
	assert.True(t, f.loadCalled)
}

// Tests fallback to FS config.
func TestFallbackToFsProvider(t *testing.T) {
	ctor := &ConstructConfig{
		Secret:       mocks.FakeNewSecretStore(nil, true),
		PluginLogger: mocks.FakeNewLogger(nil),
		Options:      map[string]string{common.LogProviderToken: "test", "location": tmpDir},
		Loader:       mocks.FakeNewPluginLoader(nil),
	}

	c := NewConfigProvider(ctor)
	defer cleanup(t)
	writeFile(t, "data.yaml", "test")

	ii := 0
	for d := range c.Load() {
		if 0 == ii {
			assert.Equal(t, "test", string(d), "data")
		}

		ii++
	}

	assert.Equal(t, 1, ii, "number of files")
}
