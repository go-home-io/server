//+build !release

package mocks

import (
	"errors"

	"go-home.io/x/server/plugins/user"
)

type fakeUserStorage struct {
	usr string
}

func (f *fakeUserStorage) Init(*user.InitDataUserStorage) error {
	return nil
}

func (f *fakeUserStorage) Authorize(headers map[string][]string) (username string, err error) {
	if "" == f.usr {
		return "", errors.New("not fount")
	}
	return f.usr, nil
}

// FakeNewUserStorage creates a fake user storage provider.
func FakeNewUserStorage(usr string) user.IUserStorage {
	return &fakeUserStorage{
		usr: usr,
	}
}
