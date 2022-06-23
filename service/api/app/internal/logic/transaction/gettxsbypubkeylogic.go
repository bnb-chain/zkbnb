package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByPubKeyLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	account   account.Account
	globalRpc globalrpc.GlobalRPC
	tx        tx.Model
	mempool   mempool.Mempool
	block     block.Block
}

func NewGetTxsByPubKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByPubKeyLogic {
	return &GetTxsByPubKeyLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		account:   account.New(svcCtx),
		globalRpc: globalrpc.New(svcCtx, ctx),
		tx:        tx.New(svcCtx),
		mempool:   mempool.New(svcCtx),
		block:     block.New(svcCtx),
	}
}

func (l *GetTxsByPubKeyLogic) GetTxsByPubKey(req *types.ReqGetTxsByPubKey) (resp *types.RespGetTxsByPubKey, err error) {
	account, err := l.account.GetAccountByPk(req.AccountPk)
	if err != nil {
		logx.Errorf("[GetAccountByPk] err:%v", err)
		return &types.RespGetTxsByPubKey{}, err
	}
	txList, _, err := l.globalRpc.GetLatestTxsListByAccountIndex(uint32(account.AccountIndex), req.Limit)
	if err != nil {
		logx.Errorf("[GetLatestTxsListByAccountIndex] err:%v", err)
		return &types.RespGetTxsByPubKey{}, err
	}
	txCount, err := l.tx.GetTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Errorf("[GetTxsTotalCountByAccountIndex] err:%v", err)
		return &types.RespGetTxsByPubKey{}, err
	}
	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Errorf("[GetMempoolTxsTotalCountByAccountIndex] err:%v", err)
		return &types.RespGetTxsByPubKey{}, err
	}
	results := make([]*types.Tx, 0)
	for _, tx := range txList {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range tx.MempoolDetails {
			txDetails = append(txDetails, &types.TxDetail{
				AssetId:        (txDetail.AssetId),
				AssetType:      (txDetail.AssetType),
				AccountIndex:   (txDetail.AccountIndex),
				AccountName:    txDetail.AccountName,
				AccountBalance: txDetail.BalanceDelta,
			})
		}
		block, err := l.block.GetBlockByBlockHeight(tx.L2BlockHeight)
		if err != nil {
			logx.Errorf("[transaction.GetTxsByPubKey] err:%v", err)
			return &types.RespGetTxsByPubKey{}, err
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
	return &types.RespGetTxsByPubKey{
		Total: uint32(txCount + mempoolTxCount),
		Txs:   results,
	}, nil
}
