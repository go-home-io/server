package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/config"
	"github.com/go-home-io/server/utils"
)

const tmpDir = "./temp_config"

func cleanup() {
	os.RemoveAll(tmpDir)
}

// Tests correct loading.
func TestFSConfig(t *testing.T) {
	utils.ConfigDir = tmpDir
	os.MkdirAll(tmpDir, os.ModePerm)
	defer cleanup()
	f, err := os.OpenFile(fmt.Sprintf("%s/data.yaml", tmpDir), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.FailNow()
	}
	_, err = f.Write([]byte("test"))
	f.Close()

	f, err = os.OpenFile(fmt.Sprintf("%s/_data.yaml", tmpDir), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.FailNow()
	}
	f.Write([]byte("test1"))
	f.Close()

	f, err = os.OpenFile(fmt.Sprintf("%s/data.txt", tmpDir), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.FailNow()
	}
	f.Write([]byte("test1"))
	f.Close()

	c := &fsConfig{}
	err = c.Init(&config.InitDataConfig{
		Logger:  mocks.FakeNewLogger(nil),
		Options: map[string]string{},
		Secret:  mocks.FakeNewSecretStore(nil, true),
	})

	if err != nil {
		t.FailNow()
	}

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
