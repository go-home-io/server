package providers

// ILoadedProvider describes loaded provider.
type ILoadedProvider interface {
	Unload()
}

// IExtendedAPIProvider describes extended API wrapper.
type IExtendedAPIProvider interface {
	ILoadedProvider
	ID() string
	Routes() []string
}
