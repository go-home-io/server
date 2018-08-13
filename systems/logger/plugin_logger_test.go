package logger

import (
	"testing"

	"github.com/go-home-io/server/mocks"
	"github.com/pkg/errors"
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
	l.Flush()

	if !debug || !info || !warn || !err || !fatal {
		t.Fail()
	}
}
