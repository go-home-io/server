package secret

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/secret"
	"go-home.io/x/server/utils"
)

// Fake plugin.
type fakePlugin struct {
	data map[string]string
}

func (*fakePlugin) FakeInit(interface{}) {
}

func (f *fakePlugin) Get(name string) (string, error) {
	return f.data[name], nil
}

func (*fakePlugin) Set(string, string) error {
	return nil
}

func (*fakePlugin) Init(*secret.InitDataSecret) error {
	return nil
}

func (*fakePlugin) UpdateLogger(common.ILoggerProvider) {
}

// Tests fallback to default fs provider.
func TestFallbackToDefault(t *testing.T) {
	createFolder(t)
	defer cleanup(t)

	ctor := &ConstructSecret{
		Loader:       mocks.FakeNewPluginLoader(nil),
		PluginLogger: mocks.FakeNewLogger(nil),
		Options:      map[string]string{common.LogProviderToken: "fs"},
	}

	prov := NewSecretProvider(ctor)
	err := prov.Set("1", "1")
	require.NoError(t, err, "set")

	_, err = ioutil.ReadFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	assert.NoError(t, err, "read")
}

// Tests fallback to default fs if plugin failed to load.
func TestFallbackToDefaultWithWrongPlugin(t *testing.T) {
	createFolder(t)
	defer cleanup(t)

	ctor := &ConstructSecret{
		Loader:       mocks.FakeNewPluginLoader(nil),
		PluginLogger: mocks.FakeNewLogger(nil),
		Options:      map[string]string{common.LogProviderToken: "test"},
	}

	prov := NewSecretProvider(ctor)
	err := prov.Set("1", "1")
	require.NoError(t, err, "set")

	_, err = ioutil.ReadFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	assert.NoError(t, err, "read")
}

// Tests error during set.
func TestSetFail(t *testing.T) {
	cleanup(t)
	defer cleanup(t)

	ctor := &ConstructSecret{
		Loader:       mocks.FakeNewPluginLoader(nil),
		PluginLogger: mocks.FakeNewLogger(nil),
		Options:      map[string]string{},
	}

	prov := NewSecretProvider(ctor)
	err := prov.Set("1", "1")
	require.Error(t, err, "set")

	_, err = ioutil.ReadFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	require.Error(t, err, "read")

	_, err = prov.Get("1")
	assert.Error(t, err, "get")
}

// Tests correct plugin load.
func TestPluginLoad(t *testing.T) {
	p := &fakePlugin{
		data: map[string]string{"val1": "data"},
	}
	ctor := &ConstructSecret{
		Loader:       mocks.FakeNewPluginLoader(p),
		PluginLogger: mocks.FakeNewLogger(nil),
		Options:      map[string]string{common.LogProviderToken: "test"},
	}
	prov := NewSecretProvider(ctor)

	// Test no panic
	defer func() {
		assert.Nil(t, recover(), "panic")
	}()
	prov.UpdateLogger(mocks.FakeNewLogger(nil))

	v, _ := prov.Get("val1")
	assert.Equal(t, "data", v, "get")
}
