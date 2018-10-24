package providers

import "go-home.io/x/server/plugins/common"

// IInternalSecret defines internal secret provider wrapper.
type IInternalSecret interface {
	common.ISecretProvider
	UpdateLogger(pluginLogger common.ILoggerProvider)
}
