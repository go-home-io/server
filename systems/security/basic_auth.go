package security

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/user"
	"github.com/go-home-io/server/utils"
	"golang.org/x/crypto/bcrypt"
)

// Implements default user provider.
type basicAuthProvider struct {
	logger          common.ILoggerProvider
	secret          common.ISecretProvider
	presetPasswords map[string]string
}

// Init load regular htpsswds file.
// Passwords must be generated with -B option.
func (b *basicAuthProvider) Init(data *user.InitDataUserStorage) error {
	b.logger = data.Logger
	b.secret = data.Secret

	possibleFile := fmt.Sprintf("%s/_users", utils.GetDefaultConfigsDir())

	if !b.readFile(possibleFile) {
		b.logger.Warn("_users file is not found, going to use secret store only")
	}

	return nil
}

// Authorize validates default basic auth header against loaded file.
// If user is not found, falls back to system's secret store.
func (b *basicAuthProvider) Authorize(headers map[string][]string) (username string, err error) {
	var auth []string

	for k, v := range headers {
		if k != "Authorization" {
			continue
		}

		if 1 != len(v) {
			continue
		}

		auth = strings.SplitN(v[0], " ", 2)
		break
	}

	if 2 != len(auth) || "Basic" != auth[0] {
		b.logger.Warn("No Basic Auth header found")
		return "", errors.New("header not found")
	}

	payload, err := base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		b.logger.Warn("Failed to decode Basic Auth header")
		return "", errors.New("can't decode header")
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if 2 != len(pair) {
		b.logger.Warn("Corrupted Basic Auth header")
		return "", errors.New("wrong header")
	}

	pwd, ok := b.presetPasswords[pair[0]]
	if ok && bcrypt.CompareHashAndPassword([]byte(pwd), []byte(pair[1])) == nil {
		b.logger.Debug("Found user in _users file", "user", pair[0])
		return pair[0], nil
	}

	pwd, err = b.secret.Get(pair[0])
	if err == nil && pwd == pair[1] {
		b.logger.Debug("Found user in secret store", "user", pair[0])
		return pair[0], nil
	}

	b.logger.Warn("User is unauthorized", "user", pair[0])
	return "", errors.New("user not found")
}

// Reads htpsswds file.
func (b *basicAuthProvider) readFile(name string) bool {
	b.presetPasswords = make(map[string]string)
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		return false
	}

	lines := strings.Split(string(bytes), "\n")
	for _, v := range lines {
		v = strings.Trim(v, " ")
		if 0 == len(v) {
			continue
		}

		parts := strings.Split(v, ":")
		if 2 != len(parts) {
			continue
		}

		b.presetPasswords[parts[0]] = parts[1]
	}

	return true
}
