package utils

import (
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/creasty/defaults"
	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/providers"
	"gopkg.in/go-playground/validator.v9"
)

// Validator implementation.
type validatorProvider struct {
	sync.Mutex
	validator *validator.Validate
	logger    common.ILoggerProvider
}

// NewValidator constructs a new validator.
func NewValidator(logger common.ILoggerProvider) providers.IValidatorProvider {
	val := &validatorProvider{
		logger: logger,
	}
	v := validator.New()
	loadNewValidator(v, logger, "percent", percent)
	loadNewValidator(v, logger, "port", port)
	loadNewValidator(v, logger, "ipv4port", ipv4port)

	val.validator = v
	return val
}

// SetLogger updates the logger.
// Since logger is loaded after first init, we need to re-assign it.
func (v *validatorProvider) SetLogger(logger common.ILoggerProvider) {
	v.logger = logger
}

// Validate performs validation of a config file.
func (v *validatorProvider) Validate(object interface{}) bool {
	v.Lock()
	defer v.Unlock()

	err := defaults.Set(object)

	if err != nil {
		v.logger.Error("Failed to set default field values", err)
		return false
	}

	err = v.validator.Struct(object)
	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			v.logger.Warn("Validation error", common.LogFieldToken, e.Field())
		}

		return false
	}
	return true
}

// Percent type validation.
func percent(fl validator.FieldLevel) bool {
	return fl.Field().Uint() <= 100
}

// Port type validation.
func port(fl validator.FieldLevel) bool {
	return isPort(fl.Field().Int())
}

// Ipv4:port type validation.
func ipv4port(fl validator.FieldLevel) bool {
	parts := strings.Split(fl.Field().String(), ":")
	if len(parts) > 2 {
		return false
	}

	ip := net.ParseIP(parts[0])
	if ip == nil || ip.To4() == nil {
		return false
	}

	if 2 == len(parts) {
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return false
		}

		return isPort(int64(port))
	}

	return true
}

// Validates whether value could be used as a port.
func isPort(val int64) bool {
	return val > 0 && val <= 65535
}

// Attempt to register a new validator
func loadNewValidator(validator *validator.Validate, logger common.ILoggerProvider,
	name string, function validator.Func) {
	if err := validator.RegisterValidation(name, function); err != nil {
		logger.Error("Failed to register validator type", err, "type", name)
	}
}
