package logic

import (
	"context"

	"github.com/bnb-chain/zkbas/errorcode"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type GetLatestAccountInfoByAccountIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetLatestAccountInfoByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestAccountInfoByAccountIndexLogic {
	return &GetLatestAccountInfoByAccountIndexLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetLatestAccountInfoByAccountIndexLogic) GetLatestAccountInfoByAccountIndex(in *globalRPCProto.ReqGetLatestAccountInfoByAccountIndex) (*globalRPCProto.RespGetLatestAccountInfoByAccountIndex, error) {
	account, err := l.commglobalmap.GetLatestAccountInfoWithCache(l.ctx, int64(in.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] err:%v", err)
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrNotFound
		}
		return nil, errorcode.RpcErrInternal
	}
	resp := &globalRPCProto.RespGetLatestAccountInfoByAccountIndex{
		AccountId:       int64(account.AccountId),
		AccountIndex:    account.AccountIndex,
		AccountName:     account.AccountName,
		PublicKey:       account.PublicKey,
		AccountNameHash: account.AccountNameHash,
		L1Address:       account.L1Address,
		Nonce:           account.Nonce,
		CollectionNonce: account.CollectionNonce,
		AccountAsset:    make([]*globalRPCProto.AssetResult, 0),
		AssetRoot:       account.AssetRoot,
		Status:          int64(account.Status),
	}
	for assetID, asset := range account.AssetInfo {
		resp.AccountAsset = append(resp.AccountAsset, &globalRPCProto.AssetResult{
			AssetId:                  uint32(assetID),
			Balance:                  asset.Balance.String(),
			LpAmount:                 asset.LpAmount.String(),
			OfferCanceledOrFinalized: asset.OfferCanceledOrFinalized.String(),
		})
	}
	return resp, nil
}
