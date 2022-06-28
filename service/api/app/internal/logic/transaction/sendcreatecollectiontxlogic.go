package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *SendCreateCollectionTxLogic) SendCreateCollectionTx(req *types.ReqSendCreateCollectionTx) (resp *types.RespSendCreateCollectionTx, err error) {
	collectionId, err := l.globalRpc.SendCreateCollectionTx(req.TxInfo)
	if err != nil {
		logx.Error("[SendCreateCollectionTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendCreateCollectionTx{CollectionId: collectionId}, nil
}
