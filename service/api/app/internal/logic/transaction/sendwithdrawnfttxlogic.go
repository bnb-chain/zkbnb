package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendWithdrawNftTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendWithdrawNftTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawNftTxLogic {
	return &SendWithdrawNftTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendWithdrawNftTxLogic) SendWithdrawNftTx(req *types.ReqSendWithdrawNftTx) (*types.RespSendWithdrawNftTx, error) {
	txIndex, err := l.globalRpc.SendWithdrawNftTx(req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendWithdrawNftTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendWithdrawNftTx{TxId: txIndex}, nil
}
