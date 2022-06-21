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
	for _, memTx := range mempoolTxs {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range memTx.MempoolDetails {
			txDetails = append(txDetails, &types.TxDetail{
				AssetId:      uint32(txDetail.AssetId),
				AssetType:    uint32(txDetail.AssetType),
				AccountIndex: int32(txDetail.AccountIndex),
				AccountName:  txDetail.AccountName,
				AccountDelta: txDetail.BalanceDelta,
			})
		}
		blockInfo, err := l.block.GetBlockByBlockHeight(memTx.L2BlockHeight)
		if err != nil {
			logx.Errorf("[GetBlockByBlockHeight]:%v", err)
			return nil, err
		}
		results = append(results, &types.Tx{
			TxHash:        memTx.TxHash,
			TxType:        uint32(memTx.TxType),
			GasFeeAssetId: uint32(memTx.GasFeeAssetId),
			GasFee:        memTx.GasFee,
			NftIndex:      uint32(memTx.NftIndex),
			PairIndex:     uint32(memTx.PairIndex),
			AssetId:       uint32(memTx.AssetId),
			TxAmount:      memTx.TxAmount,
			NativeAddress: memTx.NativeAddress,
			TxDetails:     txDetails,
			TxInfo:        memTx.TxInfo,
			ExtraInfo:     memTx.ExtraInfo,
			Memo:          memTx.Memo,
			AccountIndex:  uint32(memTx.AccountIndex),
			Nonce:         uint32(memTx.Nonce),
			ExpiredAt:     uint32(memTx.ExpiredAt),
			L2BlockHeight: uint32(memTx.L2BlockHeight),
			Status:        uint32(memTx.Status),
			CreatedAt:     uint32(memTx.CreatedAt.Unix()),
			BlockID:       uint32(blockInfo.ID),
		})
	}
	return &types.RespGetTxsByAccountIndexAndTxType{Total: uint32(txCount + mempoolTxCount), Txs: results}, nil
}
