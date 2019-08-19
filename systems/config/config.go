package config

import (
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/config"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/systems/logger"
)

// IConfigProvider provides capabilities for loading system configuration.
type IConfigProvider interface {
	Load() chan []byte
}

// Implements config provider wrapper.
type provider struct {
	Config IConfigProvider
}

// ConstructConfig contains data required for a new config provider.
type ConstructConfig struct {
	Options      map[string]string
	PluginLogger common.ILoggerProvider
	Loader       providers.IPluginLoaderProvider
	Secret       common.ISecretProvider
}

// NewConfigProvider constructs a new config provider.
func NewConfigProvider(ctor *ConstructConfig) IConfigProvider {
	requesterProvider, ok := ctor.Options[common.LogProviderToken]
	if !ok || requesterProvider == "fs" {
		return returnFsProvider(ctor)
	}

	configCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.PluginLogger,
		Provider:     requesterProvider,
		System:       systems.SysConfig.String(),
	}

	configLogger := logger.NewPluginLogger(configCtor)
	configLogger.Info("Loading config provider")

	pluginRequest := &providers.PluginLoadRequest{
		RawConfig: nil,
		InitData: &config.InitDataConfig{
			Options: ctor.Options,
			Logger:  configLogger,
			Secret:  ctor.Secret,
		},
		PluginProvider: requesterProvider,
		SystemType:     systems.SysConfig,
		ExpectedType:   config.TypeConfig,
	}
	pluginInterface, err := ctor.Loader.LoadPlugin(pluginRequest)

	if err != nil {
		configLogger.Error("Failed to load provider", err)
		return returnFsProvider(ctor)
	}

	return &provider{
		Config: pluginInterface.(config.IConfig),
	}
}

// Load invokes plugin provider and returns channel with chunks of config.
func (p *provider) Load() chan []byte {
	return p.Config.Load()
}

// Helper to return default provider.
func returnFsProvider(ctor *ConstructConfig) *provider {
	return &provider{
		Config: getFsProvider(ctor),
	}
}

// Returning default file system config loader implementation.
// nolint: dupl
func getFsProvider(ctor *ConstructConfig) *fsConfig {
	configLoggerCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.PluginLogger,
		Provider:     "go-home",
		System:       systems.SysConfig.String(),
	}

	configLogger := logger.NewPluginLogger(configLoggerCtor)
	configLogger.Info("Using default File System config loader")

	data := &config.InitDataConfig{
		Options: ctor.Options,
		Logger:  configLogger,
		Secret:  ctor.Secret,
	}

	cfg := &fsConfig{}
	cfg.Init(data) // nolint: gosec, errcheck
	return cfg
}
