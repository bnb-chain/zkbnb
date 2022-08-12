package asset

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/asset"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAssetsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := asset.NewGetAssetsLogic(r.Context(), svcCtx)
		resp, err := l.GetAssets()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
