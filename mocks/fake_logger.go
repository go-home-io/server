//+build !release

package mocks

// Fake logger
type fakeLogger struct {
	callback func(string)
}

// Prints debug level message.
func (p *fakeLogger) Debug(msg string, fields ...string) {
	if p.callback != nil {
		p.callback(msg)
	}
}

// Prints info level message.
func (p *fakeLogger) Info(msg string, fields ...string) {
	if p.callback != nil {
		p.callback(msg)
	}
}

// Prints warning level message.
func (p *fakeLogger) Warn(msg string, fields ...string) {
	if p.callback != nil {
		p.callback(msg)
	}
}

// Prints error level message.
func (p *fakeLogger) Error(msg string, err error, fields ...string) {
	if p.callback != nil {
		p.callback(msg)
	}
}

// Prints fatal level message and exits.
func (p *fakeLogger) Fatal(msg string, err error, fields ...string) {
	if p.callback != nil {
		p.callback(msg)
	}
}

// FakeNewLogger creates a fake logger provider.
func FakeNewLogger(callback func(string)) *fakeLogger {
	return &fakeLogger{
		callback: callback,
	}
}
