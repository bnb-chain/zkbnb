package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendRemoveLiquidityTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendRemoveLiquidityTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendRemoveLiquidityTxLogic {
	return &SendRemoveLiquidityTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendRemoveLiquidityTxLogic) SendRemoveLiquidityTx(req *types.ReqSendRemoveLiquidityTx) (*types.RespSendRemoveLiquidityTx, error) {
	txIndex, err := l.globalRpc.SendRemoveLiquidityTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Errorf("[transaction.SendRemoveLiquidityTx] err: %s", err.Error())
		return nil, err
	}
	return &types.RespSendRemoveLiquidityTx{TxId: txIndex}, nil
}
