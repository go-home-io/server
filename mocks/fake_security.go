//+build !release

package mocks

import (
	"errors"

	"go-home.io/x/server/providers"
)

type fakeSecurity struct {
	allow bool
}

func (f *fakeSecurity) GetUser(map[string][]string) (providers.IAuthenticatedUser, error) {
	if f.allow {
		return &fakeAuthenticatedUser{}, nil
	}

	return nil, errors.New("not found")
}

// FakeNewSecurityProvider creates a fake security provider.
func FakeNewSecurityProvider(allow bool) *fakeSecurity {
	return &fakeSecurity{
		allow: allow,
	}
}

// IFakeAuthenticatedUser adds additional capabilities to a fake user.
type IFakeAuthenticatedUser interface {
	SetAllow(bool)
}

type fakeAuthenticatedUser struct {
	allow bool
}

func (f *fakeAuthenticatedUser) Logs() bool {
	return f.allow
}

func (f *fakeAuthenticatedUser) SetAllow(allow bool) {
	f.allow = allow
}

func (*fakeAuthenticatedUser) Name() string {
	return "test"
}

func (f *fakeAuthenticatedUser) DeviceGet(string) bool {
	return f.allow
}

func (f *fakeAuthenticatedUser) DeviceCommand(string) bool {
	return f.allow
}

func (f *fakeAuthenticatedUser) DeviceHistory(string) bool {
	return f.allow
}

func (f *fakeAuthenticatedUser) Workers() bool {
	return f.allow
}

func (f *fakeAuthenticatedUser) Entities() bool {
	return f.allow
}
