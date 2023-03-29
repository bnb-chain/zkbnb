package asset

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	zkbnbtypes "github.com/bnb-chain/zkbnb/types"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/asset"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func GetAssetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetAsset
		if err := httpx.Parse(r, &req); err != nil {
			bizErr := zkbnbtypes.AppErrInvalidParam.RefineError(err)
			response.Handle(w, nil, bizErr)
			return
		}

		l := asset.NewGetAssetLogic(r.Context(), svcCtx)
		resp, err := l.GetAsset(&req)
		response.Handle(w, resp, err)
	}
}
