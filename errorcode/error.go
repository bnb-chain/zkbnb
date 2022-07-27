package errorcode

import "fmt"

type Error interface {
	Error() string
	Code() int32
	RefineError(err ...interface{}) *error
}

func New(code int32, msg string) Error {
	return new(code, msg)
}

type error struct {
	code    int32
	message string
}

func (e *error) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.message)
}

func (e *error) Code() int32 {
	return e.code
}

func (e *error) RefineError(err ...interface{}) *error {
	return new(e.Code(), e.message+", "+fmt.Sprint(err...))
}

func new(code int32, msg string) *error {
	return &error{
		code:    code,
		message: msg,
	}
}
