package config

import (
	"testing"
)

func TestCorrectNames(t *testing.T) {
	names := []string{"1-2-3.yaml", "тест7.yaml", "-data.yml", "test.data.yaml"}

	for _, v := range names {
		if !IsValidConfigFileName(v) {
			t.Errorf("%s didn't pass", v)
			t.Fail()
		}
	}
}

func TestInCorrectNames(t *testing.T) {
	names := []string{"__", ".", "123.data", "test.yaml.data"}

	for _, v := range names {
		if IsValidConfigFileName(v) {
			t.Errorf("%s didn't pass", v)
			t.Fail()
		}
	}
}
