package enums

import (
	"testing"
)

// Tests whether commands slice properly handles contains method.
func TestSliceCommandsContains(t *testing.T) {
	cmds := []Command{CmdOff, CmdOn, CmdToggle}

	if !SliceContainsCommand(cmds, CmdOn) {
		t.Fail()
	}
}

// Tests whether commands slice properly handles contains method.
func TestSliceCommandsNotContains(t *testing.T) {
	cmds := []Command{CmdOff, CmdOn, CmdToggle}

	if SliceContainsCommand(cmds, CmdSetTransitionTime) {
		t.Fail()
	}
}

// Tests whether allowed commands are calculated properly.
func TestCommandNotAllowed(t *testing.T) {
	AllowedCommands = map[DeviceType][]Command{
		DevHub: {CmdOn, CmdOff},
	}

	if CmdOff.IsCommandAllowed(DevLight) {
		t.Fail()
	}

	if CmdSetTransitionTime.IsCommandAllowed(DevHub) {
		t.Fail()
	}
}

// Tests whether allowed commands are calculated properly.
func TestCommandAllowed(t *testing.T) {
	AllowedCommands = map[DeviceType][]Command{
		DevHub: {CmdOn, CmdOff},
	}

	if ! CmdOff.IsCommandAllowed(DevHub) {
		t.Fail()
	}
}

// Tests reverse-transition from command enum into method name.
func TestCommandMethodNameConversion(t *testing.T) {
	if CmdOn.GetCommandMethodName() != "On" {
		t.Fail()
	}

	if CmdSetTransitionTime.GetCommandMethodName() != "SetTransitionTime" {
		t.Fail()
	}
}

// Tests whether properties slice properly handles contains method.
func TestSlicePropertyContains(t *testing.T) {
	props := []Property{PropScenes, PropBrightness, PropColor}

	if !SliceContainsProperty(props, PropBrightness) {
		t.Fail()
	}
}

// Tests whether properties slice properly handles contains method.
func TestSlicePropertyNotContains(t *testing.T) {
	props := []Property{PropScenes, PropBrightness, PropColor}

	if SliceContainsProperty(props, PropNumDevices) {
		t.Fail()
	}
}
