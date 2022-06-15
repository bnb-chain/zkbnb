package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBalanceByAssetIdAndAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
	account   account.AccountModel
}

func NewGetBalanceByAssetIdAndAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBalanceByAssetIdAndAccountNameLogic {
	return &GetBalanceByAssetIdAndAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx.Config, ctx),
		account:   account.New(svcCtx.Config),
	}
}

func (l *GetBalanceByAssetIdAndAccountNameLogic) GetBalanceByAssetIdAndAccountName(req *types.ReqGetBlanceByAssetIdAndAccountName) (*types.RespGetBlanceInfoByAssetIdAndAccountName, error) {
	resp := &types.RespGetBlanceInfoByAssetIdAndAccountName{}
	if utils.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountIndex] param:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	account, err := l.account.GetAccountByAccountName(req.AccountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	assets, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(uint32(account.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	for _, asset := range assets {
		if req.AssetId == asset.AssetId {
			resp.Balance = asset.Balance
		}
	}
	return resp, nil
}
