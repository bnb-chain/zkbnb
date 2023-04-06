package response

import (
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"reflect"
)

const CodeField = "Code"
const SuccessCode = 100

func Handle(w http.ResponseWriter, v interface{}, err error) {
	// If some error occurs when handling business, it is categorized into two kinds of errors.
	// For bizError, it is outputted with the http status code = 200, and the caller recognizes the code inside the bizError.
	// For sysError, it is outputted directly with the http status code = 500, and the caller does not care the error inside
	// the server and only processes the business based on the http status code, this case should rarely happen.
	if err != nil {
		switch err.(type) {
		case *types.SysError:
			httpx.Error(w, err)
		case *types.BizError:
			httpx.OkJson(w, err)
		default:
			httpx.Error(w, err)
		}
	} else {
		// If the server handles the business request successfully, here the result code is reset to the success code with
		// the reflection mechanism in golang. Of course, it is intended to do some validation and check in advance.
		reference := reflect.ValueOf(v).Elem()
		codeField := reference.FieldByName(CodeField)
		if codeField.IsValid() {
			codeField.SetInt(SuccessCode)
		}
		httpx.OkJson(w, v)
	}
}
