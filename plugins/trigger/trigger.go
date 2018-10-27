package trigger

import (
	"reflect"
	"time"

	"go-home.io/x/server/plugins/common"
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
	Timezone  *time.Location
	Triggered chan interface{}
}

// TypeTrigger is a syntax sugar around ITrigger type.
var TypeTrigger = reflect.TypeOf((*ITrigger)(nil)).Elem()
