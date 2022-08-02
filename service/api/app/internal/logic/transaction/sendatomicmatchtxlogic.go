package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendAtomicMatchTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendAtomicMatchTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAtomicMatchTxLogic {
	return &SendAtomicMatchTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendAtomicMatchTxLogic) SendAtomicMatchTx(req *types.ReqSendAtomicMatchTx) (*types.RespSendAtomicMatchTx, error) {
	txIndex, err := l.globalRpc.SendAtomicMatchTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Errorf("[transaction.SendAtomicMatchTx] err: %s", err.Error())
		return nil, err
	}
	return &types.RespSendAtomicMatchTx{TxId: txIndex}, nil
}
