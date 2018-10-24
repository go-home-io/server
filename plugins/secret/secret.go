// Package secret contains secrets store plugin definitions.
package secret

import (
	"reflect"

	"go-home.io/x/server/plugins/common"
)

// ISecret defines secrets store plugin interface.
type ISecret interface {
	common.ISecretProvider
	Init(*InitDataSecret) error
	UpdateLogger(common.ILoggerProvider)
}

// InitDataSecret has data required for initializing of a new secret store.
type InitDataSecret struct {
	Options map[string]string
	Logger  common.ILoggerProvider
}

// TypeSecret is a syntax sugar around ISecret type.
var TypeSecret = reflect.TypeOf((*ISecret)(nil)).Elem()
