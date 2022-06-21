package pair

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/pair"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPairInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetPairInfo
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pair.NewGetPairInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetPairInfo(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
