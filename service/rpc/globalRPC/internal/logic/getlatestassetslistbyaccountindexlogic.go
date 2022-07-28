package logic

import (
	"context"

	"github.com/bnb-chain/zkbas/errorcode"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
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
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetLatestAssetsListByAccountIndexLogic) GetLatestAssetsListByAccountIndex(in *globalRPCProto.ReqGetLatestAssetsListByAccountIndex) (*globalRPCProto.RespGetLatestAssetsListByAccountIndex, error) {
	accountInfo, err := l.commglobalmap.GetLatestAccountInfoWithCache(l.ctx, int64(in.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] err:%v", err)
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrNotFound
		}
		return nil, errorcode.RpcErrInternal
	}
	resp := &globalRPCProto.RespGetLatestAssetsListByAccountIndex{
		ResultAssetsList: make([]*globalRPCProto.AssetResult, 0),
	}
	for assetID, asset := range accountInfo.AssetInfo {
		resp.ResultAssetsList = append(resp.ResultAssetsList, &globalRPCProto.AssetResult{
			AssetId:                  uint32(assetID),
			Balance:                  asset.Balance.String(),
			LpAmount:                 asset.LpAmount.String(),
			OfferCanceledOrFinalized: asset.OfferCanceledOrFinalized.String(),
		})
	}
	return resp, nil
}
