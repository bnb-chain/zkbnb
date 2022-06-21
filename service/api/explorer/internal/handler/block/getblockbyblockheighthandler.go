package block

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/logic/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetBlockByBlockHeightHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetBlockByBlockHeight
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := block.NewGetBlockByBlockHeightLogic(r.Context(), svcCtx)
		resp, err := l.GetBlockByBlockHeight(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
