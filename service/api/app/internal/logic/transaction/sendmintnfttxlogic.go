package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMintNftTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	globalRpc globalrpc.GlobalRPC
}

func NewSendMintNftTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMintNftTxLogic {
	return &SendMintNftTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		globalRpc: globalrpc.New(svcCtx, ctx),
	}
}

func (l *SendMintNftTxLogic) SendMintNftTx(req *types.ReqSendMintNftTx) (*types.RespSendMintNftTx, error) {
	nftIndex, err := l.globalRpc.SendMintNftTx(req.TxInfo)
	if err != nil {
		logx.Error("[transaction.SendMintNftTx] err:%v", err)
		return nil, err
	}
	return &types.RespSendMintNftTx{NftIndex: nftIndex}, nil
}
