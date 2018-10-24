package secret

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/secret"
	"go-home.io/x/server/utils"
	"gopkg.in/yaml.v2"
)

const tmpDir = "./temp_secret"

func getInitData(logCallback func(string)) *secret.InitDataSecret {
	data := &secret.InitDataSecret{
		Logger: mocks.FakeNewLogger(logCallback),
	}

	return data
}

func createFolder(t *testing.T) {
	utils.ConfigDir = tmpDir
	err := os.MkdirAll(utils.GetDefaultConfigsDir(), os.ModePerm)
	require.NoError(t, err, "mkdir failed")
}

func writeFile(t *testing.T, data string) {
	createFolder(t)
	err := ioutil.WriteFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()),
		[]byte(data), os.ModePerm)
	require.NoError(t, err)
}

func cleanup(t *testing.T) {
	err := os.RemoveAll(tmpDir)
	require.NoError(t, err, "cleanup failed")
}

func getProvider(t *testing.T, logCallback func(string), mockData string) *fsSecret {
	writeFile(t, mockData)

	prov := &fsSecret{}
	err := prov.Init(getInitData(logCallback))
	require.NoError(t, err, "provider")
	return prov
}

// Tests proper file creation.
func TestFileCreation(t *testing.T) {
	defer cleanup(t)
	prov := getProvider(t, nil, "")

	err := prov.Set("test", "data")
	require.NoError(t, err, "set")

	fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/_secrets.yaml", utils.GetDefaultConfigsDir()))
	require.NoError(t, err, "read")

	sec := make(map[string]string)
	err = yaml.Unmarshal(fileData, sec)
	require.NoError(t, err, "yaml")
	assert.Equal(t, 1, len(sec), "length")
	assert.Equal(t, "data", sec["test"], "content")
}

// Tests file save error.
func TestSaveError(t *testing.T) {
	defer cleanup(t)
	prov := getProvider(t, nil, "")
	cleanup(t)
	err := prov.Set("test", "data")
	assert.Error(t, err)
}

// Tests wrong file format.
func TestWrongFileFormat(t *testing.T) {
	defer cleanup(t)
	prov := getProvider(t, nil, "val: -1\n-1")
	_, err := prov.Get("val")
	assert.Error(t, err)
}

// Test obtaining values and new logger.
func TestGetAndLoggerUpdate(t *testing.T) {
	defer cleanup(t)

	oldLogger := false
	prov := getProvider(t, func(s string) {
		oldLogger = true
	}, "val1: data")

	newLogger := false
	prov.UpdateLogger(mocks.FakeNewLogger(func(s string) {
		newLogger = true
	}))

	_, err := prov.Get("val2")
	assert.Error(t, err, "non existing key")

	val, _ := prov.Get("val1")
	assert.Equal(t, "data", val, "existing key")
	cleanup(t)

	oldLogger = false
	newLogger = false
	err = prov.Set("1", "1")
	assert.Error(t, err, "no folder")
	assert.False(t, oldLogger, "old")
	assert.True(t, newLogger, "new")
}
