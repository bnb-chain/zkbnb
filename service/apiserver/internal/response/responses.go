package response

import (
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func Handle(w http.ResponseWriter, v interface{}, err error) {
	if err != nil {
		switch err.(type) {
		case *types.SysError:
			httpx.Error(w, err)
		case *types.BizError:
			httpx.OkJson(w, err.Error())
		default:
			httpx.Error(w, err)
		}
	} else {
		httpx.OkJson(w, v)
	}
}
