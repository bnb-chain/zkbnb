package info

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/info"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

func GetCurrencyPricesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := info.NewGetCurrencyPricesLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrencyPrices()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
