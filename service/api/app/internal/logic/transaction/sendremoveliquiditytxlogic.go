package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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
	txIndex, err := l.globalRpc.SendRemoveLiquidityTx(req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendRemoveLiquidityTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendRemoveLiquidityTx{TxId: txIndex}, nil
}
