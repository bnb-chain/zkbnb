// error Custom error type in zecrey
// using method:
// err := error.New(10000, "Example error msg")
// fmt.Println("err:", err.Sprintf())
package zerror

type Error interface {
	Error() string
	Code() int32
	RefineError(err string) *error
}

func New(code int32, msg string) Error {
	return new(code, msg)
}
