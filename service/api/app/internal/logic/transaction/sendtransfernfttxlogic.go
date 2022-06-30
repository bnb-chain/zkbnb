package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

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
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendTransferNftTxLogic) SendTransferNftTx(req *types.ReqSendTransferNftTx) (resp *types.RespSendTransferNftTx, err error) {
	txIndex, err := l.globalRpc.SendTransferNftTx(req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendTransferNftTx] err:%v", err)
		return nil, err
	}

	return &types.RespSendTransferNftTx{TxId: txIndex}, nil
}
