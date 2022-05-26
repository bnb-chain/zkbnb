package pair

import (
	"net/http"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/pair"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetSwapAmountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetSwapAmount
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := pair.NewGetSwapAmountLogic(r.Context(), svcCtx)
		resp, err := l.GetSwapAmount(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
