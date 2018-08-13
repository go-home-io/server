package providers

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
)

// IStorageProvider defines state history storage provider.
type IStorageProvider interface {
	Heartbeat(string)
	State(*common.MsgDeviceUpdate)
	History(string) map[enums.Property]map[int64]interface{}
}
