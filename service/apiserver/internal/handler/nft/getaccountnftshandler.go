package nft

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/response"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/nft"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func GetAccountNftsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetAccountNfts
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := nft.NewGetAccountNftsLogic(r.Context(), svcCtx)
		resp, err := l.GetAccountNfts(&req)
		response.Handle(w, resp, err)
	}
}
