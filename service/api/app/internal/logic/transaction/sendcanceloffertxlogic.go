package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendCancelOfferTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendCancelOfferTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCancelOfferTxLogic {
	return &SendCancelOfferTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendCancelOfferTxLogic) SendCancelOfferTx(req *types.ReqSendCancelOfferTx) (resp *types.RespSendCancelOfferTx, err error) {
	txIndex, err := l.globalRpc.SendCancelOfferTx(req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendCancelOfferTx] err:%v", err)
		return nil, err
	}

	return &types.RespSendCancelOfferTx{TxId: txIndex}, nil
}
