package providers

import (
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/device/enums"
)

// IStorageProvider defines state history storage provider.
type IStorageProvider interface {
	Heartbeat(string)
	State(*common.MsgDeviceUpdate)
	History(string) map[enums.Property]map[int64]interface{}
}
