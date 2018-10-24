package storage

import (
	"reflect"

	"go-home.io/x/server/plugins/common"
)

// IStorage defines state storage plugin interface.
type IStorage interface {
	Init(*InitDataStorage) error
	Heartbeat(string)
	State(string, map[string]interface{})
	History(string, int) map[string]map[int64]interface{}
}

// InitDataStorage has data required for initializing of a new state storage provider.
type InitDataStorage struct {
	Logger common.ILoggerProvider
	Secret common.ISecretProvider
}

// TypeStorage is a syntax sugar around IStorage type.
var TypeStorage = reflect.TypeOf((*IStorage)(nil)).Elem()
