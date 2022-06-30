package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendAddLiquidityTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendAddLiquidityTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAddLiquidityTxLogic {
	return &SendAddLiquidityTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendAddLiquidityTxLogic) SendAddLiquidityTx(req *types.ReqSendAddLiquidityTx) (resp *types.RespSendAddLiquidityTx, err error) {
	txIndex, err := l.globalRpc.SendAddLiquidityTx(req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendAddLiquidityTx] err:%v", err)
		return nil, err
	}

	return &types.RespSendAddLiquidityTx{TxId: txIndex}, nil
}
