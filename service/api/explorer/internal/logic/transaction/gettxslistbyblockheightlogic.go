package transaction

import (
	"context"
	"strconv"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/mempool"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/tx"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsListByBlockHeightLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Tx
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
}

func NewGetTxsListByBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListByBlockHeightLogic {
	return &GetTxsListByBlockHeightLogic{
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

func (l *GetTxsListByBlockHeightLogic) GetTxsListByBlockHeight(req *types.ReqGetTxsListByBlockHeight) (resp *types.RespGetTxsListByBlockHeight, err error) {
	b, err := l.block.GetBlockByBlockHeight(int64(req.BlockHeight))
	if err != nil {
		l.Error("[transaction.GetBlockByBlockHeight] err:%v", err)
		return
	}
	txs, total, err := l.tx.GetTxsByBlockId(int64(b.ID), uint32(req.Limit), uint32(req.Offset))
	if err != nil {
		l.Error("[transaction.GetTxsByBlockId] err:%v", err)
		return
	}

	for _, txInfo := range txs {
		txAmount, _ := strconv.Atoi(txInfo.TxAmount)
		gasFee, _ := strconv.ParseInt(txInfo.GasFee, 10, 64)
		respTxs := &types.Tx{
			TxHash:        txInfo.TxHash,
			TxType:        int32(txInfo.TxType),
			GasFeeAssetId: int32(txInfo.GasFeeAssetId),
			GasFee:        int32(gasFee),
			TxStatus:      int32(txInfo.TxStatus),
			BlockHeight:   txInfo.BlockHeight,
			BlockStatus:   int32(b.BlockStatus),
			BlockId:       int32(txInfo.BlockId),
			//Todo: still need AssetAId, AssetBId?
			AssetAId:      int32(txInfo.AssetId),
			AssetBId:      int32(txInfo.AssetId),
			TxAmount:      int64(txAmount),
			NativeAddress: txInfo.NativeAddress,
			CreatedAt:     txInfo.CreatedAt.UnixNano() / 1e6,
			Memo:          txInfo.Memo,
		}
		for _, d := range txInfo.TxDetails {
			respTxs.TxDetails = append(respTxs.TxDetails, &types.TxDetail{
				AssetId:      int(d.AssetId),
				AssetType:    int(d.AssetType),
				AccountIndex: int32(d.AccountIndex),
				AccountName:  d.AccountName,
				AccountDelta: d.BalanceDelta,
			})
		}
		resp.Txs = append(resp.Txs, respTxs)
	}
	resp.Total = uint32(total)
	return
}
