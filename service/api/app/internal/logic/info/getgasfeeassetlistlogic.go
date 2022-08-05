package info

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	table "github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/errorcode"
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
	assets, err := l.svcCtx.L2AssetModel.GetAssetsList()
	if err != nil {
		logx.Errorf("[GetL2AssetsList] err: %s", err.Error())
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
