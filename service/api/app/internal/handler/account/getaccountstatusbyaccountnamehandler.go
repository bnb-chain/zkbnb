package account

import (
	"net/http"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAccountStatusByAccountNameHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetAccountStatusByAccountName
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := account.NewGetAccountStatusByAccountNameLogic(r.Context(), svcCtx)
		resp, err := l.GetAccountStatusByAccountName(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
