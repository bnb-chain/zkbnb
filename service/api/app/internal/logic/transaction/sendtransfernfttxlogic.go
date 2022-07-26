package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendTransferNftTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendTransferNftTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTransferNftTxLogic {
	return &SendTransferNftTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendTransferNftTxLogic) SendTransferNftTx(req *types.ReqSendTransferNftTx) (*types.RespSendTransferNftTx, error) {
	txIndex, err := l.globalRpc.SendTransferNftTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendTransferNftTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendTransferNftTx{TxId: txIndex}, nil
}
