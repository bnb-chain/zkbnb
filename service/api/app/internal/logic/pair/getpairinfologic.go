package pair

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPairInfoLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewGetPairInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPairInfoLogic {
	return &GetPairInfoLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx.Config),
	}
}

func (l *GetPairInfoLogic) GetPairInfo(req *types.ReqGetPairInfo) (resp *types.RespGetPairInfo, err error) {
	// todo: add your logic here and delete this line
	if utils.CheckPairIndex(req.PairIndex) {
		logx.Error("[CheckPairIndex] param:%v", req.PairIndex)
		return nil, errcode.ErrInvalidParam
	}
	resRpc, err := l.globalRPC.GetPairRatio(req.PairIndex)
	if err != nil {
		logx.Error("[GetPairRatio] err:%v", err)
		return nil, err
	}
	return resp, nil
}
