package security

import (
	"testing"
	"github.com/go-home-io/server/plugins/user"
	"github.com/go-home-io/server/mocks"
	"os"
	"fmt"
	"github.com/go-home-io/server/utils"
	"io/ioutil"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
)

func getInitData(logCallback func(string), secretData map[string]string) *user.InitDataUserStorage {
	data := &user.InitDataUserStorage{
		Logger: mocks.FakeNewLogger(logCallback),
		Secret: mocks.FakeNewSecretStore(secretData, false),
	}

	return data
}

func getProvider(logCallback func(string), mockData string) *basicAuthProvider {
	os.MkdirAll(utils.GetDefaultConfigsDir(), os.ModePerm)
	ioutil.WriteFile(fmt.Sprintf("%s/_users", utils.GetDefaultConfigsDir()), []byte(mockData), os.ModePerm)
	prov := &basicAuthProvider{}
	prov.Init(getInitData(logCallback, nil))
	return prov
}

func getAuthHeader(usr, pwd string) map[string][]string {
	pair := fmt.Sprintf("%s:%s", usr, pwd)

	return map[string][]string{"Authorization": {
		fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(pair)))}}
}

func getFileRecord(usr, pwd string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(pwd), 0)
	return fmt.Sprintf("%s:%s", usr, string(b))
}

func cleanFile() {
	os.Remove(fmt.Sprintf("%s/_users", utils.GetDefaultConfigsDir()))
	os.Remove(utils.GetDefaultConfigsDir())
}

// Tests store access error.
func TestFileAccessError(t *testing.T) {
	prov := basicAuthProvider{}
	logFound := false
	prov.Init(getInitData(func(m string) {
		if m == "_users file is not found, going to use secret store only" {
			logFound = true
		}
	}, nil))

	if !logFound {
		t.Fail()
	}
}

// Tests incorrect header.
func TestNoHeader(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, nil))
	_, err := prov.Authorize(map[string][]string{})
	if err.Error() != "header not found" {
		t.Fail()
	}
}

// Tests header with wrong data.
func TestNoBase64Header(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, nil))
	_, err := prov.Authorize(map[string][]string{"Authorization": {"Basic Wrong header"}})
	if err.Error() != "can't decode header" {
		t.Fail()
	}
}

// Tests correctly encoded header with wrong user.
func TestIncorrectAuthHeader(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, nil))
	_, err := prov.Authorize(map[string][]string{"Authorization":
	{"Basic " + base64.StdEncoding.EncodeToString([]byte("Wrong header"))}})
	if err.Error() != "wrong header" {
		t.Fail()
	}
}

// Tests that user from file storage read correctly.
func TestUserFromFile(t *testing.T) {
	data := "user1\n" + getFileRecord("user1", "123")
	prov := getProvider(nil, data)
	defer cleanFile()
	h1 := getAuthHeader("user1", "123")
	h2 := make(map[string][]string)
	h2["Fake"] = []string{"header"}
	h2["Authorization"] = h1["Authorization"]
	usr, err := prov.Authorize(h2)
	if err != nil || usr != "user1" {
		t.Fail()
	}
}

// Tests that incorrect format is ignored.
func TestIncorrectFIleFormat(t *testing.T) {
	data := "\n" +
		"user1\n" +
		getFileRecord("user1", "123") + ":wrong\n" +
		"\n"

	prov := getProvider(nil, data)
	defer cleanFile()
	_, err := prov.Authorize(getAuthHeader("user1", "123"))
	if err == nil || err.Error() != "user not found" {
		t.Fail()
	}
}

// Tests authentication from secret store.
func TestUserFromSecret(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, map[string]string{"user1": "123"}))
	usr, err := prov.Authorize(getAuthHeader("user1", "123"))
	if err != nil || usr != "user1" {
		t.Fail()
	}
}

// Tests incorrect header.
func TestIncorrectHeader(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, map[string]string{"user1": "123"}))
	headers := getAuthHeader("user1", "123")
	headers["Authorization"] = append(headers["Authorization"], "Another")
	usr, err := prov.Authorize(headers)
	if err == nil || usr == "user1" {
		t.Fail()
	}
}
