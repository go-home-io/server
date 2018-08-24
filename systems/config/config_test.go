package config

import (
	"testing"

	"fmt"
	"os"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/config"
)

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
		Secret:  mocks.FakeNewSecretStore(nil, true),
		Logger:  mocks.FakeNewLogger(nil),
		Options: map[string]string{common.LogProviderToken: "test"},
		Loader:  mocks.FakeNewPluginLoader(f),
	}

	c := NewConfigProvider(ctor)
	c.Load()

	if !f.loadCalled {
		t.FailNow()
	}
}

// Tests fallback to FS config.
func TestFallbackToFsProvider(t *testing.T) {
	ctor := &ConstructConfig{
		Secret:  mocks.FakeNewSecretStore(nil, true),
		Logger:  mocks.FakeNewLogger(nil),
		Options: map[string]string{common.LogProviderToken: "test", "location": tmpDir},
		Loader:  mocks.FakeNewPluginLoader(nil),
	}

	c := NewConfigProvider(ctor)

	os.MkdirAll(tmpDir, os.ModePerm)
	defer cleanup()
	f, err := os.OpenFile(fmt.Sprintf("%s/data.yaml", tmpDir), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.FailNow()
	}
	_, err = f.Write([]byte("test"))
	f.Close()

	ii := 0
	for d := range c.Load() {
		if 0 == ii {
			if string(d) != "test" {
				t.Error("Wrong data")
				t.FailNow()
			}
		}

		ii++
	}

	if 1 != ii {
		t.Error("Wrong len")
		t.FailNow()
	}
}
