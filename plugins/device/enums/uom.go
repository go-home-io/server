//go:generate enumer -type=UOM -transform=snake -trimprefix=UOM -json -text -yaml

package enums

// UOM defines units of measure.
type UOM int

const (
	// UOMImperial defines imperial system.
	UOMImperial UOM = iota
	// UOMMetric defines metric system.
	UOMMetric
)
