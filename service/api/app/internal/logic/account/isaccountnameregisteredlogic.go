package account

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsAccountNameRegisteredLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIsAccountNameRegisteredLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsAccountNameRegisteredLogic {
	return &IsAccountNameRegisteredLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IsAccountNameRegisteredLogic) IsAccountNameRegistered(req *types.ReqIsAccountNameRegistered) (resp *types.RespIsAccountNameRegistered, err error) {
	// todo: add your logic here and delete this line
	if utils.CheckAccountName(req.AccountName) {
		logx.Error("[CheckAccountName] req.AccountName:%v", req.AccountName)
		return nil, errcode.ErrInvalidParam
	}
	accountName := utils.FormatSting(req.AccountName)
	if utils.CheckFormatAccountName(accountName) {
		logx.Error("[CheckFormatAccountName] accountName:%v", accountName)
		return nil, errcode.ErrInvalidParam
	}
	////////
	isRegistered, err := l.svcCtx.AccountRegisterModel.IfAccountNameRegistered(accountName, nil)
	if err != nil {
		errInfo := fmt.Sprintf("[appService.account.IsAccountNameRegistered]<=>[AccountRegisterModel.IfAccountNameRegistered] %s", err.Error())
		logx.Errorf(errInfo)
		return packIsAccountNameRegistered(types.FailStatus, types.FailMsg, errInfo, result), nil
	}
	if !isRegistered {
		result.Status = 1
	}
	return packIsAccountNameRegistered(types.SuccessStatus, types.SuccessMsg, "", result), nil
}
