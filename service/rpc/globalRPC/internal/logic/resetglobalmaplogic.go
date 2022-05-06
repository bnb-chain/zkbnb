package logic

import (
	"context"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/logic/globalmapHandler"

	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetGlobalMapLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetGlobalMapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetGlobalMapLogic {
	return &ResetGlobalMapLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func packResetGlobalMap(
	status int64,
	msg string,
	err string,
) (res *globalRPCProto.RespResetGlobalMap) {
	res = &globalRPCProto.RespResetGlobalMap{
		Status: status,
		Msg:    msg,
		Err:    err,
	}
	return res
}

//  reset globalmap
func (l *ResetGlobalMapLogic) ResetGlobalMap(in *globalRPCProto.ReqResetGlobalMap) (resp *globalRPCProto.RespResetGlobalMap, err error) {
	globalmapHandler.ResetGlobalMap()
	return packResetGlobalMap(SuccessStatus, SuccessMsg, ""), nil
}
