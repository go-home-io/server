package trigger

// ErrNoActions defines no action.
type ErrNoActions struct {
}

// Error formats output.
func (e *ErrNoActions) Error() string {
	return "no action is defined"
}

// ErrInvalidActionConfig defines invalid action in config.
type ErrInvalidActionConfig struct {
}

// Error formats output.
func (*ErrInvalidActionConfig) Error() string {
	return "invalid action config"
}
