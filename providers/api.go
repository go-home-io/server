package providers

// IExtendedAPIProvider describes extended API wrapper.
type IExtendedAPIProvider interface {
	ID() string
	Routes() []string
	Unload()
}
