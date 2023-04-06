package info

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
)

func GetGasFeeAssetsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := info.NewGetGasFeeAssetsLogic(r.Context(), svcCtx)
		resp, err := l.GetGasFeeAssets()
		response.Handle(w, resp, err)
	}
}
