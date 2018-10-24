package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/mocks"
)

type testStruct struct {
	Percent uint8  `validate:"percent"`
	Port    int32  `validate:"port"`
	IpPort  string `validate:"ipv4port"`
}

// Tests success validation
func TestSuccessValidation(t *testing.T) {
	in := []*testStruct{
		{
			Percent: 0,
			Port:    8080,
			IpPort:  "127.0.0.1",
		},
		{
			Percent: 100,
			Port:    65535,
			IpPort:  "10.0.0.100:8080",
		},
	}

	validator := NewValidator(mocks.FakeNewLogger(nil))
	for _, v := range in {
		assert.True(t, validator.Validate(v), v.IpPort)
	}
}

// Tests validation without pointer.
func TestNotPointer(t *testing.T) {
	validator := NewValidator(mocks.FakeNewLogger(nil))
	d := testStruct{
		Percent: 0,
		Port:    8080,
		IpPort:  "127.0.0.1",
	}

	assert.False(t, validator.Validate(d))
}

// Tests incorrect data
func TestFailedValidation(t *testing.T) {
	in := []*testStruct{
		{
			Percent: 120,
		},
		{
			Port: 100000,
		},
		{
			IpPort: "10.0.0.100:test",
		},
		{
			IpPort: "10.0.0.100:22:123",
		},
	}

	validator := NewValidator(mocks.FakeNewLogger(nil))
	for k, v := range in {
		assert.False(t, validator.Validate(v), "%d", k)
	}
}
