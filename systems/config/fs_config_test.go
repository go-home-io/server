package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/config"
	"github.com/go-home-io/server/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tmpDir = "./temp_config"

func cleanup(t *testing.T) {
	err := os.RemoveAll(tmpDir)
	require.NoError(t, err, "cleanup failed")
}

func writeFile(t *testing.T, name string, data string) {
	utils.ConfigDir = tmpDir
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(t, err, "mkdir failed")

	f, err := os.OpenFile(fmt.Sprintf("%s/%s", tmpDir, name), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	require.NoError(t, err, "open failed")
	_, err = f.Write([]byte(data))
	require.NoError(t, err, "write failed")
	err = f.Close()
	require.NoError(t, err, "close failed")
}

// Tests correct loading.
func TestFSConfig(t *testing.T) {
	defer cleanup(t)
	writeFile(t, "data.yaml", "test_yaml")
	writeFile(t, "_data.yaml", "_test_yaml")
	writeFile(t, "data.txt", "test_txt")

	c := &fsConfig{}
	err := c.Init(&config.InitDataConfig{
		Logger: mocks.FakeNewLogger(func(s string) {
			println(s)
		}),
		Options: map[string]string{},
		Secret:  mocks.FakeNewSecretStore(nil, true),
	})

	require.NoError(t, err)

	ii := 0
	for d := range c.Load() {
		println("FILE")
		if 0 == ii {
			assert.Equal(t, "test_yaml", string(d), "data")
		}

		ii++
	}

	assert.Equal(t, 1, ii, "number of files")
}
