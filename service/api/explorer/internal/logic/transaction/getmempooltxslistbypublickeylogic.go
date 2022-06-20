package transaction

import (
	"context"

	blockmodel "github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMempoolTxsListByPublicKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

	tx        tx.Tx
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
}

func NewGetMempoolTxsListByPublicKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsListByPublicKeyLogic {
	return &GetMempoolTxsListByPublicKeyLogic{
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

func (l *GetMempoolTxsListByPublicKeyLogic) GetMempoolTxsListByPublicKey(req *types.ReqGetMempoolTxsListByPublicKey) (*types.RespGetMempoolTxsListByPublicKey, error) {
	resp := &types.RespGetMempoolTxsListByPublicKey{}
	account, err := l.account.GetAccountByPk(req.AccountPk)
	if err != nil {
		l.Error("[transaction.GetTxsByPubKey] err:%v", err)
		return nil, err
	}

	txList, _, err := l.globalRPC.GetLatestTxsListByAccountIndex(uint32(account.AccountIndex), uint32(req.Limit), uint32(req.Offset))
	if err != nil {
		l.Error("[transaction.GetTxsByPubKey] err:%v", err)
		return nil, err
	}
	txCount, err := l.tx.GetTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Error("[transaction.GetTxsByPubKey] err:%v", err)
		return nil, err
	}

	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Error("[transaction.GetTxsByPubKey] err:%v", err)
		return nil, err
	}

	for _, tx := range txList {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range tx.MempoolDetails {
			txDetails = append(txDetails, &types.TxDetail{
				AssetId:      txDetail.AssetId,
				AssetType:    txDetail.AssetType,
				AccountIndex: txDetail.AccountIndex,
				AccountName:  txDetail.AccountName,
			})
		}
		var blockInfo *blockmodel.Block
		blockInfo, err = l.block.GetBlockByBlockHeight(tx.L2BlockHeight)
		if err != nil {
			logx.Error("[transaction.GetTxsByPubKey] err:%v", err)
			return nil, err
		}
		resp.Txs = append(resp.Txs, &types.Tx{
			TxHash:        tx.TxHash,
			TxType:        tx.TxType,
			GasFeeAssetId: tx.GasFeeAssetId,
			GasFee:        tx.GasFee,
			TxStatus:      int64(tx.Status),
			BlockHeight:   tx.L2BlockHeight,
			BlockId:       int64(blockInfo.ID),
			TxAmount:      tx.TxAmount,
			TxDetails:     txDetails,
			NativeAddress: tx.NativeAddress,
			Memo:          tx.Memo,
		})
	}
	resp.Total = uint32(txCount + mempoolTxCount)
	return resp, nil
}
