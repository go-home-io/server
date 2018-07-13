// Package logger provides wrapper around go-home logger implementation.
package logger

import (
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/logger"
	"github.com/go-home-io/server/providers"
	"github.com/go-home-io/server/systems"
)

// Logger provider wrapper implementation.
type provider struct {
	logger logger.ILogger
	nodeID string
}

// ConstructLogger has data required for a new logger.
type ConstructLogger struct {
	LoggerType string
	Loader     providers.IPluginLoaderProvider
	RawConfig  []byte
	NodeID     string
}

// NewLoggerProvider constructs a new logger.
func NewLoggerProvider(ctor *ConstructLogger) (common.ILoggerProvider, error) {
	prov := provider{
		nodeID: ctor.NodeID,
	}

	pluginLoadRequest := &providers.PluginLoadRequest{
		InitData:       nil,
		RawConfig:      ctor.RawConfig,
		PluginProvider: ctor.LoggerType,
		SystemType:     systems.SysLogger,
		ExpectedType:   logger.TypeLogger,
	}

	i, err := ctor.Loader.LoadPlugin(pluginLoadRequest)
	if err != nil {
		return nil, err
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

// Flush flushes logger buffer if any.
func (p *provider) Flush() {
	p.logger.Flush()
}

// Extending logger fields with current node ID.
func (p *provider) prepareFields(fields ...string) []string {
	return append(fields, common.LogNodeToken, p.nodeID)
}
