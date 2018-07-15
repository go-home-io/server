package mocks

import (
	"github.com/go-home-io/server/plugins/user"
	"errors"
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

func FakeNewUserStorage(usr string) user.IUserStorage {
	return &fakeUserStorage{
		usr: usr,
	}
}
