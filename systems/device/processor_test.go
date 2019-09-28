package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go-home.io/x/server/plugins/device/enums"
)

var (
	processorsExists = []enums.DeviceType{enums.DevCamera}
)

// Tests that only certain device types returns processors.
func TestGetProcessors(t *testing.T) {
	for _, v := range enums.DeviceTypeValues() {
		if enums.SliceContainsDeviceType(processorsExists, v) {
			continue
		}

		p := newDeviceProcessor(v, "")
		assert.Nil(t, p, v.String())
	}
}
