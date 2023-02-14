package info

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/info"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
)

func GetLayer2BasicInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := info.NewGetLayer2BasicInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetLayer2BasicInfo()
		response.Handle(w, resp, err)
	}
}
