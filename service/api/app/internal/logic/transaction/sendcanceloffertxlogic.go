package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

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
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendCancelOfferTxLogic) SendCancelOfferTx(req *types.ReqSendCancelOfferTx) (*types.RespSendCancelOfferTx, error) {
	txIndex, err := l.globalRpc.SendCancelOfferTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendCancelOfferTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendCancelOfferTx{TxId: txIndex}, nil
}
