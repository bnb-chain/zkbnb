package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAssetsListByAccountIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetLatestAssetsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAssetsListByAccountIndexLogic {
	return &GetLatestAssetsListByAccountIndexLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx.Config),
	}
}

func (l *GetLatestAssetsListByAccountIndexLogic) GetLatestAssetsListByAccountIndex(in *globalRPCProto.ReqGetLatestAssetsListByAccountIndex) (*globalRPCProto.RespGetLatestAssetsListByAccountIndex, error) {
	accountInfo, err := l.commglobalmap.GetLatestAccountInfo(int64(in.AccountIndex))
	if err != nil {
		logx.Error("[GetLatestAccountInfo] err:%v", err)
		return nil, err
	}
	resp := &globalRPCProto.RespGetLatestAssetsListByAccountIndex{}
	for assetID, asset := range accountInfo.AssetInfo {
		resp.ResultAssetsList = append(resp.ResultAssetsList, &globalRPCProto.AssetResult{
			AssetId: uint32(assetID),
			Balance: asset.Balance.String(),
		})
	}
	return resp, nil
}
