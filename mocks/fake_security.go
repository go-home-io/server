package mocks

import (
	"errors"

	"github.com/go-home-io/server/providers"
)

type fakeSecurity struct {
	allow bool
}

func (f *fakeSecurity) GetUser(map[string][]string) (*providers.AuthenticatedUser, error) {
	if f.allow {
		return &providers.AuthenticatedUser{
			Username: "test",
			Rules:    make(map[providers.SecSystem][]*providers.BakedRule),
		}, nil
	}

	return nil, errors.New("not found")
}

func FakeNewSecurityProvider(allow bool) *fakeSecurity {
	return &fakeSecurity{
		allow: allow,
	}
}
