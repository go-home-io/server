package providers

import (
	"reflect"

	"go-home.io/x/server/systems"
)

// IPluginLoaderProvider defines plugin loader provider logic.
type IPluginLoaderProvider interface {
	LoadPlugin(*PluginLoadRequest) (interface{}, error)
}

// PluginLoadRequest has data required for loading a new plugin.
type PluginLoadRequest struct {
	SystemType         systems.SystemType
	PluginProvider     string
	RawConfig          []byte
	InitData           interface{}
	ExpectedType       reflect.Type
	DownloadTimeoutSec int
}
