package mocks

import (
	"github.com/go-home-io/server/providers"
	"errors"
)

type fakePluginLoader struct {
	returnObj interface{}
}

func (f *fakePluginLoader) LoadPlugin(*providers.PluginLoadRequest) (interface{}, error) {
	if nil == f.returnObj {
		return nil, errors.New("not found")
	}

	return f.returnObj, nil
}

func FakeNewPluginLoader(returnObj interface{}) providers.IPluginLoaderProvider {
	return &fakePluginLoader{
		returnObj: returnObj,
	}
}
