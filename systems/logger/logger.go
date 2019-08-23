// Package logger provides wrapper around go-home logger implementation.
package logger

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/plugins/logger"
	"go-home.io/x/server/providers"
	"go-home.io/x/server/systems"
	"gopkg.in/yaml.v2"
)

// Logger provider wrapper implementation.
type provider struct {
	logger logger.ILogger
	nodeID string
}

// Logger settings.
type settings struct {
	Level string `yaml:"level"`
}

// ConstructLogger has data required for a new logger.
type ConstructLogger struct {
	LoggerType string
	Loader     providers.IPluginLoaderProvider
	RawConfig  []byte
	NodeID     string
	Secret     common.ISecretProvider
	SkipLevel  int
}

// NewLoggerProvider constructs a new logger.
func NewLoggerProvider(ctor *ConstructLogger) (common.ILoggerProvider, error) {
	prov := provider{
		nodeID: ctor.NodeID,
	}

	pluginLoadRequest := &providers.PluginLoadRequest{
		InitData: &logger.InitDataLogger{
			Secret:    ctor.Secret,
			Level:     getLogLevel(ctor.RawConfig),
			SkipLevel: ctor.SkipLevel,
		},
		RawConfig:      ctor.RawConfig,
		PluginProvider: ctor.LoggerType,
		SystemType:     systems.SysLogger,
		ExpectedType:   logger.TypeLogger,
	}

	i, err := ctor.Loader.LoadPlugin(pluginLoadRequest)
	if err != nil {
		return nil, errors.Wrap(err, "plugin load failed")
	}

	prov.logger = i.(logger.ILogger)

	return &prov, nil
}

// Debug sends debug level message.
func (p *provider) Debug(msg string, fields ...string) {
	p.logger.Debug(msg, p.prepareFields(fields...)...)
}

// Info sends info level message.
func (p *provider) Info(msg string, fields ...string) {
	p.logger.Info(msg, p.prepareFields(fields...)...)
}

// Warn sends warning level message.
func (p *provider) Warn(msg string, fields ...string) {
	p.logger.Warn(msg, p.prepareFields(fields...)...)
}

// Error sends error level message.
func (p *provider) Error(msg string, err error, fields ...string) {
	fields = append(fields, common.LogErrorToken, err.Error())
	p.logger.Error(msg, p.prepareFields(fields...)...)
}

// Fatal sends fatal level message and exits.
func (p *provider) Fatal(msg string, err error, fields ...string) {
	fields = append(fields, common.LogErrorToken, err.Error())
	p.logger.Fatal(msg, p.prepareFields(fields...)...)
}

// GetSpec returns plugin specs.
func (p *provider) GetSpecs() *common.LogSpecs {
	return p.logger.GetSpecs()
}

// Query performs logs search.
func (p *provider) Query(r *common.LogHistoryRequest) []*common.LogHistoryEntry {
	if nil == r {
		r = &common.LogHistoryRequest{
			FromUTC:  0,
			ToUTC:    0,
			LogLevel: "",
			System:   "",
			Provider: "",
			DeviceID: "",
		}
	}

	if 0 == r.ToUTC {
		r.ToUTC = time.Now().UTC().Unix()
	}

	return p.logger.Query(r)
}

// Extending logger fields with current node ID.
func (p *provider) prepareFields(fields ...string) []string {
	return append(fields, common.LogWorkerToken, p.nodeID)
}

// Gets log level from config.
func getLogLevel(rawConfig []byte) logger.LogLevel {
	s := &settings{}
	err := yaml.Unmarshal(rawConfig, s)
	if err != nil {
		return logger.Info
	}

	switch strings.ToLower(s.Level) {
	case "debug", "dbg":
		return logger.Debug
	case "warn", "warning":
		return logger.Warning
	case "error", "err":
		return logger.Error
	}

	return logger.Info
}
