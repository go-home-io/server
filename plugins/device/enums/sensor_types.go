//go:generate enumer -type=SensorType -transform=kebab -trimprefix=Sen -json -text -yaml

package enums

// SensorType defines type of the sensor.
type SensorType int

const (
	// SenGeneric describes generic sensor.
	SenGeneric SensorType = iota
	//SenMotion describes motion sensor.
	SenMotion
	// SenTemperature describes temperature sensor.
	SenTemperature
	// SenButton describes button sensor.
	SenButton
	// SenLock describes lock sensor.
	SenLock
)
