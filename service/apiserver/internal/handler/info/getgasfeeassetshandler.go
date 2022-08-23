package info

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
)

func GetGasFeeAssetsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := info.NewGetGasFeeAssetsLogic(r.Context(), svcCtx)
		resp, err := l.GetGasFeeAssets()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
