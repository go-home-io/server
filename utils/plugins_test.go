package utils

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/go-home-io/server/mocks"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type wrongSettings struct {
}

// Fake settings
type fakeSettings struct {
	throw bool
}

func (f *fakeSettings) Validate() error {
	if f.throw {
		return errors.New("error")
	}

	return nil
}

// Fake plugin.
type iPlugin interface {
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

type pSuite struct {
	suite.Suite
	load *pluginLoader
}

func (p *pSuite) TearDownTest() {
	err := os.RemoveAll("./plugins")
	require.NoError(p.T(), err, "cleanup failed")
}

func (p *pSuite) SetupTest() {
	p.load = &pluginLoader{
		validator: NewValidator(mocks.FakeNewLogger(nil)),
	}
}

// Tests construct.
func (p *pSuite) TestConstruct() {
	loader := NewPluginLoader(&ConstructPluginLoader{
		Validator: nil,
	})

	_, err := loader.LoadPlugin(&providers.PluginLoadRequest{
		PluginProvider: "test",
	})

	assert.Error(p.T(), err)
}

// Tests regular load.
func (p *pSuite) TestRegularLoad() {
	_, err := p.load.loadPlugin(nil, func() (interface{}, interface{}, error) {
		return nil, nil, errors.New("test")
	})

	assert.Error(p.T(), err)
}

// Tests double load.
func (p *pSuite) TestDoubleLoad() {
	_, err := p.load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*wrongSettings)(nil)).Elem(),
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{}, &wrongSettings{}, nil
	})
	assert.Error(p.T(), err)

	_, err = p.load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*fakePlugin)(nil)).Elem(),
	}, func() (interface{}, interface{}, error) {
		return fakePlugin{}, &wrongSettings{}, nil
	})
	assert.Error(p.T(), err)
}

// Tests load with data.
func (p *pSuite) TestLoadWithData() {
	_, err := p.load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*iPlugin)(nil)).Elem(),
		InitData:     &device.InitDataDevice{},
		RawConfig:    []byte("data: test"),
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{}, &wrongSettings{}, nil
	})
	assert.Error(p.T(), err)
}

func (p *pSuite) TestThrow() {
	_, err := p.load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*iPlugin)(nil)).Elem(),
		InitData:     &device.InitDataDevice{},
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{throw: true}, &fakeSettings{}, nil
	})
	assert.Error(p.T(), err)
}

// Tests correct load.
func (p *pSuite) TestNoError() {
	_, err := p.load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*iPlugin)(nil)).Elem(),
		InitData:     &device.InitDataDevice{},
		RawConfig:    []byte("data: test"),
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{}, &fakeSettings{}, nil
	})
	assert.NoError(p.T(), err)
}

// Test proxy.
func (p *pSuite) TestProxy() {
	p.load.pluginsProxy = "http://wrong_url"
	_, err := p.load.loadPlugin(&providers.PluginLoadRequest{
		ExpectedType: reflect.TypeOf((*iPlugin)(nil)).Elem(),
		InitData:     &device.InitDataDevice{},
	}, func() (interface{}, interface{}, error) {
		return &fakePlugin{throw: true}, &fakeSettings{}, nil
	})
	assert.Error(p.T(), err)
}

// Tests plugin loading scenarios.
func TestErrorsScenarios(t *testing.T) {
	suite.Run(t, new(pSuite))
}
