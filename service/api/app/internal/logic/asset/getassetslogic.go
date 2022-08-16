package asset

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAssetsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetsLogic {
	return &GetAssetsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAssetsLogic) GetAssets(req *types.ReqGetAll) (resp *types.Assets, err error) {
	total, err := l.svcCtx.MemCache.GetAssetTotalCountWithFallback(func() (interface{}, error) {
		return l.svcCtx.AssetModel.GetAssetsTotalCount()
	})
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	resp = &types.Assets{
		Assets: make([]*types.Asset, 0),
		Total:  uint32(total),
	}
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	assets, err := l.svcCtx.AssetModel.GetAssetsList(int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	resp.Assets = make([]*types.Asset, 0)
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, &types.Asset{
			AssetId:       asset.AssetId,
			AssetName:     asset.AssetName,
			AssetDecimals: asset.Decimals,
			AssetSymbol:   asset.AssetSymbol,
			AssetAddress:  asset.L1Address,
			IsGasAsset:    asset.IsGasAsset,
		})
	}
	return resp, nil
}
