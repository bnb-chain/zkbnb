package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNextNonceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewGetNextNonceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNextNonceLogic {
	return &GetNextNonceLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *GetNextNonceLogic) GetNextNonce(in *globalRPCProto.ReqGetNextNonce) (*globalRPCProto.RespGetNextNonce, error) {
	accountInfo, err := l.commglobalmap.GetLatestAccountInfo(l.ctx, int64(in.AccountIndex))
	if err != nil {
		logx.Errorf("[GetLatestAccountInfo] err:%v", err)
		return nil, err
	}
	return &globalRPCProto.RespGetNextNonce{
		Nonce: uint64(accountInfo.Nonce),
	}, nil
}
