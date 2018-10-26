package mocks

import (
	"go-home.io/x/server/plugins/common"
	"go-home.io/x/server/providers"
)

type fakeValidator struct {
	success bool
}

func (f *fakeValidator) SetLogger(logger common.ILoggerProvider) {
}

func (f *fakeValidator) Validate(interface{}) bool {
	return f.success
}

// FakeNewValidator creates a new fake validation provider.
func FakeNewValidator(success bool) providers.IValidatorProvider {
	return &fakeValidator{
		success: success,
	}
}
