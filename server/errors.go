package server

import "fmt"

// ErrUnknownDevice defines unknown device error.
type ErrUnknownDevice struct {
	ID string
}

// Error formats output.
func (e *ErrUnknownDevice) Error() string {
	return fmt.Sprintf("device %s is unknown", e.ID)
}

// ErrUnknownCommand defines unknown command error.
type ErrUnknownCommand struct {
	Name string
}

// Error formats output.
func (e *ErrUnknownCommand) Error() string {
	return fmt.Sprintf("command %s is unknown", e.Name)
}

// ErrUnknownGroup defines unknown group error.
type ErrUnknownGroup struct {
	Name string
}

// Error formats output.
func (e *ErrUnknownGroup) Error() string {
	return fmt.Sprintf("group %s is unknown", e.Name)
}

// ErrUnsupportedCommand defines unsupported on device command error.
type ErrUnsupportedCommand struct {
	Name string
}

// Error formats output.
func (e *ErrUnsupportedCommand) Error() string {
	return fmt.Sprintf("command %s is not supported", e.Name)
}

// ErrBadRequest defines generic server error.
type ErrBadRequest struct {
}

// Error formats output.
func (e *ErrBadRequest) Error() string {
	return "bad request"
}
