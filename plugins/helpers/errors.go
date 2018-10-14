package helpers

import "fmt"

// ErrBoolConvert defines boolean convert error.
type ErrBoolConvert struct {
}

// Error formats output.
func (*ErrBoolConvert) Error() string {
	return "error converting bool"
}

// ErrArgumentsMismatch defines incorrect number of arguments.
type ErrArgumentsMismatch struct {
	Count int
}

// Error formats output.
func (e *ErrArgumentsMismatch) Error() string {
	return fmt.Sprintf("arguments count missmatch, received: %d", e.Count)
}

// ErrWrongArgument defines wrong argument.
type ErrWrongArgument struct {
	Message string
}

// Error formats output.
func (e *ErrWrongArgument) Error() string {
	return e.Message
}

// ErrJqSyntax defines wrong jq syntax.
type ErrJqSyntax struct {
	Message string
}

// Error formats output.
func (e *ErrJqSyntax) Error() string {
	return e.Message
}
