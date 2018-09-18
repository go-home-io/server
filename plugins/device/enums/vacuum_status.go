//go:generate enumer -type=VacStatus -transform=snake -trimprefix=Vac -json -text -yaml

package enums

// VacStatus defines Vacuum device status.
type VacStatus int

const (
	// VacUnknown describes unknown status.
	VacUnknown VacStatus = iota
	// VacCleaning describes a vacuum in a cleaning stage.
	VacCleaning
	// VacPaused describes a vacuum in a paused state.
	VacPaused
	// VacDocked describes a vacuum in a docked/docking state.
	VacDocked
	// VacCharging describes a vacuum in a charging state.
	VacCharging
	// VacFull describes a vacuum in a full state.
	VacFull
)
