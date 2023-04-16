package account

import (
	"context"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/permctrl"
	types2 "github.com/bnb-chain/zkbnb/types"
	"strconv"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	permissionCheck *permctrl.PermissionCheck
}

func NewGetPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionLogic {
	permissionCheck := permctrl.NewPermissionCheck()
	return &GetPermissionLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		permissionCheck: permissionCheck,
	}
}

func (l *GetPermissionLogic) GetPermission(req *types.ReqGetPermission) (resp *types.Permission, err error) {

	l1Address, err := l.fetchAccountL1AddressFromReq(req)
	if err != nil {
		return nil, err
	}

	resp = &types.Permission{}
	permit, err := l.permissionCheck.CheckPermission(l1Address, req.TxType)
	if err != nil {
		resp.Message = l.resolveErrorMessage(err)
	}
	resp.Permit = permit

	return resp, nil
}

func (l *GetPermissionLogic) fetchAccountL1AddressFromReq(req *types.ReqGetPermission) (string, error) {
	switch req.By {
	case queryByIndex:
		accountIndex, err := strconv.ParseInt(req.Value, 10, 64)
		if err != nil || accountIndex < 0 {
			return "", types2.AppErrInvalidAccountIndex
		}
		account, err := l.svcCtx.StateFetcher.GetLatestAccount(accountIndex)
		if err != nil {
			return "", types2.AppErrInvalidAccountIndex
		}
		return account.L1Address, nil
	case queryByL1Address:
		return req.Value, nil
	}
	return "", types2.AppErrInvalidParam.RefineError("param by should be index|l1_address")
}

func (l *GetPermissionLogic) resolveErrorMessage(err error) string {
	if sysErr, ok := err.(*types2.SysError); ok {
		return sysErr.Message
	}
	if bizErr, ok := err.(*types2.BizError); ok {
		return bizErr.Message
	}
	return err.Error()
}
