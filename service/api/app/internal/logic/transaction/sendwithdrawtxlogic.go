package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

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
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendWithdrawTxLogic) SendWithdrawTx(req *types.ReqSendWithdrawTx) (resp *types.RespSendWithdrawTx, err error) {
	txIndex, err := l.globalRpc.SendWithdrawTx(req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendWithdrawTx] err:%v", err)
		return nil, err
	}

	return &types.RespSendWithdrawTx{TxId: txIndex}, nil
}
