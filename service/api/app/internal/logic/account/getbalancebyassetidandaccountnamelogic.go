package account

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/checker"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetBalanceByAssetIdAndAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
	account   account.Model
}

func NewGetBalanceByAssetIdAndAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBalanceByAssetIdAndAccountNameLogic {
	return &GetBalanceByAssetIdAndAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
		account:   account.New(svcCtx),
	}
}

func (l *GetBalanceByAssetIdAndAccountNameLogic) GetBalanceByAssetIdAndAccountName(req *types.ReqGetBlanceByAssetIdAndAccountName) (*types.RespGetBlanceInfoByAssetIdAndAccountName, error) {
	resp := &types.RespGetBlanceInfoByAssetIdAndAccountName{}
	if checker.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountIndex] param:%v", req.AccountName)
		return nil, errorcode.AppErrInvalidParam
	}
	account, err := l.account.GetAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	assets, err := l.globalRPC.GetLatestAssetsListByAccountIndex(l.ctx, uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAssetsListByAccountIndex] err:%v", err)
		return nil, err
	}
	for _, asset := range assets {
		if req.AssetId == asset.AssetId {
			resp.Balance = asset.Balance
		}
	}
	return resp, nil
}
