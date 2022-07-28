package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetTxByHashLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	mempool mempool.Mempool
	block   block.Block
	tx      tx.Model
}

func NewGetTxByHashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxByHashLogic {
	return &GetTxByHashLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		mempool: mempool.New(svcCtx),
		block:   block.New(svcCtx),
		tx:      tx.New(svcCtx),
	}
}

func (l *GetTxByHashLogic) GetTxByHash(req *types.ReqGetTxByHash) (*types.RespGetTxByHash, error) {
	resp := &types.RespGetTxByHash{}
	tx, err := l.tx.GetTxByTxHash(l.ctx, req.TxHash)
	if err == nil {
		resp.Tx = *utils.GormTx2Tx(tx)
	}
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
		memppolTx, err := l.mempool.GetMempoolTxByTxHash(req.TxHash)
		if err != nil {
			logx.Errorf("[GetMempoolTxByTxHash]:%v", err)
			if err == errorcode.DbErrNotFound {
				return nil, errorcode.AppErrNotFound
			}
			return nil, errorcode.AppErrInternal
		}
		resp.Tx = *utils.MempoolTx2Tx(memppolTx)
	}
	if resp.Tx.TxType == commonTx.TxTypeSwap {
		txInfo, err := commonTx.ParseSwapTxInfo(tx.TxInfo)
		if err != nil {
			logx.Errorf("[ParseSwapTxInfo]:%v", err)
			return nil, errorcode.AppErrInternal
		}
		resp.AssetAId = txInfo.AssetAId
		resp.AssetBId = txInfo.AssetBId
	}
	block, err := l.block.GetBlockByBlockHeight(l.ctx, resp.Tx.BlockHeight)
	if err == nil {
		resp.CommittedAt = block.CommittedAt
		resp.ExecutedAt = block.CreatedAt.Unix()
		resp.VerifiedAt = block.VerifiedAt
	}
	return resp, nil
}
