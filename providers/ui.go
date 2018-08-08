package providers

// ILocationProvider defines UI location.
type ILocationProvider interface {
	ID() string
	Devices() []string
}
