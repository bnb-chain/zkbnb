package info

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
)

func GetProtocolRateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := info.NewGetProtocolRateLogic(r.Context(), svcCtx)
		resp, err := l.GetProtocolRate(true)
		response.Handle(w, resp, err)
	}
}
