package root

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/root"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
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
