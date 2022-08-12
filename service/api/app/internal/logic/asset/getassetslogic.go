package asset

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GetAssetsLogic) GetAssets() (resp *types.RespGetAssets, err error) {
	assets, err := l.svcCtx.AssetModel.GetAssetsList()
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp.Assets = []*types.Asset{}
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
