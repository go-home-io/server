package secret

import (
	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/secret"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"go-home.io/x/server/systems/logger"
)

const (
	// Config logs system value.
	logSystem = "secret"
)

// Secrets store wrapper implementation.
type provider struct {
	Secret secret.ISecret

	logger   common.ILoggerProvider
	provider string
}

// ConstructSecret has data required for a new secrets provider.
type ConstructSecret struct {
	Options      map[string]string
	PluginLogger common.ILoggerProvider
	Loader       providers.IPluginLoaderProvider
}

// NewSecretProvider constructs a new secrets store provider.
func NewSecretProvider(ctor *ConstructSecret) providers.IInternalSecret {
	if 0 == len(ctor.Options) {
		return returnFsProvider(ctor)
	}

	requesterProvider, ok := ctor.Options[common.LogProviderToken]
	if !ok || requesterProvider == "fs" {
		return returnFsProvider(ctor)
	}

	secretCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.PluginLogger,
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
		Secret:   pluginInterface.(secret.ISecret),
		logger:   secretLogger,
		provider: requesterProvider,
	}
}

// Get returns secret value or throws an error if it wan't found.
func (s *provider) Get(name string) (string, error) {
	s.logger.Debug("Requesting secret", common.LogSecretToken, name, common.LogSystemToken, logSystem)
	value, err := s.Secret.Get(name)
	if err != nil {
		s.logger.Error("Can't find requested secret", err,
			common.LogSecretToken, name, common.LogSystemToken, logSystem)
		return "", errors.Wrap(err, "secret not found")
	}

	return value, nil
}

// Set saves a new secret or updates existing one.
func (s *provider) Set(name string, data string) error {
	s.logger.Debug("Setting a new secret", common.LogSecretToken, name, common.LogSystemToken, logSystem)
	err := s.Secret.Set(name, data)

	if err != nil {
		s.logger.Error("Failed to add a new secret", err, common.LogSecretToken,
			name, common.LogSystemToken, logSystem)
		return errors.Wrap(err, "add secret failed")
	}
	return nil
}

// UpdateLogger updates a secret's provider logger.
// Since this component loads before main logger, we need to update it.
func (s *provider) UpdateLogger(pluginLogger common.ILoggerProvider) {
	secretLoggerCtor := &logger.ConstructPluginLogger{
		SystemLogger: pluginLogger,
		Provider:     s.provider,
		System:       systems.SysSecret.String(),
	}
	secretLogger := logger.NewPluginLogger(secretLoggerCtor)
	s.logger = secretLogger
	s.Secret.UpdateLogger(secretLogger)
}

// Helper to return default provider.
func returnFsProvider(ctor *ConstructSecret) *provider {
	prov := getFsProvider(ctor)
	return &provider{
		Secret:   prov,
		logger:   prov.logger,
		provider: "go-home",
	}
}

// Returning default file system config loader implementation.
// nolint: dupl
func getFsProvider(ctor *ConstructSecret) *fsSecret {
	secretLoggerCtor := &logger.ConstructPluginLogger{
		SystemLogger: ctor.PluginLogger,
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
	s.Init(data) // nolint: gosec, errcheck
	return s
}
