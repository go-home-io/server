package config

import (
	"path/filepath"
	"reflect"

	"github.com/go-home-io/server/plugins/common"
)

// IConfig defines config plugin interface.
type IConfig interface {
	Init(*InitDataConfig) error
	Load() chan []byte
}

// InitDataConfig has data required for initializing config reader plugin.
type InitDataConfig struct {
	Options map[string]string
	Logger  common.ILoggerProvider
	Secret  common.ISecretProvider
}

// IsValidConfigFileName checks whether config file name (if plugin deals with files) is valid.
func IsValidConfigFileName(name string) bool {
	name = filepath.Base(name)

	if name[0] == '_' {
		return false
	}

	name = filepath.Ext(name)
	return name == ".yaml" || name == ".yml"
}

// TypeConfig is a syntax sugar around IConfig type.
var TypeConfig = reflect.TypeOf((*IConfig)(nil)).Elem()
