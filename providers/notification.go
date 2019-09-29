package providers

// INotificationProvider defines notification provider.
type INotificationProvider interface {
	GetID() string
	Message(string)
}
