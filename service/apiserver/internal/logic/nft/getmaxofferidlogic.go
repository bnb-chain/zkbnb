package nft

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/core/executor"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetMaxOfferIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMaxOfferIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMaxOfferIdLogic {
	return &GetMaxOfferIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMaxOfferIdLogic) GetMaxOfferId(req *types.ReqGetMaxOfferId) (resp *types.MaxOfferId, err error) {
	account, err := l.svcCtx.StateFetcher.GetLatestAccount(int64(req.AccountIndex))
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrAccountNotFound
		}
		return nil, types2.AppErrInternal
	}

	maxOfferId := int64(0)
	var maxOfferIdAsset *types2.AccountAsset
	for _, asset := range account.AssetInfo {
		if asset.OfferCanceledOrFinalized != nil && asset.OfferCanceledOrFinalized.Cmp(types2.ZeroBigInt) > 0 {
			if maxOfferIdAsset == nil || asset.AssetId > maxOfferIdAsset.AssetId {
				maxOfferIdAsset = asset
			}
		}
	}

	if maxOfferIdAsset != nil {
		offerCancelOrFinalized := int64(0)
		bitLen := maxOfferIdAsset.OfferCanceledOrFinalized.BitLen()
		for i := bitLen; i >= 0; i-- {
			if maxOfferIdAsset.OfferCanceledOrFinalized.Bit(i) == 1 {
				offerCancelOrFinalized = int64(i)
				break
			}
		}
		maxOfferId = maxOfferIdAsset.AssetId * executor.OfferPerAsset
		maxOfferId = maxOfferId + offerCancelOrFinalized
	}

	return &types.MaxOfferId{
		OfferId: uint64(maxOfferId),
	}, nil
}
