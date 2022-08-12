package info

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/sysConfigName"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetGasAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGasAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasAccountLogic {
	return &GetGasAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGasAccountLogic) GetGasAccount() (resp *types.RespGetGasAccount, err error) {
	accountIndexConfig, err := l.svcCtx.MemCache.GetSysConfigWithFallback(sysConfigName.GasAccountIndex, func() (interface{}, error) {
		return l.svcCtx.SysConfigModel.GetSysConfigByName(sysConfigName.GasAccountIndex)
	})
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	accountIndex, err := strconv.ParseInt(accountIndexConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("invalid account index: %s", accountIndexConfig.Value)
		return nil, errorcode.AppErrInternal
	}

	account, err := l.svcCtx.MemCache.GetAccountWithFallback(accountIndex, func() (interface{}, error) {
		return l.svcCtx.AccountModel.GetAccountByAccountIndex(accountIndex)
	})
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	resp = &types.RespGetGasAccount{
		AccountStatus: int64(account.Status),
		AccountIndex:  account.AccountIndex,
		AccountName:   account.AccountName,
	}
	return resp, nil
}
