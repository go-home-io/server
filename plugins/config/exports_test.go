package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests allowed names.
func TestCorrectNames(t *testing.T) {
	names := []string{"1-2-3.yaml", "тест7.yaml", "-data.yml", "test.data.yaml"}

	for _, v := range names {
		assert.True(t, IsValidConfigFileName(v), "%s didn't pass", v)
	}
}

// Tests wrong names.
func TestInCorrectNames(t *testing.T) {
	names := []string{"__", ".", "123.data", "test.yaml.data"}

	for _, v := range names {
		assert.False(t, IsValidConfigFileName(v), "%s didn't pass", v)
	}
}
