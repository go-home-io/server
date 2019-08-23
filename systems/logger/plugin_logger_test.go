package logger

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/mocks"
)

// Tests that every operation invoked correctly.
func TestPluginLogger(t *testing.T) {
	debug := false
	info := false
	warn := false
	err := false
	fatal := false

	ctor := &ConstructPluginLogger{
		Provider: "test",
		SystemLogger: mocks.FakeNewLogger(func(s string) {
			switch s {
			case "Debug":
				debug = true
			case "Info":
				info = true
			case "Warn":
				warn = true
			case "Error":
				err = true
			case "Fatal":
				fatal = true
			}
		}),
		System: "test",
	}

	l := NewPluginLogger(ctor)
	l.Debug("Debug")
	l.Info("Info")
	l.Warn("Warn")
	l.Error("Error", errors.New(""))
	l.Fatal("Fatal", errors.New(""))
	assert.False(t, l.GetSpecs().IsHistorySupported)
	assert.Nil(t, l.Query(nil))

	assert.True(t, debug, "debug")
	assert.True(t, info, "info")
	assert.True(t, warn, "warn")
	assert.True(t, err, "err")
	assert.True(t, fatal, "fatal")
}
