// Package utils contains various helpers.
package utils

import (
	"errors"
	"fmt"
	"plugin"
	"reflect"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
	"gopkg.in/yaml.v2"
)

const (
	// PluginEntryPointMethodName is the name of main plugin method.
	PluginEntryPointMethodName = "Load"
	// PluginInterfaceInitMethodName is the name of first initialization method.
	PluginInterfaceInitMethodName = "Init"
)

// ConstructPluginLoader contains params required for creating a new plugin loader instance.
type ConstructPluginLoader struct {
	PluginsFolder string
	Validator     providers.IValidatorProvider
}

// Plugins loader.
type pluginLoader struct {
	pluginsFolder string
	validator     providers.IValidatorProvider

	loadedPlugins map[string]func() (interface{}, interface{}, error)
}

// NewPluginLoader creates a new plugins loader.
func NewPluginLoader(ctor *ConstructPluginLoader) providers.IPluginLoaderProvider {
	loc := ctor.PluginsFolder
	if "" == loc {
		loc = fmt.Sprintf("%s/plugins", GetCurrentWorkingDir())
	}

	loader := pluginLoader{
		pluginsFolder: loc,
		validator:     ctor.Validator,
		loadedPlugins: make(map[string]func() (interface{}, interface{}, error)),
	}

	return &loader
}

// LoadPlugin loads requested plugin.
// Returns main interface implementation which should be casted to package interface.
func (l *pluginLoader) LoadPlugin(request *providers.PluginLoadRequest) (interface{}, error) {
	pKey := getPluginKey(request.SystemType, request.PluginProvider)
	if method, ok := l.loadedPlugins[pKey]; ok {
		return l.loadPlugin(request, method)
	}
	p, err := plugin.Open(fmt.Sprintf("%s/%s.so", l.pluginsFolder, pKey))

	if err != nil {
		return nil, errors.New("didn't find plugin file")
	}

	LoadSymbol, err := p.Lookup(PluginEntryPointMethodName)
	if err != nil {
		return nil, errors.New("didn't find entry point")
	}
	LoadMethod := LoadSymbol.(func() (interface{}, interface{}, error))
	if LoadMethod == nil {
		return nil, errors.New("wrong entry point signature")
	}

	l.loadedPlugins[pKey] = LoadMethod

	return l.loadPlugin(request, LoadMethod)
}

// Internal plugin cache key.
func getPluginKey(subSystemType systems.SystemType, pluginName string) string {
	switch subSystemType {
	case systems.SysDevice:
		return pluginName
	default:
		return fmt.Sprintf("%s/%s", subSystemType.String(), pluginName)
	}
}

// Performs actual plugin load
func (l *pluginLoader) loadPlugin(request *providers.PluginLoadRequest,
	loadMethod func() (interface{}, interface{}, error)) (interface{}, error) {
	pluginObject, settingsObject, err := loadMethod()
	if err != nil {
		return nil, err
	}

	if !reflect.TypeOf(pluginObject).AssignableTo(request.ExpectedType) {
		return nil, errors.New("plugin doesn't implement requested interface")
	}

	if nil == request.RawConfig {
		err = l.initPlugin(request, pluginObject)
		if err != nil {
			return nil, err
		}

		return pluginObject, nil
	}

	err = yaml.Unmarshal(request.RawConfig, settingsObject)
	if err != nil {
		return nil, err
	}

	settingsInterface, ok := settingsObject.(common.ISettings)

	if !ok {
		return nil, errors.New("wrong settings signature")
	}
	if !l.validator.Validate(settingsObject) {
		return nil, errors.New("invalid config")
	}

	err = settingsInterface.Validate()
	if err != nil {
		return nil, err
	}

	err = l.initPlugin(request, pluginObject)
	if err != nil {
		return nil, err
	}

	return pluginObject, nil
}

// Calling Init method of a plugin.
func (l *pluginLoader) initPlugin(request *providers.PluginLoadRequest, pluginObject interface{}) error {
	method := reflect.ValueOf(pluginObject).MethodByName(PluginInterfaceInitMethodName)
	if !method.IsValid() {
		return errors.New("init method not found")
	}

	var results []reflect.Value

	if nil == request.InitData {
		results = method.Call(nil)
	} else {
		val := reflect.ValueOf(request.InitData)

		if reflect.ValueOf(request.InitData).Kind() != method.Type().In(0).Kind() {
			val = val.Elem()
		}

		rv := reflect.ValueOf(method.Type().In(0))
		for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			rv = rv.Elem()
		}

		results = method.Call([]reflect.Value{val})
	}

	if len(results) > 0 && results[0].Interface() != nil {
		return results[0].Interface().(error)
	}

	return nil
}
