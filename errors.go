package plist

import (
	"errors"
	"fmt"
	"reflect"
)

// Unmarshal could be done, only to pointer
var ErrMustBePointer = errors.New("Value must be a pointer.")

// UnexpectedTokenError appears when parse found something unexpected in data stream
type UnexpectedTokenError struct {
	Expected string
	Got      interface{}
	Offset   int64
}

func NewUnexpectedTokenError(expected string, got interface{}, offset int64) *UnexpectedTokenError {
	return &UnexpectedTokenError{
		expected, got, offset,
	}
}

func (err *UnexpectedTokenError) Error() string {
	return fmt.Sprintf("Unexpected token: Expects %s, got %#v (offset=%d)",
		err.Expected,
		err.Got,
		err.Offset)
}

type CannotParseTypeError struct {
	Value reflect.Value
}

func (err *CannotParseTypeError) Error() string {
	n := err.Value.Type().Name()
	if n == "" {
		n = "<<unnamed>>"
	}
	return fmt.Sprintf("Cannot parse type %s", n)
}
