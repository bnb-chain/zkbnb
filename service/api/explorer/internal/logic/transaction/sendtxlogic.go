package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"
)

type SendTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRPC globalrpc.GlobalRPC
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendTxLogic) SendTx(req *types.ReqSendTx) (*types.RespSendTx, error) {
	txId, err := l.globalRPC.SendTx(req.TxType, req.TxInfo)
	if err != nil {
		logx.Errorf("[transaction.SendTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendTx{TxId: txId}, nil
}
