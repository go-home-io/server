package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests whether commands slice properly handles contains method.
func TestSliceCommandsContains(t *testing.T) {
	cmds := []Command{CmdOff, CmdOn, CmdToggle}
	for _, v := range cmds {
		assert.True(t, SliceContainsCommand(cmds, v), v.String())
	}
}

// Tests whether commands slice properly handles contains method.
func TestSliceCommandsNotContains(t *testing.T) {
	cmds := []Command{CmdOff, CmdOn, CmdToggle}
	assert.False(t, SliceContainsCommand(cmds, CmdSetTransitionTime))
}

// Tests whether allowed commands are calculated properly.
func TestCommandNotAllowed(t *testing.T) {
	AllowedCommands = map[DeviceType][]Command{
		DevHub: {CmdOn, CmdOff},
	}

	assert.False(t, CmdOff.IsCommandAllowed(DevLight), CmdOff.String())
	assert.False(t, CmdSetTransitionTime.IsCommandAllowed(DevHub), CmdSetTransitionTime.String())
}

// Tests whether allowed commands are calculated properly.
func TestCommandAllowed(t *testing.T) {
	AllowedCommands = map[DeviceType][]Command{
		DevHub: {CmdOn, CmdOff},
	}

	assert.True(t, CmdOff.IsCommandAllowed(DevHub))
}

// Tests reverse-transition from command enum into method name.
func TestCommandMethodNameConversion(t *testing.T) {
	assert.Equal(t, "On", CmdOn.GetCommandMethodName(), CmdOn.String())
	assert.Equal(t, "SetTransitionTime",
		CmdSetTransitionTime.GetCommandMethodName(), CmdSetTransitionTime.String())
}

// Tests whether properties slice properly handles contains method.
func TestSlicePropertyContains(t *testing.T) {
	props := []Property{PropScenes, PropBrightness, PropColor}
	for _, v := range props {
		assert.True(t, SliceContainsProperty(props, v), v.String())
	}
}

// Tests whether properties slice properly handles contains method.
func TestSlicePropertyNotContains(t *testing.T) {
	props := []Property{PropScenes, PropBrightness, PropColor}
	assert.False(t, SliceContainsProperty(props, PropNumDevices))
}

// Test various property conversions.
func TestPropertyConversions(t *testing.T) {
	data := []struct {
		in   string
		prop Property
		out  string
	}{
		{
			in:   "on",
			prop: PropOn,
			out:  "On",
		},
		{
			in:   "battery_level",
			prop: PropBatteryLevel,
			out:  "BatteryLevel",
		},
	}

	for _, v := range data {
		p, err := PropertyString(v.in)
		assert.NoError(t, err, "property string %s", v.prop.String())
		assert.Equal(t, v.prop, p, "property string %s", v.prop.String())

		o := p.GetPropertyName()
		assert.Equal(t, v.out, o, "property name %s", v.prop.String())

	}
}

// Test various command conversions.
func TestCommandConversions(t *testing.T) {
	data := []struct {
		in  string
		cmd Command
		out string
	}{
		{
			in:  "on",
			cmd: CmdOn,
			out: "On",
		},
		{
			in:  "set-brightness",
			cmd: CmdSetBrightness,
			out: "SetBrightness",
		},
	}

	for _, v := range data {
		c, err := CommandString(v.in)
		assert.NoError(t, err, "command string %s", v.cmd.String())
		assert.Equal(t, v.cmd, c, "command string %s", v.cmd.String())

		o := c.GetCommandMethodName()
		assert.Equal(t, v.out, o, "command method name %s", v.cmd.String())
	}
}

// Test helper IsPropertyAllowed.
func TestIsPropertyAllowed(t *testing.T) {
	AllowedProperties = map[DeviceType][]Property{
		DevLight: {PropOn, PropBrightness},
	}

	assert.True(t, PropOn.IsPropertyAllowed(DevLight), PropOn.String())
	assert.True(t, PropBrightness.IsPropertyAllowed(DevLight), PropBrightness.String())
	assert.False(t, PropBatteryLevel.IsPropertyAllowed(DevLight), PropBatteryLevel.String())
}

// Tests helper SliceContainsDeviceType.
func TestSliceContainsDeviceType(t *testing.T) {
	slice := []DeviceType{DevLight, DevHub}

	assert.False(t, SliceContainsDeviceType(slice, DevSwitch), "not contains")

	for _, v := range slice {
		assert.True(t, SliceContainsDeviceType(slice, v), "contains %s", v.String())
	}
}
