// Package logger contains logger plugin definitions.
package logger

import (
	"reflect"

	"github.com/go-home-io/server/plugins/common"
)

// ILogger defines logger plugin interface.
type ILogger interface {
	Init(*InitDataLogger) error
	Debug(msg string, fields ...string)
	Info(msg string, fields ...string)
	Warn(msg string, fields ...string)
	Error(msg string, fields ...string)
	Fatal(msg string, fields ...string)
	Flush()
}

// LogLevel represents log level for the plugin.
type LogLevel int

const (
	// Info describes info log level.
	Info LogLevel = iota
	// Debug describes debug log level.
	Debug
	// Warning describes warn log level.
	Warning
	// Error describes error log level.
	Error
)

// InitDataLogger has data required for initializing a new logger.
type InitDataLogger struct {
	Secret common.ISecretProvider
	Level  LogLevel
}

// TypeLogger is a syntax sugar around ILogger type.
var TypeLogger = reflect.TypeOf((*ILogger)(nil)).Elem()
