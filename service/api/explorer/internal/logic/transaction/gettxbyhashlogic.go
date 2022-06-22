package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxByHashLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Model
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
}

func NewGetTxByHashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxByHashLogic {
	return &GetTxByHashLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		tx:        tx.New(svcCtx),
		block:     block.New(svcCtx),
		account:   account.New(svcCtx),
		mempool:   mempool.New(svcCtx),
		globalRPC: globalrpc.New(svcCtx, ctx),
	}
}

func (l *GetTxByHashLogic) GetTxByHash(req *types.ReqGetTxByHash) (*types.RespGetTxByHash, error) {
	tx, err := l.tx.GetTxByTxHash(req.TxHash)
	if err != nil {
		logx.Errorf("[GetTxByTxHash]:%v", err)
		return nil, err
	}
	return &types.RespGetTxByHash{
		Txs: *utils.GormTx2Tx(tx),
	}, nil
}
