package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
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
		logx.Errorf("[GetLatestAccountInfo] err: %s", err.Error())
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.RpcErrNotFound
		}
		return nil, errorcode.RpcErrInternal
	}
	return &globalRPCProto.RespGetNextNonce{
		Nonce: uint64(accountInfo.Nonce),
	}, nil
}
