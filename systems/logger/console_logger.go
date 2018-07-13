package logger

import (
	"os"

	"github.com/go-home-io/server/plugins/common"
	"github.com/sirupsen/logrus"
)

// Default console logger.
type consoleLogger struct {
}

// Debug prints debug level message.
func (p *consoleLogger) Debug(msg string, fields ...string) {
	logrus.WithFields(withFields(fields...)).Debug(msg)
}

// Info prints info level message.
func (p *consoleLogger) Info(msg string, fields ...string) {
	logrus.WithFields(withFields(fields...)).Info(msg)
}

// Warn prints warning level message.
func (p *consoleLogger) Warn(msg string, fields ...string) {
	logrus.WithFields(withFields(fields...)).Warn(msg)
}

// Error prints error level message.
func (p *consoleLogger) Error(msg string, err error, fields ...string) {
	fields = append(fields, common.LogErrorToken, err.Error())
	logrus.WithFields(withFields(fields...)).Error(msg)
}

// Fatal prints fatal level message and exits.
func (p *consoleLogger) Fatal(msg string, err error, fields ...string) {
	fields = append(fields, common.LogErrorToken, err.Error())
	logrus.WithFields(withFields(fields...)).Fatal(msg)
}

// Flush don't needed for a console worker.
func (p *consoleLogger) Flush() {
}

// NewConsoleLogger constructs a new console worker.
func NewConsoleLogger() common.ILoggerProvider {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	return &consoleLogger{}
}

// Helper method to add generic fields to the output.
func withFields(fields ...string) map[string]interface{} {
	fLen := len(fields)
	result := make(map[string]interface{}, int(fLen/2))
	for ii := 0; ii < fLen; ii += 2 {
		if ii+1 >= fLen {
			break
		}

		result[fields[ii]] = fields[ii+1]
	}

	return result
}
