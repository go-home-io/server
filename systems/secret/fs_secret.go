package secret

import (
	"github.com/go-home-io/server/plugins/secret"
	"github.com/go-home-io/server/plugins/common"
)

type fsSecret struct {
	location string
	logger   common.ILoggerProvider
}

func (s *fsSecret) Init(data *secret.InitDataSecret) error {
	return nil
}

func (s *fsSecret) Get(string) (string, error) {
	return "", nil
}

func (s *fsSecret) Set(name string, data string) error {
	return nil
}

func (s *fsSecret) UpdateLogger(provider common.ILoggerProvider) {
	s.logger = provider
}
