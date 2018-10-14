package security

import "fmt"

// ErrNoHeader defines absence of basic auth header.
type ErrNoHeader struct {
}

// Error formats output.
func (*ErrNoHeader) Error() string {
	return "basic auth header not found"
}

// ErrIncorrectHeader defines incorrect basic auth header.
type ErrIncorrectHeader struct {
}

// Error formats output.
func (*ErrIncorrectHeader) Error() string {
	return "basic auth header could not be decoded"
}

// ErrCorruptedHeader defines corrupted basic auth header.
type ErrCorruptedHeader struct {
	Header string
}

// Error formats output.
func (e *ErrCorruptedHeader) Error() string {
	return fmt.Sprintf("basic auth header is incorrect: %s", e.Header)
}

// ErrUserNotFound defines unknown user.
type ErrUserNotFound struct {
	User string
}

// Error formats output.
func (e *ErrUserNotFound) Error() string {
	return fmt.Sprintf("user %s not found", e.User)
}
