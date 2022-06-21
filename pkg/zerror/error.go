package zerror

type zError struct {
	code    int32
	message string
}

func (e *zError) Error() string {
	return e.message
}

func (e *zError) Code() int32 {
	return e.code
}

func (e *zError) RefineError(err string) Error {
	e.message = e.message + err
	return e
}

func newError(code int32, msg string) Error {
	return &zError{
		code:    code,
		message: msg,
	}
}
