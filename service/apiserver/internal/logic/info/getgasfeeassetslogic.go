package info

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbas/types"
)

type GetGasFeeAssetsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGasFeeAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasFeeAssetsLogic {
	return &GetGasFeeAssetsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGasFeeAssetsLogic) GetGasFeeAssets() (resp *types.GasFeeAssets, err error) {
	resp = &types.GasFeeAssets{Assets: make([]types.Asset, 0)}

	assets, err := l.svcCtx.AssetModel.GetGasAssets()
	if err != nil {
		return nil, types2.AppErrInternal
	}

	for _, asset := range assets {
		resp.Assets = append(resp.Assets, types.Asset{
			Id:         asset.AssetId,
			Name:       asset.AssetName,
			Decimals:   asset.Decimals,
			Symbol:     asset.AssetSymbol,
			Address:    asset.L1Address,
			IsGasAsset: asset.IsGasAsset,
		})
	}
	return resp, nil
}
