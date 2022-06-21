package info

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/logic/info"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetLayer2BasicInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetLayer2BasicInfo
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := info.NewGetLayer2BasicInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetLayer2BasicInfo(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
