package types

import (
	"fmt"
)

type Error interface {
	Error() string
	Code() int32
	RefineError(err ...interface{}) Error
}

func New(code int32, msg string) Error {
	return newError(code, msg)
}

type commonError struct {
	code    int32
	message string
}

func (e *commonError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.message)
}

func (e *commonError) Code() int32 {
	return e.code
}

func (e *commonError) RefineError(err ...interface{}) Error {
	return newError(e.Code(), e.message+fmt.Sprint(err...))
}

func newError(code int32, msg string) Error {
	return &commonError{
		code:    code,
		message: msg,
	}
}
