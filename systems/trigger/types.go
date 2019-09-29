//go:generate enumer -type=triggerSystem -transform=kebab -trimprefix=trigger -json -text -yaml

package trigger

import (
	"github.com/gobwas/glob"
	"go-home.io/x/server/plugins/device/enums"
)

// triggerSystem describes known for trigger systems.
type triggerSystem int

const (
	// triggerDevice describes device command actions.
	triggerDevice triggerSystem = iota
	// triggerScript describes script invoke action.
	triggerScript
	// triggerNotification describes notification system action.
	triggerNotification
)

const (
	// Describes system entry
	system = "system"
)

// Device action.
type triggerActionDevice struct {
	Entity  string      `yaml:"entity" validate:"required"`
	Command string      `yaml:"command" validate:"required"`
	Args    interface{} `yaml:"args"`

	prepArgs   map[string]interface{}
	prepEntity glob.Glob
	cmd        enums.Command
}

// Notification action.
type triggerActionNotification struct {
	Entity  string `yaml:"entity" validate:"required"`
	Message string `yaml:"message"`

	prepEntity glob.Glob
}

// Trigger config.
type trigger struct {
	Actions   []map[string]interface{} `yaml:"actions" validate:"gt=0"`
	ActiveHrs string                   `yaml:"activeHrs"`
}
