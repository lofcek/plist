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
	Got interface{}
}

func NewUnexpectedTokenError(expected string, got interface{}) *UnexpectedTokenError {
	return &UnexpectedTokenError {
		expected, got,
	}
}

func (err *UnexpectedTokenError) Error() string {
	return fmt.Sprintf("Unexpected token: Expects %s, got %v",
		err.Expected,
		err.Got)
}

type CannotParseTypeError struct {
	Value reflect.Value
}

func (err *CannotParseTypeError) Error() string {
	return fmt.Sprintf("Cannot parse type %s", err.Value.Type().Name())
}
