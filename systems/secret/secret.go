package secret

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/secret"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
	"github.com/go-home-io/server/systems/logger"
)

const (
	// Config logs system value.
	logSystem = "secret"
)

// Secrets store wrapper implementation.
type provider struct {
	Secret secret.ISecret

	logger common.ILoggerProvider
}

// ConstructSecret has data required for a new secrets provider.
type ConstructSecret struct {
	Options map[string]string
	Logger  common.ILoggerProvider
	Loader  providers.IPluginLoaderProvider
}

// NewSecretProvider constructs a new secrets store provider.
func NewSecretProvider(ctor *ConstructSecret) common.ISecretProvider {
	if 0 == len(ctor.Options) {
		return returnFsProvider(ctor)
	}

	requesterProvider, ok := ctor.Options[common.LogProviderToken]
	if !ok || requesterProvider == "fs" {
		return returnFsProvider(ctor)
	}

	secretCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.Logger,
		Provider:     requesterProvider,
		System:       systems.SysSecret.String(),
	}

	secretLogger := logger.NewPluginLogger(secretCtor)
	secretLogger.Info("Loading secret provider")

	pluginRequest := &providers.PluginLoadRequest{
		RawConfig: nil,
		InitData: &secret.InitDataSecret{
			Options: ctor.Options,
			Logger:  secretLogger,
		},
		PluginProvider: requesterProvider,
		SystemType:     systems.SysSecret,
		ExpectedType:   secret.TypeSecret,
	}
	pluginInterface, err := ctor.Loader.LoadPlugin(pluginRequest)

	if err != nil {
		secretLogger.Error("Failed to load provider", err)
		return returnFsProvider(ctor)
	}

	return &provider{
		Secret: pluginInterface.(secret.ISecret),
		logger: secretLogger,
	}
}

// Get returns secret value or throws an error if it wan't found.
func (s *provider) Get(name string) (string, error) {
	s.logger.Debug("Requesting secret", common.LogSecretToken, name, common.LogSystemToken, logSystem)
	value, err := s.Secret.Get(name)
	if err != nil {
		s.logger.Error("Can't find requested secret", err, common.LogSecretToken, name, common.LogSystemToken, logSystem)
		return "", err
	}

	return value, nil
}

// Set saves a new secret or updates existing one.
func (s *provider) Set(name string, data string) error {
	s.logger.Debug("Setting a new secret", common.LogSecretToken, name, common.LogSystemToken, logSystem)
	err := s.Secret.Set(name, data)

	if err != nil {
		s.logger.Error("Failed to add a new secret", err, common.LogSecretToken, name, common.LogSystemToken, logSystem)
		return err
	}
	return nil
}

// UpdateLogger updates a secret's provider logger.
// Since this component loads before main logger, we need to update it.
func (s *provider) UpdateLogger(provider common.ILoggerProvider) {
	s.logger = provider
	s.Secret.UpdateLogger(provider)
}

// Helper to return default provider.
func returnFsProvider(ctor *ConstructSecret) *provider {
	return &provider{
		Secret: getFsProvider(ctor),
	}
}

// Returning default file system config loader implementation.
// nolint: dupl
func getFsProvider(ctor *ConstructSecret) *fsSecret {
	secretLoggerCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.Logger,
		Provider:     "go-home",
		System:       systems.SysSecret.String(),
	}

	secretLogger := logger.NewPluginLogger(secretLoggerCtor)
	secretLogger.Info("Using default File System secret")

	data := &secret.InitDataSecret{
		Options: ctor.Options,
		Logger:  secretLogger,
	}

	s := &fsSecret{}
	s.Init(data)
	return s
}
