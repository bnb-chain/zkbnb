package block

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

func GetCurrentBlockHeightHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := block.NewGetCurrentBlockHeightLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentBlockHeight()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
