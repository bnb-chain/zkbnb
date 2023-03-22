package info

import (
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPlatformFeeRateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := info.NewGetPlatformFeeRateLogic(r.Context(), svcCtx)
		resp, err := l.GetPlatformFeeRate(true)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
