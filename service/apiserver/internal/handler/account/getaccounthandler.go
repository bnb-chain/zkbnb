package account

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/account"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func GetAccountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetAccount
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := account.NewGetAccountLogic(r.Context(), svcCtx)
		resp, err := l.GetAccount(&req)
		response.Handle(w, resp, err)
	}
}
