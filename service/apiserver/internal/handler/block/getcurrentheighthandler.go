package block

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/block"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
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
