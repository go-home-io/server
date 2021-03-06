package user

import (
	"reflect"

	"go-home.io/x/server/plugins/common"
)

// IUserStorage defines user store plugin interface.
type IUserStorage interface {
	Init(*InitDataUserStorage) error
	Authorize(headers map[string][]string) (username string, err error)
}

// InitDataUserStorage has data required for initializing a new user storage.
type InitDataUserStorage struct {
	Logger    common.ILoggerProvider
	Secret    common.ISecretProvider
	RawConfig []byte
}

// TypeUserStorage is a syntax sugar around IUserStorage type.
var TypeUserStorage = reflect.TypeOf((*IUserStorage)(nil)).Elem()
