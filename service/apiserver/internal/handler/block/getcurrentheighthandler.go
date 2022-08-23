package block

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/block"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetCurrentHeightHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := block.NewGetCurrentHeightLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentHeight()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
