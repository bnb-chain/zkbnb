package transaction

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/transaction"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetBlockTxsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetBlockTxs
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := transaction.NewGetBlockTxsLogic(r.Context(), svcCtx)
		resp, err := l.GetBlockTxs(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
