package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendSwapTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendSwapTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSwapTxLogic {
	return &SendSwapTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendSwapTxLogic) SendSwapTx(req *types.ReqSendSwapTx) (*types.RespSendSwapTx, error) {
	txIndex, err := l.globalRpc.SendSwapTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendSwapTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendSwapTx{TxId: txIndex}, nil
}
