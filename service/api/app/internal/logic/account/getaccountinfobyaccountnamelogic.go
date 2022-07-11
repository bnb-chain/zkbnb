package account

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/common/checker"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAccountInfoByAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
	account   account.Model
}

func NewGetAccountInfoByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountInfoByAccountNameLogic {
	return &GetAccountInfoByAccountNameLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
		account:   account.New(svcCtx),
	}
}

func (l *GetAccountInfoByAccountNameLogic) GetAccountInfoByAccountName(req *types.ReqGetAccountInfoByAccountName) (*types.RespGetAccountInfoByAccountName, error) {
	if checker.CheckAccountName(req.AccountName) {
		logx.Errorf("[CheckAccountName] req.AccountName:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	accountName := checker.FormatSting(req.AccountName)
	if checker.CheckFormatAccountName(accountName) {
		logx.Errorf("[CheckFormatAccountName] accountName:%v", accountName)
		return nil, errcode.ErrInvalidParam
	}
	info, err := l.account.GetAccountByAccountName(l.ctx, accountName)
	if err != nil {
		logx.Errorf("[GetAccountByAccountName] accountName:%v, err:%v", accountName, err)
		return nil, err
	}
	account, err := l.globalRPC.GetLatestAccountInfoByAccountIndex(uint32(info.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfoByAccountIndex] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetAccountInfoByAccountName{
		AccountIndex: uint32(account.AccountIndex),
		AccountPk:    account.PublicKey,
		Nonce:        account.Nonce,
		Assets:       make([]*types.AccountAsset, 0),
	}
	for _, asset := range account.AccountAsset {
		resp.Assets = append(resp.Assets, &types.AccountAsset{
			AssetId:                  asset.AssetId,
			Balance:                  asset.Balance,
			LpAmount:                 asset.LpAmount,
			OfferCanceledOrFinalized: asset.OfferCanceledOrFinalized,
		})
	}
	return resp, nil
}
