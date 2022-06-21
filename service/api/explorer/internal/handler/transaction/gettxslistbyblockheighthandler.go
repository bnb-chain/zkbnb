package transaction

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/logic/transaction"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTxsListByBlockHeightHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetTxsListByBlockHeight
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := transaction.NewGetTxsListByBlockHeightLogic(r.Context(), svcCtx)
		resp, err := l.GetTxsListByBlockHeight(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
