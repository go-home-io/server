package trigger

import (
	"reflect"

	"github.com/go-home-io/server/plugins/common"
)

// ITrigger defines event trigger interface.
type ITrigger interface {
	Init(*InitDataTrigger) error
}

// InitDataTrigger has data required for initializing a new trigger.
type InitDataTrigger struct {
	Logger    common.ILoggerProvider
	Secret    common.ISecretProvider
	FanOut    common.IFanOutProvider
	Triggered chan interface{}
}

// TypeTrigger is a syntax sugar around ITrigger type.
var TypeTrigger = reflect.TypeOf((*ITrigger)(nil)).Elem()
