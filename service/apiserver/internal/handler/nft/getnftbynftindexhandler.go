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

func GetNftByNftIndexHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetNft
		if err := httpx.Parse(r, &req); err != nil {
			bizErr := zkbnbtypes.AppErrInvalidParam.RefineError(err)
			response.Handle(w, nil, bizErr)
			return
		}

		l := nft.NewGetNftByNftIndexLogic(r.Context(), svcCtx)
		resp, err := l.GetNftByNftIndex(&req)
		response.Handle(w, resp, err)
	}
}
