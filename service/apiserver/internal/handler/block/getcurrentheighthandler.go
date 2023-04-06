package block

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/block"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
)

func GetCurrentHeightHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := block.NewGetCurrentHeightLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentHeight()
		response.Handle(w, resp, err)
	}
}
