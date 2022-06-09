package info

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetsListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAssetsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetsListLogic {
	return &GetAssetsListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAssetsListLogic) GetAssetsList(req *types.ReqGetAssetsList) (resp *types.RespGetAssetsList, err error) {
	// l1Assets, err := l.svcCtx.L1AssetInfo.GetAssets()
	// if err != nil {
	// 	errInfo := fmt.Sprintf("[explorer.info.GetAssetsList]<=>[L1AssetInfoModel.GetAssets] %s", err.Error())
	// 	logx.Error(errInfo)
	// 	return packGetAssetsListResp(logic.FailStatus, "fail", errInfo, respResult), nil
	// }

	l2Assets, e := l.svcCtx.L2AssetInfo.GetL2AssetsList()
	if e != nil {
		err = fmt.Errorf("[explorer.info.GetAssetsList]<=>%v", e)
		l.Error(err)
		return
	}

	// l1 := make([]*types.L1Asset, 0)
	// for _, l1Asset := range l1Assets {
	// 	l1 = append(l1, &types.L1Asset{
	// 		ChainId:         int32(l1Asset.ChainId),
	// 		L1AssetId:       l1Asset.AssetId,
	// 		L1AssetAddr:     l1Asset.AssetAddress,
	// 		L1AssetDecimals: l1Asset.Decimals,
	// 		L1AssetSymbol:   l1Asset.AssetSymbol,
	// 	})
	// }
	for _, l2Asset := range l2Assets {
		resp.Assets = append(resp.Assets, &types.Asset{
			AssetId:       l2Asset.AssetId,
			AssetAddr:     l2Asset.AssetAddress,
			AssetDecimals: l2Asset.Decimals,
			AssetSymbol:   l2Asset.AssetSymbol,
		})
	}

	return
}
