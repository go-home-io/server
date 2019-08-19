//+build !release

package mocks

import (
	"errors"

	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
)

// IFakePlugin adds additional capabilities to a fake plugin loader.
type IFakePlugin interface {
	FakeInit(interface{})
}

type fakePluginLoader struct {
	returnObj interface{}
}

func (f *fakePluginLoader) LoadPlugin(r *providers.PluginLoadRequest) (interface{}, error) {
	if nil == f.returnObj {
		return nil, errors.New("not found")
	}

	i, ok := f.returnObj.(IFakePlugin)
	if ok {
		i.FakeInit(r.InitData)
	}

	return f.returnObj, nil
}

func (f *fakePluginLoader) UpdateLogger(common.ILoggerProvider) {
}

// FakeNewPluginLoader creates a new fake plugin loader.
func FakeNewPluginLoader(returnObj interface{}) providers.IPluginLoaderProvider {
	return &fakePluginLoader{
		returnObj: returnObj,
	}
}
