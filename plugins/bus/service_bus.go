package bus

import (
	"reflect"

	"github.com/go-home-io/server/plugins/common"
)

// IServiceBus defines service bus plugin interface.
type IServiceBus interface {
	Init(data InitDataServiceBus) error
	Subscribe(channel string, queue chan RawMessage) error
	Unsubscribe(channel string)
	Publish(channel string, messages ...interface{})
	Ping() error
}

// InitDataServiceBus has data required for initializing service bus plugin.
type InitDataServiceBus struct {
	Logger common.ILoggerProvider
	NodeID string
}

// RawMessage has un-parsed data from the service bus.
type RawMessage struct {
	Body []byte
}

// TypeServiceBus is a syntax sugar around IServiceBus type.
var TypeServiceBus = reflect.TypeOf((*IServiceBus)(nil)).Elem()
