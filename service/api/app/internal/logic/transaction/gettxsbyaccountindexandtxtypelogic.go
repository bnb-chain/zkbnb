package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByAccountIndexAndTxTypeLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Tx
	globalRPC globalrpc.GlobalRPC
	block     block.Block
	mempool   mempool.Mempool
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
	}
}

func (l *GetTxsByAccountIndexAndTxTypeLogic) GetTxsByAccountIndexAndTxType(req *types.ReqGetTxsByAccountIndexAndTxType) (resp *types.RespGetTxsByAccountIndexAndTxType, err error) {
	txCount, err := l.tx.GetTxsTotalCountByAccountIndex(int64(req.AccountIndex))
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
				AssetId:      uint32(txDetail.AssetId),
				AssetType:    uint32(txDetail.AssetType),
				AccountIndex: int32(txDetail.AccountIndex),
				AccountName:  txDetail.AccountName,
				AccountDelta: txDetail.BalanceDelta,
			})
		}
		block, err := l.block.GetBlockByBlockHeight(tx.L2BlockHeight)
		if err != nil {
			logx.Errorf("[GetBlockByBlockHeight]:%v", err)
			return nil, err
		}
		results = append(results, &types.Tx{
			TxHash:        tx.TxHash,
			TxType:        uint32(tx.TxType),
			GasFeeAssetId: uint32(tx.GasFeeAssetId),
			GasFee:        tx.GasFee,
			NftIndex:      uint32(tx.NftIndex),
			PairIndex:     uint32(tx.PairIndex),
			AssetId:       uint32(tx.AssetId),
			TxAmount:      tx.TxAmount,
			NativeAddress: tx.NativeAddress,
			TxDetails:     txDetails,
			TxInfo:        tx.TxInfo,
			ExtraInfo:     tx.ExtraInfo,
			Memo:          tx.Memo,
			AccountIndex:  uint32(tx.AccountIndex),
			Nonce:         uint32(tx.Nonce),
			ExpiredAt:     uint32(tx.ExpiredAt),
			L2BlockHeight: uint32(tx.L2BlockHeight),
			Status:        uint32(tx.Status),
			CreatedAt:     uint32(tx.CreatedAt.Unix()),
			BlockID:       uint32(block.ID),
		})
	}
	return &types.RespGetTxsByAccountIndexAndTxType{Total: uint32(txCount + mempoolTxCount), Txs: results}, nil
}
