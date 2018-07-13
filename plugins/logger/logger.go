// Package logger contains logger plugin definitions.
package logger

import "reflect"

// ILogger defines logger plugin interface.
type ILogger interface {
	Init() error
	Debug(msg string, fields ...string)
	Info(msg string, fields ...string)
	Warn(msg string, fields ...string)
	Error(msg string, fields ...string)
	Fatal(msg string, fields ...string)
	Flush()
}

// TypeLogger is a syntax sugar around ILogger type.
var TypeLogger = reflect.TypeOf((*ILogger)(nil)).Elem()
