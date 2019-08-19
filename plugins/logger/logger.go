// Package logger contains logger plugin definitions.
package logger

import (
	"reflect"
	"time"

	"go-home.io/x/server/plugins/common"
)

// ILogger defines logger plugin interface.
type ILogger interface {
	Init(*InitDataLogger) error
	GetSpecs() *LogSpecs
	Query(*LogHistoryRequest) []*LogHistoryEntry
	Debug(msg string, fields ...string)
	Info(msg string, fields ...string)
	Warn(msg string, fields ...string)
	Error(msg string, fields ...string)
	Fatal(msg string, fields ...string)
}

// LogSpecs defines logger specifications.
type LogSpecs struct {
	IsHistorySupported bool
}

// LogHistoryRequest defines history request.
type LogHistoryRequest struct {
	From     time.Time
	To       time.Time
	LogLevel string
	System   string
	Provider string
	DeviceID string
}

// LogHistoryEntry defines single log entry.
type LogHistoryEntry struct {
	UTCTimestamp int64
	LogLevel     string
	System       string
	DeviceID     string
	Provider     string
	Message      string
	Properties   map[string]string
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
	Secret    common.ISecretProvider
	Level     LogLevel
	SkipLevel int
}

// TypeLogger is a syntax sugar around ILogger type.
var TypeLogger = reflect.TypeOf((*ILogger)(nil)).Elem()
