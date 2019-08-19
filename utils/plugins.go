// Package utils contains various helpers.
package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"time"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"gopkg.in/yaml.v2"
)

const (
	// PluginEntryPointMethodName is the name of main plugin method.
	PluginEntryPointMethodName = "Load"
	// PluginInterfaceInitMethodName is the name of first initialization method.
	PluginInterfaceInitMethodName = "Init"
	// PluginCDNUrlFormat is format for bintray CDN.
	PluginCDNUrlFormat = "https://dl.bintray.com/go-home-io/%s/%s"
	// Logger system.
	logSystem = "plugins"
	// Log token
	logPluginToken = "plugin"
)

// Arch describes build architecture.
var Arch string

// Version describes build version.
var Version string

// ConstructPluginLoader contains params required for creating a new plugin loader instance.
type ConstructPluginLoader struct {
	PluginsFolder string
	PluginsProxy  string
	Validator     providers.IValidatorProvider
	Logger        common.ILoggerProvider
}

// Plugins loader.
type pluginLoader struct {
	pluginsFolder string
	pluginsProxy  string
	validator     providers.IValidatorProvider
	logger        common.ILoggerProvider

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
		logger:        ctor.Logger,
	}

	return &loader
}

// UpdateLogger updates internal logger
func (l *pluginLoader) UpdateLogger(logger common.ILoggerProvider) {
	l.logger = logger
}

// LoadPlugin loads requested plugin.
// Returns main interface implementation which should be casted to package interface.
//noinspection GoUnhandledErrorResult
func (l *pluginLoader) LoadPlugin(request *providers.PluginLoadRequest) (interface{}, error) {
	pKey := getPluginKey(request.SystemType, request.PluginProvider)
	if method, ok := l.loadedPlugins[pKey]; ok {
		l.logger.Info("Loading plugin from cache", common.LogSystemToken, logSystem, logPluginToken, pKey)
		return l.loadPlugin(request, method)
	}

	l.logger.Info("Loading plugin", common.LogSystemToken, logSystem, logPluginToken, pKey)
	fileName := l.getActualFileName(pKey)
	defer func() {
		if recover() != nil {
			os.Remove(fileName) // nolint: gosec, errcheck
			l.logger.Fatal("Error opening plugin, corrupted? Removing the .so file",
				errors.New("panic"), common.LogSystemToken, logSystem, logPluginToken, pKey)
		}
	}()

	if _, err := os.Stat(fileName); err != nil {
		err = l.downloadFile(pKey, fileName, time.Duration(request.DownloadTimeoutSec)*time.Second)
		if err != nil {
			return nil, errors.Wrap(err, "download failed")
		}
	}

	p, err := plugin.Open(fileName)
	if err != nil {
		// We want to delete failed plugin
		os.Remove(fileName) // nolint: gosec, errcheck
		return nil, errors.Wrap(err, "lib open failed")
	}

	LoadSymbol, err := p.Lookup(PluginEntryPointMethodName)
	if err != nil {
		return nil, &ErrNoEntryPoint{}
	}
	LoadMethod := LoadSymbol.(func() (interface{}, interface{}, error))
	if LoadMethod == nil {
		return nil, &ErrWrongSignature{}
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
		return nil, errors.Wrap(err, "load call failed")
	}

	if !reflect.TypeOf(pluginObject).AssignableTo(request.ExpectedType) {
		return nil, &ErrWrongInterface{}
	}

	if nil == request.RawConfig || nil == settingsObject {
		err = l.initPlugin(request, pluginObject)
		if err != nil {
			return nil, errors.Wrap(err, "init failed")
		}

		return pluginObject, nil
	}

	err = yaml.Unmarshal(request.RawConfig, settingsObject)
	if err != nil {
		return nil, errors.Wrap(err, "yaml un-marshal failed")
	}

	settingsInterface, ok := settingsObject.(common.ISettings)

	if !ok {
		return nil, &ErrWrongSettingsSignature{}
	}
	if !l.validator.Validate(settingsObject) {
		return nil, &ErrInvalidConfig{}
	}

	err = settingsInterface.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "settings validate failed")
	}

	err = l.initPlugin(request, pluginObject)
	if err != nil {
		return nil, errors.Wrap(err, "init plugin failed")
	}

	return pluginObject, nil
}

// Calling Init method of a plugin.
func (l *pluginLoader) initPlugin(request *providers.PluginLoadRequest, pluginObject interface{}) (err error) {
	method := reflect.ValueOf(pluginObject).MethodByName(PluginInterfaceInitMethodName)
	if !method.IsValid() {
		return &ErrNoInit{}
	}

	defer func() {
		if recover() != nil {
			err = &ErrInitPanic{}
		}
	}()

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
//noinspection GoUnhandledErrorResult
func (l *pluginLoader) downloadFile(pluginKey string, actualName string, timeout time.Duration) error {
	name := strings.Replace(pluginKey, "/", "_", -1)
	name = fmt.Sprintf("%s-%s.so.tar.gz", name, Version)
	if "" != l.pluginsProxy {
		r, err := l.downloadFileFromRemote(fmt.Sprintf("%s/%s/%s", l.pluginsProxy, Arch, name), timeout)
		if err == nil && r.Body != nil {
			defer r.Body.Close() // nolint: errcheck
		}
		return err
	}

	archName := fmt.Sprintf("%s.tar.gz", actualName)
	if _, err := os.Stat(archName); err != nil {
		l.logger.Info("Downloading file", common.LogSystemToken, logSystem, logPluginToken, pluginKey)
		err = os.MkdirAll(filepath.Dir(actualName), os.ModePerm)
		if err != nil {
			l.logger.Error("Failed to create a folder", err, common.LogSystemToken, logSystem,
				logPluginToken, pluginKey)
			return errors.Wrap(err, "mkdir failed")
		}
		out, err := os.Create(archName)
		if err != nil {
			l.logger.Error("Failed to create a file", err, common.LogSystemToken, logSystem,
				logPluginToken, pluginKey)
			return errors.Wrap(err, "archive create failed")
		}

		defer out.Close() // nolint: errcheck
		res, err := l.downloadFileFromRemote(fmt.Sprintf(PluginCDNUrlFormat, Arch, name), timeout)
		if err != nil {
			os.Remove(archName) // nolint: gosec, errcheck
			return err
		}

		defer res.Body.Close() // nolint: errcheck
		_, err = io.Copy(out, res.Body)
		if err != nil {
			l.logger.Error("Failed to write a file", err, common.LogSystemToken, logSystem,
				logPluginToken, pluginKey)
			os.Remove(archName) // nolint: gosec, errcheck
			return errors.Wrap(err, "copy file failed")
		}
	}

	err := archiver.TarGz.Open(archName, filepath.Dir(actualName))
	if err != nil {
		l.logger.Error("Failed to un-tar a file", err, common.LogSystemToken, logSystem,
			logPluginToken, pluginKey)
		os.Remove(archName) // nolint: gosec, errcheck
		return errors.Wrap(err, "un-tar failed")
	}

	return nil
}

// Downloads plugin from remote.
func (l *pluginLoader) downloadFileFromRemote(url string, timeout time.Duration) (*http.Response, error) {
	client := http.Client{
		Timeout: timeout,
	}

	r, err := client.Get(url)

	if err != nil || r.StatusCode != http.StatusOK {
		l.logger.Error("Failed to download a file", err, common.LogSystemToken, logSystem,
			common.LogURLToken, url)
		return nil, &ErrDownload{}
	}

	return r, nil
}
