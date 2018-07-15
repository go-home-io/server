//+build !release

package mocks

import (
	"errors"

	"github.com/go-home-io/server/plugins/common"
)

type fakeSecret struct {
	data map[string]string
	isRO bool
}

func (f *fakeSecret) Get(name string) (string, error) {
	if nil == f.data {
		return "", errors.New("not found")
	}

	k, ok := f.data[name]
	if !ok {
		return "", errors.New("not found")
	}

	return k, nil
}

func (f *fakeSecret) Set(name string, value string) error {
	if f.isRO {
		return errors.New("error")
	}

	f.data[name] = value

	return nil
}

func FakeNewSecretStore(data map[string]string, readOnly bool) common.ISecretProvider {
	return &fakeSecret{
		data: data,
		isRO: readOnly,
	}
}
