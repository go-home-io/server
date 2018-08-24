package providers

import "github.com/go-home-io/server/plugins/device/enums"

// IGroupProvider describes group device provider.
type IGroupProvider interface {
	ID() string
	Devices() []string
	InvokeCommand(enums.Command, map[string]interface{})
}
