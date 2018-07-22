package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-home-io/server/plugins/device/enums"
)

// TimeNow returns epoch UTC.
func TimeNow() int64 {
	return time.Now().UTC().Unix()
}

// VerifyDeviceProvider transforms device provider from yaml config into actual type.
func VerifyDeviceProvider(configType string) enums.DeviceType {
	parts := strings.SplitN(configType, "/", 2)
	if len(parts) < 2 {
		return enums.DevUnknown
	}

	t, err := enums.DeviceTypeString(parts[0])
	if err != nil {
		return enums.DevUnknown
	}

	return t
}

// NormalizeDeviceName validates that final device name is correct.
func NormalizeDeviceName(raw string) string {
	raw = strings.ToLower(raw)
	replacer := strings.NewReplacer("%", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		";", "_",
		".", "_",
		"$", "_",
		"-", "_",
		" ", "_")
	return replacer.Replace(raw)
}

// GetCurrentWorkingDir returns application working directory.
func GetCurrentWorkingDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working dir")
	}

	return cwd
}

// GetDefaultConfigsDir returns default config directory which is cwd/configs.
func GetDefaultConfigsDir() string {
	if ConfigDir != "" {
		return ConfigDir
	}

	return fmt.Sprintf("%s/configs", GetCurrentWorkingDir())
}

// ConfigDir allows to re-write default config directory.
var ConfigDir = ""
