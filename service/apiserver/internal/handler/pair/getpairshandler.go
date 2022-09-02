package pair

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/pair"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
)

func GetPairsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := pair.NewGetPairsLogic(r.Context(), svcCtx)
		resp, err := l.GetPairs()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
