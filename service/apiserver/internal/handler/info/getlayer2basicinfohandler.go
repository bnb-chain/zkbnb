package info

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetLayer2BasicInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := info.NewGetLayer2BasicInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetLayer2BasicInfo()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
