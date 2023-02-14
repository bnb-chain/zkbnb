package info

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func SearchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqSearch
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := info.NewSearchLogic(r.Context(), svcCtx)
		resp, err := l.Search(&req)
		response.Handle(w, resp, err)
	}
}
