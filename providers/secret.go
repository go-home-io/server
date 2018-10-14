package providers

import "github.com/go-home-io/server/plugins/common"

// IInternalSecret defines internal secret provider wrapper.
type IInternalSecret interface {
	common.ISecretProvider
	UpdateLogger(pluginLogger common.ILoggerProvider)
}
