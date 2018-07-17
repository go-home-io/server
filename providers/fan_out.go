package providers

import "github.com/go-home-io/server/plugins/common"

// IInternalFanOutProvider defines internal interface for the fan-out channel.
// It extends regular IFanOutProvider which is available for plugins.
type IInternalFanOutProvider interface {
	common.IFanOutProvider

	ChannelInDeviceUpdates() chan *common.MsgDeviceUpdate
	SubscribeTriggerUpdates() (int64, chan string)
	UnSubscribeTriggerUpdates(int64)
	ChannelInTriggerUpdates() chan string
}
