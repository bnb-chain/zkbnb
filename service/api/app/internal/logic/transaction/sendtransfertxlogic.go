package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendTransferTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendTransferTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTransferTxLogic {
	return &SendTransferTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendTransferTxLogic) SendTransferTx(req *types.ReqSendTransferTx) (*types.RespSendTransferTx, error) {
	txIndex, err := l.globalRpc.SendTransferTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Errorf("[transaction.SendTransferTx] err: %s", err.Error())
		return nil, err
	}
	return &types.RespSendTransferTx{TxId: txIndex}, nil
}
