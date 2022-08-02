package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
		logx.Errorf("[transaction.SendCancelOfferTx] err: %s", err.Error())
		return nil, err
	}
	return &types.RespSendCancelOfferTx{TxId: txIndex}, nil
}
