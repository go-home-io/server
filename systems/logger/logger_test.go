package logger

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/mocks"
	"go-home.io/x/server/plugins/logger"
)

type fakePlugin struct {
	callback func(string)
}

func (f *fakePlugin) Init(*logger.InitDataLogger) error {
	return nil
}

func (f *fakePlugin) Debug(msg string, fields ...string) {
	f.callback(msg)
}

func (f *fakePlugin) Info(msg string, fields ...string) {
	f.callback(msg)
}

func (f *fakePlugin) Warn(msg string, fields ...string) {
	f.callback(msg)
}

func (f *fakePlugin) Error(msg string, fields ...string) {
	f.callback(msg)
}

func (f *fakePlugin) Fatal(msg string, fields ...string) {
	f.callback(msg)
}

// Tests error plugin loading.
func TestErrorLoad(t *testing.T) {
	ctor := &ConstructLogger{
		Secret:     mocks.FakeNewSecretStore(nil, true),
		Loader:     mocks.FakeNewPluginLoader(nil),
		RawConfig:  []byte(""),
		NodeID:     "",
		LoggerType: "test",
	}

	_, err := NewLoggerProvider(ctor)
	assert.Error(t, err)
}

// Tests correct methods invocations.
func TestLogger(t *testing.T) {
	debug := false
	info := false
	warn := false
	err1 := false
	fatal := false

	pl := &fakePlugin{
		callback: func(s string) {
			switch s {
			case "Debug":
				debug = true
			case "Info":
				info = true
			case "Warn":
				warn = true
			case "Error":
				err1 = true
			case "Fatal":
				fatal = true
			}
		},
	}

	ctor := &ConstructLogger{
		Secret:     mocks.FakeNewSecretStore(nil, true),
		Loader:     mocks.FakeNewPluginLoader(pl),
		RawConfig:  []byte("level: wrong"),
		NodeID:     "",
		LoggerType: "test",
	}

	l, err := NewLoggerProvider(ctor)
	require.NoError(t, err, "failed to load provider")

	l.Debug("Debug")
	l.Info("Info")
	l.Warn("Warn")
	l.Error("Error", errors.New(""))
	l.Fatal("Fatal", errors.New(""))

	assert.True(t, debug, "debug")
	assert.True(t, info, "info")
	assert.True(t, warn, "warn")
	assert.True(t, err1, "err1")
	assert.True(t, fatal, "fatal")
}

// Tests loading log level.
func TestLogLevel(t *testing.T) {
	in := []struct {
		In       string
		Expected logger.LogLevel
	}{
		{
			In:       "warning",
			Expected: logger.Warning,
		},
		{
			In:       "warn",
			Expected: logger.Warning,
		},
		{
			In:       "error",
			Expected: logger.Error,
		},
		{
			In:       "err",
			Expected: logger.Error,
		},
		{
			In:       "debug",
			Expected: logger.Debug,
		},
		{
			In:       "dbg",
			Expected: logger.Debug,
		},
		{
			In:       "info",
			Expected: logger.Info,
		},
		{
			In:       "incorrect",
			Expected: logger.Info,
		},
	}

	for _, v := range in {
		assert.Equal(t, v.Expected, getLogLevel([]byte(fmt.Sprintf("level: %s", v.In))), v.In)
	}
}
