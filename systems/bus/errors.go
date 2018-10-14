package bus

// ErrUnknownType defines an unknown message type error.
type ErrUnknownType struct {
}

// Error formats output.
func (*ErrUnknownType) Error() string {
	return "unknown message type"
}

// ErrOldMessage defines an old message error.
type ErrOldMessage struct {
}

// Error formats output.
func (*ErrOldMessage) Error() string {
	return "message is too old"
}

// ErrCorruptedMessage defines a corrupted message error.
type ErrCorruptedMessage struct {
}

// Error formats output.
func (*ErrCorruptedMessage) Error() string {
	return "failed to unmarshal bus message"
}
