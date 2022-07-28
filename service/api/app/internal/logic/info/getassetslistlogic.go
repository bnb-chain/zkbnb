package info

import (
	"context"

	"github.com/bnb-chain/zkbas/errorcode"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/l2asset"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAssetsListLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	l2asset l2asset.L2asset
}

func NewGetAssetsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetsListLogic {
	return &GetAssetsListLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		l2asset: l2asset.New(svcCtx),
	}
}

func (l *GetAssetsListLogic) GetAssetsList(req *types.ReqGetAssetsList) (*types.RespGetAssetsList, error) {
	assets, err := l.l2asset.GetL2AssetsList(l.ctx)
	if err != nil {
		logx.Errorf("[GetL2AssetsList] err:%v", err)
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetAssetsList{}
	resp.Assets = []*types.AssetInfo{}
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, &types.AssetInfo{
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
