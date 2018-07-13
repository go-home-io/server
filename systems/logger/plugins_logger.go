package logger

import (
	"github.com/go-home-io/server/plugins/common"
)

// Plugin logger implementation.
type pluginLogger struct {
	systemLogger common.ILoggerProvider
	pluginFields []string
}

// ConstructPluginLogger has data required for a new plugin logger.
type ConstructPluginLogger struct {
	SystemLogger common.ILoggerProvider
	System       string
	Provider     string
}

// NewPluginLogger constructs a new plugin logger.
// This is another level of abstraction which adds system type
// and provider name to the actual logger.
// This logger should be passed to actual plugin.
func NewPluginLogger(ctor *ConstructPluginLogger) common.ILoggerProvider {
	return &pluginLogger{
		systemLogger: ctor.SystemLogger,
		pluginFields: []string{common.LogSystemToken, ctor.System, common.LogProviderToken, ctor.Provider},
	}
}

// Debug sends debug level message.
func (l *pluginLogger) Debug(msg string, fields ...string) {
	l.systemLogger.Debug(msg, append(fields, l.pluginFields...)...)
}

// Info sends info level message.
func (l *pluginLogger) Info(msg string, fields ...string) {
	l.systemLogger.Info(msg, append(fields, l.pluginFields...)...)
}

// Warn sends warning level message.
func (l *pluginLogger) Warn(msg string, fields ...string) {
	l.systemLogger.Warn(msg, append(fields, l.pluginFields...)...)
}

// Error sends error level message.
func (l *pluginLogger) Error(msg string, err error, fields ...string) {
	l.systemLogger.Error(msg, err, append(fields, l.pluginFields...)...)
}

// Fatal sends fatal level message and exits.
func (l *pluginLogger) Fatal(msg string, err error, fields ...string) {
	l.systemLogger.Fatal(msg, err, append(fields, l.pluginFields...)...)
}

// Flush flushes logger buffer if any.
func (l *pluginLogger) Flush() {
	l.systemLogger.Flush()
}
