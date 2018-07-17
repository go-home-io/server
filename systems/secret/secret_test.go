package secret

import (
	"testing"
	"github.com/go-home-io/server/mocks"
	"io/ioutil"
	"fmt"
	"github.com/go-home-io/server/utils"
	"os"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/secret"
)

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

func TestFallbackToDefault(t *testing.T) {
	os.MkdirAll(utils.GetDefaultConfigsDir(), os.ModePerm)
	defer cleanFile()

	ctor := &ConstructSecret{
		Loader:  mocks.FakeNewPluginLoader(nil),
		Logger:  mocks.FakeNewLogger(nil),
		Options: map[string]string{common.LogProviderToken: "fs"},
	}

	prov := NewSecretProvider(ctor)
	err := prov.Set("1", "1")
	if err != nil {
		t.FailNow()
	}

	_, err = ioutil.ReadFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	if err != nil {
		t.FailNow()
	}
}

func TestFallbackToDefaultWithWrongPlugin(t *testing.T) {
	os.MkdirAll(utils.GetDefaultConfigsDir(), os.ModePerm)
	defer cleanFile()

	ctor := &ConstructSecret{
		Loader:  mocks.FakeNewPluginLoader(nil),
		Logger:  mocks.FakeNewLogger(nil),
		Options: map[string]string{common.LogProviderToken: "test"},
	}

	prov := NewSecretProvider(ctor)
	err := prov.Set("1", "1")
	if err != nil {
		t.FailNow()
	}

	_, err = ioutil.ReadFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	if err != nil {
		t.FailNow()
	}
}

func TestSetFail(t *testing.T) {
	cleanFile()
	defer cleanFile()

	ctor := &ConstructSecret{
		Loader:  mocks.FakeNewPluginLoader(nil),
		Logger:  mocks.FakeNewLogger(nil),
		Options: map[string]string{},
	}

	prov := NewSecretProvider(ctor)

	err := prov.Set("1", "1")
	if err == nil {
		t.FailNow()
	}

	_, err = ioutil.ReadFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	if err == nil {
		t.FailNow()
	}

	_, err = prov.Get("1")
	if err == nil {
		t.FailNow()
	}
}

func TestPluginLoad(t *testing.T) {
	p := &fakePlugin{
		data: map[string]string{"val1": "data"},
	}
	ctor := &ConstructSecret{
		Loader:  mocks.FakeNewPluginLoader(p),
		Logger:  mocks.FakeNewLogger(nil),
		Options: map[string]string{common.LogProviderToken: "test"},
	}
	prov := NewSecretProvider(ctor)
	// Test no panic
	prov.UpdateLogger(mocks.FakeNewLogger(nil))

	v, _ := prov.Get("val1")
	if "data" != v {
		t.Fail()
	}
}
