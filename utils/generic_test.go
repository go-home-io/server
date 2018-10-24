package utils

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/plugins/device/enums"
)

// Tests that we're returning current time.
func TestTimeNow(t *testing.T) {
	assert.Equal(t, time.Now().UTC().Unix(), TimeNow())
}

// Test no see.
func TestIsLongTimeNoSee(t *testing.T) {
	assert.False(t, IsLongTimeNoSee(TimeNow()-LongTimeNoSee+1))
	assert.True(t, IsLongTimeNoSee(TimeNow()-LongTimeNoSee-1))
}

// Tests correct device provider parsing.
func TestVerifyDeviceProvider(t *testing.T) {
	data := []struct {
		in  string
		out enums.DeviceType
	}{
		{
			in:  "hub/hue",
			out: enums.DevHub,
		},
		{
			in:  "light/zengge",
			out: enums.DevLight,
		},
		{
			in:  "wrong/device",
			out: enums.DevUnknown,
		},
		{
			in:  "wrong/device/provider",
			out: enums.DevUnknown,
		},
		{
			in:  "device/hub",
			out: enums.DevUnknown,
		},
		{
			in:  "hub",
			out: enums.DevUnknown,
		},
	}

	for _, v := range data {
		assert.Equal(t, v.out, VerifyDeviceProvider(v.in), v.in)
	}
}

// Tests config dir location.
func TestGetDefaultConfigsDir(t *testing.T) {
	ConfigDir = ""
	cd, _ := os.Getwd()
	assert.Equal(t, fmt.Sprintf("%s/configs", cd), GetDefaultConfigsDir(), "regular")

	ConfigDir = "testData"
	assert.Equal(t, ConfigDir, GetDefaultConfigsDir(), "changed")
}

// Tests devices name normalization.
func TestNormalizeDeviceName(t *testing.T) {
	data := map[string]string{
		"device 1": "device_1",
		"device-2": "device_2",
		"device.3": "device_3",
		"device%4": "device_4",
		"девайс$5": "девайс_5",
	}

	for k, v := range data {
		assert.Equal(t, v, NormalizeDeviceName(k), k)
	}
}
