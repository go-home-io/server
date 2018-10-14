package logger

import (
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// Tests proper fields allocation.
func TestCorrectFields(t *testing.T) {
	r := withFields("f1", "f1", "f2", "f2")
	assert.Equal(t, 2, len(r), "first")
	r = withFields("f1", "f1", "f2", "f2", "f3")
	assert.Equal(t, 2, len(r), "second")
}

func TestLogMethods(t *testing.T) {
	called := false
	monkey.Patch(colorPrint, func(_ string, _ color.Attribute) {
		called = true
	})
	defer monkey.UnpatchAll()

	c := &consoleLogger{}

	c.Debug("test")
	assert.True(t, called, "debug")
	called = false

	c.Info("test")
	assert.True(t, called, "info")
	called = false

	c.Warn("test")
	assert.True(t, called, "warn")
	called = false

	c.Error("test", errors.New("test"))
	assert.True(t, called, "error")
	called = false

	exitCalled := false
	monkey.Patch(os.Exit, func(code int) {
		assert.NotEqual(t, 0, code, "exit code")
		exitCalled = true
	})

	c.Fatal("test", errors.New("test"))
	assert.True(t, called, "fatal")
	assert.True(t, exitCalled, "fatal exit not called")

}
