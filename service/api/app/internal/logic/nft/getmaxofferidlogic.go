package nft

import (
	"context"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/core"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	maxOfferId := int64(0)
	var maxOfferIdAsset *commonAsset.AccountAsset
	for _, asset := range account.AssetInfo {
		if asset.OfferCanceledOrFinalized != nil && asset.OfferCanceledOrFinalized.Cmp(big.NewInt(0)) > 0 {
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
		maxOfferId = maxOfferIdAsset.AssetId * core.OfferPerAsset
		maxOfferId = maxOfferId + offerCancelOrFinalized
	}

	return &types.MaxOfferId{
		OfferId: uint64(maxOfferId),
	}, nil
}
