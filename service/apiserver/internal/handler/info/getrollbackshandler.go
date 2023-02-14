package info

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func GetRollbacksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetRollbacks
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := info.NewGetRollbacksLogic(r.Context(), svcCtx)
		resp, err := l.GetRollbacks(&req)
		response.Handle(w, resp, err)
	}
}
