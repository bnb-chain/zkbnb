package account

import (
	"net/http"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/logic/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAccountInfoByPubKeyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetAccountInfoByPubKey
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := account.NewGetAccountInfoByPubKeyLogic(r.Context(), svcCtx)
		resp, err := l.GetAccountInfoByPubKey(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
