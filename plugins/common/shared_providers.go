package common

// ISecretProvider defines secrets provider which will be passed to every plugin.
type ISecretProvider interface {
	Get(string) (string, error)
	Set(string, string) error
}

// ISettings describes interface used by every plugin.
// After loading plugin, go-home will invoke internal validation and then call this method.
type ISettings interface {
	Validate() error
}

// ILoggerProvider defines logger provider which will be passed to every plugin.
type ILoggerProvider interface {
	Debug(msg string, fields ...string)
	Info(msg string, fields ...string)
	Warn(msg string, fields ...string)
	Error(msg string, err error, fields ...string)
	Fatal(msg string, err error, fields ...string)
	Flush()
}
