package providers

import "go-home.io/x/server/plugins/device/enums"

// IGroupProvider describes group device provider.
type IGroupProvider interface {
	ID() string
	Devices() []string
	InvokeCommand(enums.Command, map[string]interface{})
}
