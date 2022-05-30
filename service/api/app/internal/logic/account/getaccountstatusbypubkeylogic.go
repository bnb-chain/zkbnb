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

type GetAccountStatusByPubKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountStatusByPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountStatusByPubKeyLogic {
	return &GetAccountStatusByPubKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountStatusByPubKeyLogic) GetAccountStatusByPubKey(req *types.ReqGetAccountStatusByPubKey) (resp *types.RespGetAccountStatusByPubKey, err error) {
	// todo: add your logic here and delete this line
	if utils.CheckAccountPK(req.AccountPk) {
		logx.Error("[CheckAccountPK] param:%v", req.AccountPk)
		return nil, errcode.ErrInvalidParam
	}
	account, err := l.account.GetAccountByAccountName(req.AccountName)
	if err != nil {
		logx.Error("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}

	////////
	expire_time := 0
	if account.Status == 2 {
		accountRegister, err := l.svcCtx.AccountRegisterModel.GetAccountRegisterInfoByPublicKey(req.AccountPk)
		if err != nil {
			errInfo := fmt.Sprintf("[appService.account.GetAccountStatusByPubKey]<=>[AccountRegisterModel.GetAccountRegisterInfoByPublic] %s", err.Error())
			logx.Errorf(errInfo)
			return packGetAccountStatusByPubKey(types.FailStatus, types.FailMsg, errInfo, result), nil
		}
		h, _ := time.ParseDuration("-24h")
		expire_time = int(accountRegister.Model.CreatedAt.Add(h).Unix())
	}
	result = types.ResultGetAccountStatusByPubKey{
		AccountStatus: uint8(accountStatus),
		AccountName:   accountName,
		ExpireTime:    int64(expire_time),
	}
	return packGetAccountStatusByPubKey(types.SuccessStatus, types.SuccessMsg, "", result), nil
}
