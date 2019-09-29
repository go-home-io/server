// Package notification contains notifications system plugin definitions.
package notification

import (
	"reflect"

	"go-home.io/x/server/plugins/common"
)

// INotification defines a notification system plugin.
type INotification interface {
	Init(*InitDataNotification) error
	Message(string) error
}

// InitDataNotification has data required for initializing a new plugin.
type InitDataNotification struct {
	Logger common.ILoggerProvider
	Secret common.ISecretProvider
}

// TypeNotification is a syntax sugar around INotification type.
var TypeNotification = reflect.TypeOf((*INotification)(nil)).Elem()
