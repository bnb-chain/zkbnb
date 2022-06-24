package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/txdetail"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByAccountIndexAndTxTypeLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Model
	globalRPC globalrpc.GlobalRPC
	block     block.Block
	mempool   mempool.Mempool
	txDetail  txdetail.Model
}

func NewGetTxsByAccountIndexAndTxTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountIndexAndTxTypeLogic {
	return &GetTxsByAccountIndexAndTxTypeLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		tx:        tx.New(svcCtx),
		globalRPC: globalrpc.New(svcCtx, ctx),
		block:     block.New(svcCtx),
		mempool:   mempool.New(svcCtx),
		txDetail:  txdetail.New(svcCtx),
	}
}

func (l *GetTxsByAccountIndexAndTxTypeLogic) GetTxsByAccountIndexAndTxType(req *types.ReqGetTxsByAccountIndexAndTxType) (resp *types.RespGetTxsByAccountIndexAndTxType, err error) {
	txCount, err := l.txDetail.GetTxsTotalCountByAccountIndex(l.ctx, int64(req.AccountIndex))
	if err != nil {
		logx.Error("[GetTxsTotalCountByAccountIndex] err:%v", err)
		return nil, err
	}
	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndex(int64(req.AccountIndex))
	if err != nil {
		logx.Error("[GetMempoolTxsTotalCountByAccountIndex] err:%v", err)
		return nil, err
	}
	mempoolTxs, err := l.globalRPC.GetLatestTxsListByAccountIndexAndTxType(uint64(req.AccountIndex), uint64(req.TxType), uint64(req.Limit), uint64(req.Offset))
	if err != nil {
		logx.Error("[GetLatestTxsListByAccountIndexAndTxType] err:%v", err)
		return nil, err
	}
	results := make([]*types.Tx, 0)
	for _, tx := range mempoolTxs {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range tx.MempoolDetails {
			txDetails = append(txDetails, &types.TxDetail{
				AssetId:      (txDetail.AssetId),
				AssetType:    (txDetail.AssetType),
				AccountIndex: (txDetail.AccountIndex),
				AccountName:  txDetail.AccountName,
				AccountDelta: txDetail.BalanceDelta,
			})
		}
		block, err := l.block.GetBlockByBlockHeight(l.ctx, tx.L2BlockHeight)
		if err != nil {
			logx.Errorf("[GetBlockByBlockHeight]:%v", err)
			return nil, err
		}
		results = append(results, &types.Tx{
			TxHash:        tx.TxHash,
			TxType:        (tx.TxType),
			GasFeeAssetId: (tx.GasFeeAssetId),
			GasFee:        tx.GasFee,
			NftIndex:      (tx.NftIndex),
			PairIndex:     (tx.PairIndex),
			AssetId:       (tx.AssetId),
			TxAmount:      tx.TxAmount,
			NativeAddress: tx.NativeAddress,
			TxDetails:     txDetails,
			TxInfo:        tx.TxInfo,
			ExtraInfo:     tx.ExtraInfo,
			Memo:          tx.Memo,
			AccountIndex:  (tx.AccountIndex),
			Nonce:         (tx.Nonce),
			ExpiredAt:     (tx.ExpiredAt),
			BlockHeight:   (tx.L2BlockHeight),
			Status:        int64(tx.Status),
			CreatedAt:     (tx.CreatedAt.Unix()),
			BlockId:       int64(block.ID),
		})
	}
	return &types.RespGetTxsByAccountIndexAndTxType{Total: uint32(txCount + mempoolTxCount), Txs: results}, nil
}
