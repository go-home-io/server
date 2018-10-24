package api

import (
	"reflect"

	"github.com/gobwas/glob"
	"github.com/gorilla/mux"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
)

// IExtendedAPI defines extended API plugin interface.
type IExtendedAPI interface {
	Init(*InitDataAPI) error
	Routes() []string
	Unload()
}

// InitDataAPI has data required for initializing a new API.
type InitDataAPI struct {
	InternalRootRouter *mux.Router
	ExternalAPIRouter  *mux.Router
	Logger             common.ILoggerProvider
	Secret             common.ISecretProvider
	FanOut             common.IFanOutProvider
	IsMaster           bool
	Communicator       IExtendedAPICommunicator
}

// IExtendedAPICommunicator defines
type IExtendedAPICommunicator interface {
	Subscribe(queue chan []byte) error
	Publish(messages ...IExtendedAPIMessage)
	InvokeDeviceCommand(deviceRegexp glob.Glob, cmd enums.Command, data map[string]interface{})
}

// IExtendedAPIMessage defines internal message between plugins
// running on master and worker.
type IExtendedAPIMessage interface {
	SetSendTime(int64)
}

// ExtendedAPIMessage is a base type for master <-> worker communications.
type ExtendedAPIMessage struct {
	SendTime int64 `json:"st"`
}

// SetSendTime sets base message SendTime.
func (e *ExtendedAPIMessage) SetSendTime(t int64) {
	e.SendTime = t
}

// TypeExtendedAPI is a syntax sugar around IExtendedAPI type.
var TypeExtendedAPI = reflect.TypeOf((*IExtendedAPI)(nil)).Elem()
