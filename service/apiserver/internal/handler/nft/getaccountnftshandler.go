package nft

import (
	"net/http"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/nft"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
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
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
