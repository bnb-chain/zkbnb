package info

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/errorcode"
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

func (l *GetGasAccountLogic) GetGasAccount(req *types.ReqGetGasAccount) (resp *types.RespGetGasAccount, err error) {
	accountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[GetGasAccountLogic] get sys config error: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	accountIndex, err := strconv.ParseInt(accountIndexConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("[GetGasAccountLogic] invalid account index: %s", accountIndexConfig.Value)
		return nil, errorcode.AppErrInternal
	}

	accountModel, err := l.svcCtx.AccountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		logx.Errorf("[GetGasAccountLogic] get account error, index: %d, err: %s", accountIndex, err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp = &types.RespGetGasAccount{
		AccountStatus: int64(accountModel.Status),
		AccountIndex:  accountModel.AccountIndex,
		AccountName:   accountModel.AccountName,
	}
	return resp, nil
}
