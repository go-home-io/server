package utils

import (
	"testing"

	"github.com/go-home-io/server/mocks"
)

type testStruct struct {
	Percent uint8  `validate:"percent"`
	Port    int32  `validate:"port"`
	IpPort  string `validate:"ipv4port"`
}

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
		if !validator.Validate(v) {
			t.Fail()
		}
	}
}

func TestNotPointer(t *testing.T) {
	validator := NewValidator(mocks.FakeNewLogger(nil))

	if validator.Validate(testStruct{
		Percent: 0,
		Port:    8080,
		IpPort:  "127.0.0.1",
	}) {
		t.Fail()
	}
}

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
	for _, v := range in {
		if validator.Validate(v) {
			t.Fail()
		}
	}
}
