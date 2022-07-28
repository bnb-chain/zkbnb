package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendCreateCollectionTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendCreateCollectionTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCreateCollectionTxLogic {
	return &SendCreateCollectionTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendCreateCollectionTxLogic) SendCreateCollectionTx(req *types.ReqSendCreateCollectionTx) (*types.RespSendCreateCollectionTx, error) {
	collectionId, err := l.globalRpc.SendCreateCollectionTx(l.ctx, req.TxInfo)
	if err != nil {
		logx.Error("[SendCreateCollectionTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendCreateCollectionTx{CollectionId: collectionId}, nil
}
