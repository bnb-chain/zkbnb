package info

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	table "github.com/bnb-chain/zkbas/common/model/asset"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetGasFeeAssetListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGasFeeAssetListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasFeeAssetListLogic {
	return &GetGasFeeAssetListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGasFeeAssetListLogic) GetGasFeeAssetList(req *types.ReqGetGasFeeAssetList) (*types.RespGetGasFeeAssetList, error) {
	assets, err := l.svcCtx.AssetModel.GetAssetsList()
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetGasFeeAssetList{
		Assets: make([]types.AssetInfo, 0),
	}
	for _, asset := range assets {
		if asset.IsGasAsset != table.IsGasAsset {
			continue
		}
		resp.Assets = append(resp.Assets, types.AssetInfo{
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
