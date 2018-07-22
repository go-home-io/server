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

// Test various property conversions.
func TestPropertyConversions(t *testing.T) {
	in := []string{"on", "battery_level"}
	prop := []Property{PropOn, PropBatteryLevel}
	out := []string{"On", "BatteryLevel"}

	for i, v := range in {
		p, _ := PropertyString(v)
		if p != prop[i] {
			t.Fail()
		}

		o := p.GetPropertyName()
		if o != out[i] {
			t.Fail()
		}
	}
}

// Test various command conversions.
func TestCommandConversions(t *testing.T) {
	in := []string{"on", "set-brightness"}
	cmd := []Command{CmdOn, CmdSetBrightness}
	out := []string{"On", "SetBrightness"}

	for i, v := range in {
		p, _ := CommandString(v)
		if p != cmd[i] {
			t.Fail()
		}

		o := p.GetCommandMethodName()
		if o != out[i] {
			t.Fail()
		}
	}
}

// Test helper IsPropertyAllowed.
func TestIsPropertyAllowed(t *testing.T) {
	AllowedProperties = map[DeviceType][]Property{
		DevLight: {PropOn, PropBrightness},
	}

	if !PropOn.IsPropertyAllowed(DevLight) {
		t.Fail()
	}

	if PropBatteryLevel.IsPropertyAllowed(DevLight) {
		t.Fail()
	}

	if PropBrightness.IsPropertyAllowed(DevHub) {
		t.Fail()
	}
}


// Tests helper SliceContainsDeviceType.
func TestSliceContainsDeviceType(t *testing.T){
	slice := []DeviceType{DevLight, DevHub}

	if SliceContainsDeviceType(slice, DevSwitch){
		t.Fail()
	}

	for _, v := range slice {
		if ! SliceContainsDeviceType(slice, v){
			t.Fail()
		}
	}
}