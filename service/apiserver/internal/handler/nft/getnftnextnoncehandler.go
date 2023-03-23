package nft

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	zkbnbtypes "github.com/bnb-chain/zkbnb/types"
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/nft"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetNftNextNonceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetNftNextNonce
		if err := httpx.Parse(r, &req); err != nil {
			bizErr := zkbnbtypes.AppErrInvalidParam.RefineError(err)
			response.Handle(w, nil, bizErr)
			return
		}

		l := nft.NewGetNftNextNonceLogic(r.Context(), svcCtx)
		resp, err := l.GetNftNextNonce(&req)
		response.Handle(w, resp, err)
	}
}
