// Package utils contains various helpers.
package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
	"github.com/mholt/archiver"
	"gopkg.in/yaml.v2"
)

const (
	// PluginEntryPointMethodName is the name of main plugin method.
	PluginEntryPointMethodName = "Load"
	// PluginInterfaceInitMethodName is the name of first initialization method.
	PluginInterfaceInitMethodName = "Init"
	// PluginCDNUrlFormat is format for bintray CDN.
	PluginCDNUrlFormat = "https://dl.bintray.com/go-home-io/%s/%s"
)

// Arch describes build architecture.
var Arch string = "amd64"

// Version describes build version.
var Version string = "0.0.5"

// ConstructPluginLoader contains params required for creating a new plugin loader instance.
type ConstructPluginLoader struct {
	PluginsFolder string
	PluginsProxy  string
	Validator     providers.IValidatorProvider
}

// Plugins loader.
type pluginLoader struct {
	pluginsFolder string
	pluginsProxy  string
	validator     providers.IValidatorProvider
	logger        *log.Logger

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
		pluginsProxy:  ctor.PluginsProxy,
		validator:     ctor.Validator,
		loadedPlugins: make(map[string]func() (interface{}, interface{}, error)),
		logger:        log.New(os.Stdout, "plugins", log.LstdFlags),
	}

	return &loader
}

// LoadPlugin loads requested plugin.
// Returns main interface implementation which should be casted to package interface.
func (l *pluginLoader) LoadPlugin(request *providers.PluginLoadRequest) (interface{}, error) {
	pKey := getPluginKey(request.SystemType, request.PluginProvider)
	if method, ok := l.loadedPlugins[pKey]; ok {
		l.logger.Printf("Loading plugin %s from cache", pKey)
		return l.loadPlugin(request, method)
	}

	l.logger.Printf("Loading plugin %s", pKey)
	fileName := l.getActualFileName(pKey)
	defer func() {
		if recover() != nil {
			os.Remove(fileName) // nolint: gosec
			l.logger.Fatal("Error opening plugin, corrupted?")
		}
	}()

	if _, err := os.Stat(fileName); err != nil {
		err = l.downloadFile(pKey, fileName)
		if err != nil {
			return nil, err
		}
	}

	p, err := plugin.Open(fileName)
	if err != nil {
		// We want to delete failed plugin
		//os.Remove(fileName) // nolint: gosec
		return nil, err
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

	if nil == request.RawConfig || nil == settingsObject {
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

// Gets actual plugin name.
func (l *pluginLoader) getActualFileName(pluginKey string) string {
	actualVersion := ""
	if "" != Version {
		actualVersion = fmt.Sprintf("-%s", Version)
	}

	return fmt.Sprintf("%s/%s%s.so", l.pluginsFolder, pluginKey, actualVersion)
}

// Downloads plugin from bintray CDN.
func (l *pluginLoader) downloadFile(pluginKey string, actualName string) error {
	name := strings.Replace(pluginKey, "/", "_", -1)
	name = fmt.Sprintf("%s-%s.so.tar.gz", name, Version)
	if "" != l.pluginsProxy {
		return l.downloadFileThroughProxy(name)
	}

	archName := fmt.Sprintf("%s.tar.gz", actualName)
	if _, err := os.Stat(archName); err != nil {
		l.logger.Println("Downloading " + name)
		err = os.MkdirAll(filepath.Dir(actualName), os.ModePerm)
		if err != nil {
			l.logger.Println("Failed to load " + name + ": " + err.Error())
			return err
		}
		out, err := os.Create(archName)
		if err != nil {
			l.logger.Println("Failed to load " + name + ": " + err.Error())
			return err
		}

		defer out.Close()
		downloadURL := fmt.Sprintf(PluginCDNUrlFormat, Arch, name)
		res, err := http.Get(downloadURL) // nolint: gosec
		if err != nil {
			os.Remove(archName) // nolint: gosec
			l.logger.Println("Failed to get " + downloadURL)
			return err
		}

		defer res.Body.Close()
		_, err = io.Copy(out, res.Body)
		if err != nil {
			l.logger.Println("Failed to save " + name + ": " + err.Error())
			os.Remove(archName) // nolint: gosec
			return err
		}
	}

	err := archiver.TarGz.Open(archName, filepath.Dir(actualName))
	if err != nil {
		l.logger.Println("Failed to un-archive " + name + ": " + err.Error())
		os.Remove(archName) // nolint: gosec
		return err
	}

	return nil
}

// Downloads plugin from bintray through proxy.
func (l *pluginLoader) downloadFileThroughProxy(name string) error {
	r, err := http.Get(fmt.Sprintf("%s/%s/%s", l.pluginsProxy, Arch, name))
	if err != nil || r.StatusCode != http.StatusOK {
		l.logger.Printf("Failed to download %s through proxy", name)
	}

	return nil
}
