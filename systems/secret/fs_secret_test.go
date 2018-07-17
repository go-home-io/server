package secret

import (
	"github.com/go-home-io/server/utils"
	"os"
	"fmt"
	"io/ioutil"
	"github.com/go-home-io/server/plugins/secret"
	"github.com/go-home-io/server/mocks"
	"testing"
	"gopkg.in/yaml.v2"
)

func getInitData(logCallback func(string)) *secret.InitDataSecret {
	data := &secret.InitDataSecret{
		Logger: mocks.FakeNewLogger(logCallback),
	}

	return data
}

func getProvider(logCallback func(string), mockData string) *fsSecret {
	os.MkdirAll(utils.GetDefaultConfigsDir(), os.ModePerm)
	ioutil.WriteFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()),
		[]byte(mockData), os.ModePerm)
	prov := &fsSecret{}
	prov.Init(getInitData(logCallback))
	return prov
}

func cleanFile() {
	os.Remove(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	os.Remove(utils.GetDefaultConfigsDir())
}

// Tests proper file creation.
func TestFileCreation(t *testing.T) {
	os.MkdirAll(utils.GetDefaultConfigsDir(), os.ModePerm)
	defer os.Remove(utils.GetDefaultConfigsDir())
	prov := &fsSecret{}
	prov.Init(getInitData(nil))

	err := prov.Set("test", "data")
	if err != nil {
		t.FailNow()
	}

	fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	if err != nil {
		t.FailNow()
	}

	defer cleanFile()

	sec := make(map[string]string)
	err = yaml.Unmarshal(fileData, sec)
	if err != nil {
		t.FailNow()
	}

	if sec["test"] != "data" {
		t.Fail()
	}
}

// Tests file save error.
func TestSaveError(t *testing.T) {
	cleanFile()
	defer cleanFile()
	prov := &fsSecret{}
	prov.Init(getInitData(nil))

	err := prov.Set("test", "data")

	if err == nil {
		t.Fail()
	}
}

// Tests wrong file format.
func TestWrongFileFormat(t *testing.T){
	defer cleanFile()
	prov := getProvider(nil, "val: -1\n-1")
	_, err := prov.Get("val")
	if err == nil{
		t.Fail()
	}
}

// Test obtaining values and new logger.
func TestGetAndLoggerUpdate(t *testing.T) {
	defer cleanFile()
	oldLogger := false
	prov := getProvider(func(s string) {
		oldLogger = true
	}, "val1: data")

	newLogger := false
	prov.UpdateLogger(mocks.FakeNewLogger(func(s string) {
		newLogger = true
	}))


	_, err := prov.Get("val2")
	if err == nil {
		t.FailNow()
	}

	val, _ := prov.Get("val1")
	if val != "data" {
		t.Fail()
	}

	cleanFile()
	oldLogger = false
	newLogger = false
	prov.Set("1", "1")
	if oldLogger || !newLogger {
		t.FailNow()
	}
}
