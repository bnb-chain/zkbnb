package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestAssetInfoByAccountIndexAndAssetIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	account       account.AccountModel
	l2asset       l2asset.L2asset
	mempool       mempool.Mempool
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetLatestAssetInfoByAccountIndexAndAssetIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAssetInfoByAccountIndexAndAssetIdLogic {

	return &GetLatestAssetInfoByAccountIndexAndAssetIdLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		account:       account.New(svcCtx.Config),
		l2asset:       l2asset.New(svcCtx.Config),
		mempool:       mempool.New(svcCtx.Config),
		commglobalmap: commglobalmap.New(svcCtx.Config),
	}
}

func (l *GetLatestAssetInfoByAccountIndexAndAssetIdLogic) GetLatestAssetInfoByAccountIndexAndAssetId(in *globalRPCProto.ReqGetLatestAssetInfoByAccountIndexAndAssetId) (*globalRPCProto.RespGetLatestAssetInfoByAccountIndexAndAssetId, error) {
	if utils.CheckAccountIndex(in.AccountIndex) {
		logx.Error("[CheckAccountIndex] param:%v", in.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	accountInfo, err := l.commglobalmap.GetLatestAccountInfo(int64(in.AccountIndex))
	if err != nil {
		logx.Error("[GetLatestAccountInfo] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetLatestAssetInfoByAccountIndexAndAssetId{
		Balance: accountInfo.AssetInfo[int64(in.AssetId)].Balance.String(),
	}, nil
}
