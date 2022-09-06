package info

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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

func (l *GetGasAccountLogic) GetGasAccount() (resp *types.GasAccount, err error) {
	accountIndexConfig, err := l.svcCtx.MemCache.GetSysConfigWithFallback(types2.GasAccountIndex, func() (interface{}, error) {
		return l.svcCtx.SysConfigModel.GetSysConfigByName(types2.GasAccountIndex)
	})
	if err != nil {
		return nil, types2.AppErrInternal
	}

	accountIndex, err := strconv.ParseInt(accountIndexConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("invalid account index: %s", accountIndexConfig.Value)
		return nil, types2.AppErrInternal
	}

	account, err := l.svcCtx.MemCache.GetAccountWithFallback(accountIndex, func() (interface{}, error) {
		return l.svcCtx.AccountModel.GetAccountByIndex(accountIndex)
	})
	if err != nil {
		return nil, types2.AppErrInternal
	}

	resp = &types.GasAccount{
		Status: int64(account.Status),
		Index:  account.AccountIndex,
		Name:   account.AccountName,
	}
	return resp, nil
}
