package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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
	assets, err := l.l2asset.GetL2AssetsList()
	if err != nil {
		logx.Errorf("[GetL2AssetsList] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetAssetsList{}
	resp.Assets = []*types.AssetInfo{}
	for _, asset := range assets {
		resp.Assets = append(resp.Assets, &types.AssetInfo{
			AssetId:       uint32(asset.AssetId),
			AssetName:     asset.AssetName,
			AssetDecimals: uint32(asset.Decimals),
			AssetSymbol:   asset.AssetSymbol,
			AssetAddress:  asset.L1Address,
		})
	}
	return resp, nil
}
