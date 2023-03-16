package transaction

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	types2 "github.com/bnb-chain/zkbnb/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/transaction"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func GetAccountTxsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetAccountTxs
		if err := httpx.Parse(r, &req); err != nil {
			bizErr := types2.AppErrInvalidParam.RefineError(err)
			response.Handle(w, nil, bizErr)
			return
		}

		l := transaction.NewGetAccountTxsLogic(r.Context(), svcCtx)
		resp, err := l.GetAccountTxs(&req)
		response.Handle(w, resp, err)
	}
}
