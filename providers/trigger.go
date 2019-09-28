package providers

// ITriggerProvider defines events-trigger.
type ITriggerProvider interface {
	GetID() string
	GetLastTriggeredTime() int64
}
