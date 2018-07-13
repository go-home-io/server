package providers

import "github.com/go-home-io/server/plugins/common"

// IValidatorProvider defines yaml structures validator logic.
type IValidatorProvider interface {
	SetLogger(logger common.ILoggerProvider)
	Validate(interface{}) bool
}
