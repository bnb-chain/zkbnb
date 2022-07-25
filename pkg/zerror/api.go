// error Custom error type in zecrey
// using method:
// err := error.New(10000, "Example error msg")
// fmt.Println("err:", err.Sprintf())
// error code in [10000,20000) represent business error
// error code in [20000,30000) represent logic layer error
// error code in [30000,40000) represent repo layer error
package zerror

type Error interface {
	Error() string
	Code() int32
	RefineError(err ...interface{}) *error
}

func New(code int32, msg string) Error {
	return new(code, msg)
}
