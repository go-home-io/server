package utils

import (
	"errors"
	"github.com/go-home-io/server/plugins/device"
	"testing"
	"github.com/go-home-io/server/providers"
	"reflect"
	"github.com/go-home-io/server/mocks"
)

type wrongSettings struct {
}

type fakeSettings struct {
	throw bool
}

func (f *fakeSettings) Validate() error {
	if f.throw {
		return errors.New("error")
	}

	return nil
}

type IPlugin interface {
	device.IDevice
}

type fakePlugin struct {
	throw bool
}

func (f *fakePlugin) Init(*device.InitDataDevice) error {
	if f.throw {
		return errors.New("error")
	}

	return nil
}

func (*fakePlugin) Unload() {
}

func (*fakePlugin) GetName() string {
	return ""
}

func (*fakePlugin) GetSpec() *device.Spec {
	return nil
}

// Tests error while loading plugins.
func TestErrorsScenarios(t *testing.T) {
	loader := NewPluginLoader(&ConstructPluginLoader{
		Validator: nil,
	})

	_, err := loader.LoadPlugin(&providers.PluginLoadRequest{
		PluginProvider: "test",
	})

	if err == nil {
		t.Fail()
	}

	load := &pluginLoader{
		validator: NewValidator(mocks.FakeNewLogger(nil)),
	}
	_, err = load.loadPlugin(nil, func() (interface{}, interface{}, error) {
		return nil, nil, errors.New("test")
	})

	if err == nil {
		t.Fail()
	}

	_, err = load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*wrongSettings)(nil)).Elem(),
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{}, &wrongSettings{}, nil
	})

	_, err = load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*fakePlugin)(nil)).Elem(),
	}, func() (interface{}, interface{}, error) {
		return fakePlugin{}, &wrongSettings{}, nil
	})

	if err == nil {
		t.Fail()
	}

	_, err = load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*IPlugin)(nil)).Elem(),
		InitData:     &device.InitDataDevice{},
		RawConfig:    []byte("data: test"),
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{}, &wrongSettings{}, nil
	})

	if err == nil {
		t.Fail()
	}

	_, err = load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*IPlugin)(nil)).Elem(),
		InitData:     &device.InitDataDevice{},
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{throw: true}, &fakeSettings{}, nil
	})

	if err == nil {
		t.Fail()
	}

	_, err = load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*IPlugin)(nil)).Elem(),
		InitData:     &device.InitDataDevice{},
		RawConfig:    []byte("data: test"),
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{}, &fakeSettings{throw: true}, nil
	})

	if err == nil {
		t.Fail()
	}

	_, err = load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*IPlugin)(nil)).Elem(),
		InitData:     &device.InitDataDevice{},
		RawConfig:    []byte("data: test"),
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{}, &fakeSettings{}, nil
	})

	if err != nil {
		t.Fail()
	}
}
