package common

import (
	"go-home.io/x/server/plugins/device/enums"
)

// ISecretProvider defines secrets provider which will be passed to every plugin.
type ISecretProvider interface {
	Get(string) (string, error)
	Set(name string, data string) error
}

// ISettings describes interface used by every plugin.
// After loading plugin, go-home will invoke internal validation and then call this method.
type ISettings interface {
	Validate() error
}

// LogSpecs defines logger specifications.
type LogSpecs struct {
	IsHistorySupported bool
}

// LogHistoryRequest defines history request.
// ToUTC will be always populated by the logger system.
// FromUTC might be 0.
// All missing fields will be empty string.
type LogHistoryRequest struct {
	FromUTC  int64  `json:"from"`
	ToUTC    int64  `json:"to"`
	LogLevel string `json:"log_level"`
	System   string `json:"system"`
	Provider string `json:"provider"`
	DeviceID string `json:"device_id"`
	WorkerID string `json:"worker"`
}

// LogHistoryEntry defines single log entry.
type LogHistoryEntry struct {
	UTCTimestamp int64             `json:"timestamp"`
	LogLevel     string            `json:"log_level"`
	System       string            `json:"system"`
	DeviceID     string            `json:"device_id"`
	Provider     string            `json:"provider"`
	WorkerID     string            `json:"worker"`
	Message      string            `json:"message"`
	Properties   map[string]string `json:"properties"`
}

// ILoggerProvider defines logger provider which will be passed to every plugin.
type ILoggerProvider interface {
	Debug(msg string, fields ...string)
	Info(msg string, fields ...string)
	Warn(msg string, fields ...string)
	Error(msg string, err error, fields ...string)
	Fatal(msg string, err error, fields ...string)
	GetSpecs() *LogSpecs
	Query(*LogHistoryRequest) []*LogHistoryEntry
}

// IPluginLoggerProvider defines additional method for adding extra fields.
type IPluginLoggerProvider interface {
	ILoggerProvider
	AddFields(map[string]string)
}

// MsgDeviceUpdate contains data with updates device's state.
type MsgDeviceUpdate struct {
	ID        string
	Name      string
	State     map[enums.Property]interface{}
	FirstSeen bool
	Type      enums.DeviceType
}

// IFanOutProvider defines interface used for distributing
// device updates even across all system.
type IFanOutProvider interface {
	SubscribeDeviceUpdates() (int64, chan *MsgDeviceUpdate)
	UnSubscribeDeviceUpdates(int64)
}
