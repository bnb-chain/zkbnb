package root

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/root"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
)

func GetStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := root.NewGetStatusLogic(r.Context(), svcCtx)
		resp, err := l.GetStatus()
		response.Handle(w, resp, err)
	}
}
