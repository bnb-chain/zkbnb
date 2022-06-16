package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByAccountNameLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	account   account.AccountModel
	tx        tx.Tx
	globalRpc globalrpc.GlobalRPC
	mempool   mempool.Mempool
	block     block.Block
}

func NewGetTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountNameLogic {
	return &GetTxsByAccountNameLogic{
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

func (l *GetTxsByAccountNameLogic) GetTxsByAccountName(req *types.ReqGetTxsByAccountName) (resp *types.RespGetTxsByAccountName, err error) {
	account, err := l.account.GetAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	txList, _, err := l.globalRpc.GetLatestTxsListByAccountIndex(uint32(account.AccountIndex), req.Limit)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	txCount, err := l.tx.GetTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	results := make([]*types.Tx, 0)
	for _, tx := range txList {
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
			logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
			return &types.RespGetTxsByAccountName{}, err
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
	return &types.RespGetTxsByAccountName{Total: uint32(txCount + mempoolTxCount), Txs: results}, nil
}
