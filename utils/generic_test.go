package utils

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-home-io/server/plugins/device/enums"
)

// Tests that we're returning current time.
func TestTimeNow(t *testing.T) {
	if TimeNow() != time.Now().UTC().Unix() {
		t.Fail()
	}
}

// Test no see.
func TestIsLongTimeNoSee(t *testing.T) {
	if IsLongTimeNoSee(TimeNow() - LongTimeNoSee + 1) {
		t.Fail()
	}

	if !IsLongTimeNoSee(TimeNow() - LongTimeNoSee - 1) {
		t.Fail()
	}
}

// Tests correct device provider parsing.
func TestVerifyDeviceProvider(t *testing.T) {
	in := []string{"hub/hue", "light/zengge", "wrong/device",
		"wrong/device/provider", "device/hub", "hub"}
	out := []enums.DeviceType{enums.DevHub, enums.DevLight, enums.DevUnknown,
		enums.DevUnknown, enums.DevUnknown, enums.DevUnknown}

	for i, v := range in {
		if VerifyDeviceProvider(v) != out[i] {
			t.Fail()
		}
	}
}

// Tests config dir location.
func TestGetDefaultConfigsDir(t *testing.T) {
	ConfigDir = ""
	cd, _ := os.Getwd()
	if fmt.Sprintf("%s/configs", cd) != GetDefaultConfigsDir() {
		t.Fail()
	}

	ConfigDir = "testData"
	if ConfigDir != GetDefaultConfigsDir() {
		t.Fail()
	}
}

// Tests devices name normalization.
func TestNormalizeDeviceName(t *testing.T) {
	in := []string{"device 1", "device-2", "device.3", "device%4", "девайс$5"}
	out := []string{"device_1", "device_2", "device_3", "device_4", "девайс_5"}

	for i, v := range in {
		if NormalizeDeviceName(v) != out[i] {
			t.Fail()
		}
	}
}
