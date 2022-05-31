package account

import (
	"context"
	"fmt"
	"time"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBalanceByAssetIdAndAccountNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBalanceByAssetIdAndAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBalanceByAssetIdAndAccountNameLogic {
	return &GetBalanceByAssetIdAndAccountNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBalanceByAssetIdAndAccountNameLogic) GetBalanceByAssetIdAndAccountName(req *types.ReqGetBlanceByAssetIdAndAccountName) (resp *types.RespGetBlanceInfoByAssetIdAndAccountName, err error) {
	if utils.CheckAccountName(req.AccountName) {
		logx.Error("[CheckAccountIndex] param:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	account, err := l.account.GetAccountByAccountName(req.AccountName)
	if err != nil {
		logx.Error("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	// TODO: get latestAssets from globalRPC
	expire_time := 0
	if account.Status == 2 {
		accountRegister, err := l.svcCtx.AccountRegisterModel.GetAccountRegisterInfoByName(req.AccountName)
		if err != nil {
			errInfo := fmt.Sprintf("[appService.account.GetAccountStatusByAccountName]<=>[AccountRegisterModel.GetAccountRegisterInfoByName] %s", err.Error())
			logx.Errorf(errInfo)
			return packGetAccountStatusByAccountName(types.FailStatus, types.FailMsg, errInfo, result), nil
		}
		h, _ := time.ParseDuration("-24h")
		expire_time = int(accountRegister.Model.CreatedAt.Add(h).Unix())
	}
	result = types.ResultGetAccountStatusByAccountName{
		AccountStatus: uint8(accountStatus),
		PublicKey:     pk,
		ExpireTime:    int64(expire_time),
	}
	return packGetAccountStatusByAccountName(types.SuccessStatus, types.SuccessMsg, "", result), nil
}
