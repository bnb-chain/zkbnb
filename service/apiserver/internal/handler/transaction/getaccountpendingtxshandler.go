package transaction

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/transaction"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func GetAccountPendingTxsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetAccountPendingTxs
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := transaction.NewGetAccountPendingTxsLogic(r.Context(), svcCtx)
		resp, err := l.GetAccountPendingTxs(&req)
		response.Handle(w, resp, err)
	}
}
