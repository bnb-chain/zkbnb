package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendWithdrawTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendWithdrawTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawTxLogic {
	return &SendWithdrawTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendWithdrawTxLogic) SendWithdrawTx(req *types.ReqSendWithdrawTx) (*types.RespSendWithdrawTx, error) {
	txIndex, err := l.globalRpc.SendWithdrawTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendWithdrawTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendWithdrawTx{TxId: txIndex}, nil
}
