package transaction

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/transaction"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetTxByHashHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetTxByHash
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := transaction.NewGetTxByHashLogic(r.Context(), svcCtx)
		resp, err := l.GetTxByHash(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
