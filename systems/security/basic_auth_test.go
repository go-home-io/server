package security

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/user"
	"go-home.io/x/server/utils"
	"golang.org/x/crypto/bcrypt"
)

const tmpDir = "./temp_users"

func getInitData(logCallback func(string), secretData map[string]string) *user.InitDataUserStorage {
	data := &user.InitDataUserStorage{
		Logger: mocks.FakeNewLogger(logCallback),
		Secret: mocks.FakeNewSecretStore(secretData, false),
	}

	return data
}

func getProvider(t *testing.T, logCallback func(string), mockData string) *basicAuthProvider {
	utils.ConfigDir = tmpDir
	err := os.MkdirAll(utils.GetDefaultConfigsDir(), os.ModePerm)
	require.NoError(t, err, "mkdir failed")
	err = ioutil.WriteFile(fmt.Sprintf("%s/_users", utils.GetDefaultConfigsDir()),
		[]byte(mockData), os.ModePerm)
	require.NoError(t, err, "write file failed")
	prov := &basicAuthProvider{}
	err = prov.Init(getInitData(logCallback, nil))
	require.NoError(t, err, "prov failed")
	return prov
}

func getAuthHeader(usr, pwd string) map[string][]string {
	pair := fmt.Sprintf("%s:%s", usr, pwd)

	return map[string][]string{"Authorization": {
		fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(pair)))}}
}

func getCookieHeader(usr, pwd string) map[string][]string {
	pair := fmt.Sprintf("%s:%s", usr, pwd)

	return map[string][]string{"Cookie": {
		fmt.Sprintf("Wrong=Data;X-Authorization;X-Authorization=Basic %s",
			base64.StdEncoding.EncodeToString([]byte(pair)))}}
}

func getFileRecord(usr, pwd string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(pwd), 0)
	return fmt.Sprintf("%s:%s", usr, string(b))
}

func cleanup(t *testing.T) {
	err := os.RemoveAll(tmpDir)
	require.NoError(t, err, "cleanup failed")
}

// Tests store access error.
//noinspection GoUnhandledErrorResult
func TestFileAccessError(t *testing.T) {
	prov := basicAuthProvider{}
	logFound := false
	prov.Init(getInitData(func(m string) {
		if m == "_users file is not found, going to use secret store only" {
			logFound = true
		}
	}, nil))

	assert.True(t, logFound)
}

// Tests incorrect header.
//noinspection GoUnhandledErrorResult
func TestNoHeader(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, nil))
	_, err := prov.Authorize(map[string][]string{})
	assert.IsType(t, &ErrNoHeader{}, err)
}

// Tests header with wrong data.
//noinspection GoUnhandledErrorResult
func TestNoBase64Header(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, nil))
	_, err := prov.Authorize(map[string][]string{"Authorization": {"Basic Wrong header"}})
	assert.IsType(t, &ErrIncorrectHeader{}, err)
}

// Tests correctly encoded header with wrong user.
//noinspection GoUnhandledErrorResult
func TestIncorrectAuthHeader(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, nil))
	_, err := prov.Authorize(map[string][]string{
		"Authorization": {"Basic " + base64.StdEncoding.EncodeToString([]byte("Wrong header"))}})
	assert.IsType(t, &ErrCorruptedHeader{}, err)
}

// Tests that user from file storage read correctly.
func TestUserFromFile(t *testing.T) {
	data := "user1\n" + getFileRecord("user1", "123")
	prov := getProvider(t, nil, data)
	defer cleanup(t)
	h1 := getAuthHeader("user1", "123")
	h2 := make(map[string][]string)
	h2["Fake"] = []string{"header"}
	h2["Authorization"] = h1["Authorization"]
	usr, err := prov.Authorize(h2)
	require.NoError(t, err)
	assert.Equal(t, "user1", usr)
}

// Tests that incorrect format is ignored.
func TestIncorrectFIleFormat(t *testing.T) {
	data := "\n" +
		"user1\n" +
		getFileRecord("user1", "123") + ":wrong\n" +
		"\n"

	prov := getProvider(t, nil, data)
	defer cleanup(t)
	_, err := prov.Authorize(getAuthHeader("user1", "123"))
	assert.IsType(t, &ErrUserNotFound{}, err)
}

// Tests authentication from secret store.
//noinspection GoUnhandledErrorResult
func TestUserFromSecret(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, map[string]string{"user1": "123"}))
	usr, err := prov.Authorize(getCookieHeader("user1", "123"))
	assert.NoError(t, err)
	assert.Equal(t, "user1", usr)
}

// Tests incorrect header.
//noinspection GoUnhandledErrorResult
func TestIncorrectHeader(t *testing.T) {
	prov := basicAuthProvider{}
	prov.Init(getInitData(nil, map[string]string{"user1": "123"}))
	headers := getAuthHeader("user1", "123")
	headers["Authorization"] = append(headers["Authorization"], "Another")
	usr, err := prov.Authorize(headers)
	assert.Error(t, err)
	assert.NotEqual(t, "user1", usr)
}
