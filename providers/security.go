//go:generate enumer -type=SecVerb -transform=kebab -trimprefix=SecVerb -json -text -yaml
//go:generate enumer -type=SecSystem -transform=kebab -trimprefix=SecSystem -json -text -yaml

package providers

import (
	"github.com/gobwas/glob"
)

// ISecurityProvider defines security provider.
type ISecurityProvider interface {
	GetUser(map[string][]string) (IAuthenticatedUser, error)
}

// IAuthenticatedUser describes authenticated user.
type IAuthenticatedUser interface {
	Name() string
	DeviceGet(string) bool
	DeviceCommand(string) bool
	DeviceHistory(string) bool
	Workers() bool
	Entities() bool
	Logs() bool
}

// SecVerb describes allowed rules for the role.
type SecVerb int

const (
	// SecVerbAll describes all allowed operation rules.
	SecVerbAll SecVerb = iota
	// SecVerbGet describes get operation rule.
	SecVerbGet
	// SecVerbCommand describes execute command rule.
	SecVerbCommand
	// SecVerbHistory describes get history command rule
	SecVerbHistory
)

// SecSystem describes possible role's rule system.
type SecSystem int

const (
	// SecSystemAll describes all possible systems.
	SecSystemAll SecSystem = iota
	// SecSystemDevice describes devices' system.
	SecSystemDevice
	// SecSystemCore describes core components.
	SecSystemCore
)

// SecRoleRule has data, describing single security rule.
type SecRoleRule struct {
	System    string    `yaml:"system" validate:"required,oneof=* device core"`
	Resources []string  `yaml:"resources" validate:"unique,min=1"`
	Verbs     []SecVerb `yaml:"-"`
	StrVerb   []string  `yaml:"verbs" validate:"unique,min=1,oneof=* get command history"`
}

// SecRole has data, describing single security role.
type SecRole struct {
	Name  string        `yaml:"name" validate:"required"`
	Users []string      `yaml:"users" validate:"unique,min=1"`
	Rules []SecRoleRule `yaml:"rules" validate:"min=1"`
}

// BakedRule is a helper type with pre-compiled regexps.
type BakedRule struct {
	Resources []glob.Glob
	System    SecSystem
	Get       bool
	Command   bool
	History   bool
}
