package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/go-home-io/server/plugins/common"
)

// Default console logger.
type consoleLogger struct {
}

// Debug prints debug level message.
func (p *consoleLogger) Debug(msg string, fields ...string) {
	output(msg, withFields(fields...), color.FgCyan)
}

// Info prints info level message.
func (p *consoleLogger) Info(msg string, fields ...string) {
	output(msg, withFields(fields...), color.FgGreen)
}

// Warn prints warning level message.
func (p *consoleLogger) Warn(msg string, fields ...string) {
	output(msg, withFields(fields...), color.FgYellow)
}

// Error prints error level message.
func (p *consoleLogger) Error(msg string, err error, fields ...string) {
	fields = append(fields, common.LogErrorToken, err.Error())
	output(msg, withFields(fields...), color.FgRed)
}

// Fatal prints fatal level message and exits.
func (p *consoleLogger) Fatal(msg string, err error, fields ...string) {
	fields = append(fields, common.LogErrorToken, err.Error())
	output(msg, withFields(fields...), color.FgRed)
	os.Exit(1)
}

// Flush don't needed for a console worker.
func (p *consoleLogger) Flush() {
}

// NewConsoleLogger constructs a new console worker.
func NewConsoleLogger() common.ILoggerProvider {
	return &consoleLogger{}
}

// Helper method to add generic fields to the output.
func withFields(fields ...string) map[string]string {
	fLen := len(fields)
	result := make(map[string]string, int(fLen/2))
	for ii := 0; ii < fLen; ii += 2 {
		if ii+1 >= fLen {
			break
		}

		result[fields[ii]] = fields[ii+1]
	}

	return result
}

// Prepares final string.
func output(msg string, fields map[string]string, c color.Attribute) {
	newM := fmt.Sprintf("%s   %s", time.Now().Local().Format(time.StampMilli), msg)

	for k, v := range fields {
		newM = fmt.Sprintf("%s\n          %s: %s", newM, k, v)
	}

	colorPrint(newM, c)
}

// Outputs final string.
func colorPrint(msg string, c color.Attribute) {
	msgC := color.New(c)
	//noinspection GoUnhandledErrorResult
	msgC.Println(msg) // nolint: gosec
}
