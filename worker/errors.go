package worker

// ErrUnloadFailed defines failed unload error.
type ErrUnloadFailed struct {
}

// Error formats output.
func (*ErrUnloadFailed) Error() string {
	return "plugin unload failed"
}
