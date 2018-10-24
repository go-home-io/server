package providers

import "go-home.io/x/server/plugins/common"

// IValidatorProvider defines yaml structures validator logic.
type IValidatorProvider interface {
	SetLogger(logger common.ILoggerProvider)
	Validate(interface{}) bool
}
