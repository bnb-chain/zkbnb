package transaction

import (
	"context"
	"strconv"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsListByAccountIndexLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Tx
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
}

func NewGetTxsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListByAccountIndexLogic {
	return &GetTxsListByAccountIndexLogic{
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

func (l *GetTxsListByAccountIndexLogic) GetTxsListByAccountIndex(req *types.ReqGetTxsListByAccountIndex) (resp *types.RespGetTxsListByAccountIndex, err error) {
	account, err := l.account.GetAccountByPk(req.AccountIndex)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
		return
	}
	// txCount, err := l.svcCtx.Tx.GetTxsTotalCountByAccountIndex(account.AccountIndex)
	// if err != nil {
	// 	logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
	// 	return
	// }
	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
		return
	}
	mempoolTxs, total, err := l.globalRPC.GetLatestTxsListByAccountIndex(uint32(account.AccountIndex), uint32(req.Limit), uint32(req.Offset))
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
		return
	}

	for _, tx := range mempoolTxs {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range tx.MempoolDetails {
			txDetails = append(txDetails, &types.TxDetail{
				AssetId:      int(txDetail.AssetId),
				AssetType:    int(txDetail.AssetType),
				AccountIndex: int32(txDetail.AccountIndex),
				AccountName:  txDetail.AccountName,
				AccountDelta: txDetail.BalanceDelta,
			})
		}
		txAmount, _ := strconv.Atoi(tx.TxAmount)
		gasFee, _ := strconv.ParseInt(tx.GasFee, 10, 64)
		blockInfo, err := l.block.GetBlockByBlockHeight(tx.L2BlockHeight)
		if err != nil {
			logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
			return nil, err
		}
		resp.Txs = append(resp.Txs, &types.Tx{
			TxHash:        tx.TxHash,
			TxType:        int32(tx.TxType),
			GasFeeAssetId: int32(tx.GasFeeAssetId),
			GasFee:        int32(gasFee),
			TxStatus:      int32(tx.Status),
			BlockHeight:   int64(tx.L2BlockHeight),
			BlockStatus:   int32(blockInfo.BlockStatus),
			BlockId:       int32(blockInfo.ID),
			//Todo: still need AssetAId, AssetBId?
			AssetAId:      int32(tx.AssetId),
			AssetBId:      int32(tx.AssetId),
			TxAmount:      int64(txAmount),
			TxDetails:     txDetails,
			NativeAddress: tx.NativeAddress,
			CreatedAt:     tx.CreatedAt.UnixNano() / 1e6,
			Memo:          tx.Memo,
		})
	}
	resp.Total = total + uint32(mempoolTxCount)
	return
}
