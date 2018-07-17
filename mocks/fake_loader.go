package mocks

import (
	"github.com/go-home-io/server/providers"
	"errors"
)

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

func FakeNewPluginLoader(returnObj interface{}) providers.IPluginLoaderProvider {
	return &fakePluginLoader{
		returnObj: returnObj,
	}
}
