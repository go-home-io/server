//go:generate enumer -type=SystemType -transform=kebab -trimprefix=Sys

package systems

// SystemType is an enum describing known system types.
type SystemType int

const (
	// SysGoHome describes server or worker system.
	SysGoHome SystemType = iota
	// SysLogger describes logger system.
	SysLogger
	// SysBus describes service bus system.
	SysBus
	// SysDevice describes device system.
	SysDevice
	// SysSecret describes secret store system.
	SysSecret
	// SysConfig describes config provider system.
	SysConfig
	// SysSecurity describes security provider system.
	SysSecurity
	// SysTrigger describes event trigger system.
	SysTrigger
	// SysAPI describes API extensions system.
	SysAPI
)
