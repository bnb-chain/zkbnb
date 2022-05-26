package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l1asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetsListLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	l1asset l1asset.L1asset
	l2asset l2asset.L2asset
}

func NewGetAssetsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetsListLogic {
	return &GetAssetsListLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		l1asset: l1asset.New(svcCtx.Config),
		l2asset: l2asset.New(svcCtx.Config),
	}
}

func (l *GetAssetsListLogic) GetAssetsList(req *types.ReqGetAssetsList) (resp *types.RespGetAssetsList, err error) {
	l1AssetsInfo, err := l.l1asset.GetAssets()
	if err != nil {
		logx.Error("[GetAssets] err:%v", err)
		return nil, err
	}
	resp.L1Assets = make([]*types.L1Asset, 0)
	for _, l1Asset := range l1AssetsInfo {
		resp.L1Assets = append(resp.L1Assets, &types.L1Asset{
			L1AssetId:       uint16(l1Asset.AssetId),
			L1AssetAddr:     l1Asset.AssetAddress,
			L1AssetDecimals: uint8(l1Asset.Decimals),
			L1AssetSymbol:   l1Asset.AssetSymbol,
		})
	}
	l2AssetsInfo, err := l.l2asset.GetL2AssetsList()
	if err != nil {
		logx.Error("[GetAssets] err:%v", err)
		return nil, err
	}
	resp.L2Assets = make([]*types.L2Asset, 0)
	for _, l2AssetInfo := range l2AssetsInfo {
		resp.L2Assets = append(resp.L2Assets, &types.L2Asset{
			L2AssetId:       uint16(l2AssetInfo.L2AssetId),
			L2AssetName:     l2AssetInfo.L2AssetName,
			L2AssetDecimals: uint8(l2AssetInfo.L2Decimals),
			L2AssetSymbol:   l2AssetInfo.L2Symbol,
		})
	}
	return resp, nil
}
