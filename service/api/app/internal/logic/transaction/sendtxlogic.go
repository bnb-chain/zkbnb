package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type SendTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.RespSendTx, err error) {
	//err := utils.CheckRequestParam(utils.TypeTxType, reflect.ValueOf(req.TxType))
	txId, err := l.globalRpc.SendTx(l.ctx, req.TxType, req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendTx{TxId: txId}, nil
}
