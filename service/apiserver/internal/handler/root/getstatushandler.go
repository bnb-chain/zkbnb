package root

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/root"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := root.NewGetStatusLogic(r.Context(), svcCtx)
		resp, err := l.GetStatus()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
