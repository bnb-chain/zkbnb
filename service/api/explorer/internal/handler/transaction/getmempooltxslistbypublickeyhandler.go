package transaction

import (
	"net/http"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/logic/transaction"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetMempoolTxsListByPublicKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetMempoolTxsListByPublicKey
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := transaction.NewGetMempoolTxsListByPublicKeyLogic(r.Context(), svcCtx)
		resp, err := l.GetMempoolTxsListByPublicKey(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
