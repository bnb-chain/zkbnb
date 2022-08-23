package block

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/block"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetBlockHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetBlock
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := block.NewGetBlockLogic(r.Context(), svcCtx)
		resp, err := l.GetBlock(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
