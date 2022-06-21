package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/globalrpc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
	accountInfo, err := l.account.GetAccountByAccountName(l.ctx, req.AccountName)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	txList, _, err := l.globalRpc.GetLatestTxsListByAccountIndex(uint32(accountInfo.AccountIndex), req.Limit)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	txCount, err := l.tx.GetTxsTotalCountByAccountIndex(accountInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndex(accountInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
		return nil, err
	}
	results := make([]*types.Tx, 0)
	for _, txInfo := range txList {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range txInfo.MempoolDetails {
			txDetails = append(txDetails, &types.TxDetail{
				AssetId:      uint32(txDetail.AssetId),
				AssetType:    uint32(txDetail.AssetType),
				AccountIndex: int32(txDetail.AccountIndex),
				AccountName:  txDetail.AccountName,
				AccountDelta: txDetail.BalanceDelta,
			})
		}
		blockInfo, err := l.block.GetBlockByBlockHeight(txInfo.L2BlockHeight)
		if err != nil {
			logx.Errorf("[transaction.GetTxsByAccountName] err:%v", err)
			return &types.RespGetTxsByAccountName{}, err
		}
		results = append(results, &types.Tx{
			TxHash:        txInfo.TxHash,
			TxType:        uint32(txInfo.TxType),
			GasFeeAssetId: uint32(txInfo.GasFeeAssetId),
			GasFee:        txInfo.GasFee,
			NftIndex:      uint32(txInfo.NftIndex),
			PairIndex:     uint32(txInfo.PairIndex),
			AssetId:       uint32(txInfo.AssetId),
			TxAmount:      txInfo.TxAmount,
			NativeAddress: txInfo.NativeAddress,
			TxDetails:     txDetails,
			TxInfo:        txInfo.TxInfo,
			ExtraInfo:     txInfo.ExtraInfo,
			Memo:          txInfo.Memo,
			AccountIndex:  uint32(txInfo.AccountIndex),
			Nonce:         uint32(txInfo.Nonce),
			ExpiredAt:     uint32(txInfo.ExpiredAt),
			L2BlockHeight: uint32(txInfo.L2BlockHeight),
			Status:        uint32(txInfo.Status),
			CreatedAt:     uint32(txInfo.CreatedAt.Unix()),
			BlockID:       uint32(blockInfo.ID),
		})
	}
	return &types.RespGetTxsByAccountName{Total: uint32(txCount + mempoolTxCount), Txs: results}, nil
}
