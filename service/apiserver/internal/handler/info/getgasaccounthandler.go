package info

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetGasAccountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := info.NewGetGasAccountLogic(r.Context(), svcCtx)
		resp, err := l.GetGasAccount()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
