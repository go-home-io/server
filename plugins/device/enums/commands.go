//go:generate enumer -type=Command -transform=kebab -trimprefix=Cmd -json -text -yaml

// Package enums contains various enumerations and rules for device plugins.
package enums

import "strings"

// Command describes enum with known device commands.
type Command int

const (
	// CmdOn describes turning on command.
	CmdOn Command = iota
	// CmdOff describes turning off command.
	CmdOff
	// CmdToggle describes toggling on-off status command.
	CmdToggle
	// CmdSetColor describes color changing on command.
	CmdSetColor
	// CmdSetScene describes turning on certain scene command.
	CmdSetScene
	// CmdSetBrightness describes changing brightness command.
	CmdSetBrightness
	// CmdSetTransitionTime describes transition time changing command.
	CmdSetTransitionTime
)

// AllowedCommands contains set of all possible allowed commands per device type.
var AllowedCommands = map[DeviceType][]Command{
	DevHub:   {},
	DevLight: {CmdToggle, CmdOn, CmdOff, CmdSetColor, CmdSetTransitionTime, CmdSetBrightness, CmdSetScene},
}

// SliceContainsCommand checks whether slice contains certain command.
func SliceContainsCommand(s []Command, e Command) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// IsCommandAllowed checks whether command is allowed for this device type.
func (i Command) IsCommandAllowed(deviceType DeviceType) bool {
	slice, ok := AllowedCommands[deviceType]
	if !ok {
		return false
	}

	return SliceContainsCommand(slice, i)
}

// GetCommandMethodName transforms string representation of the command into actual method name.
func (i Command) GetCommandMethodName() string {
	parts := strings.Split(i.String(), "-")
	result := ""
	for _, v := range parts {
		result += strings.Title(v)
	}

	return result
}
