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

type GetTxsByAccountIndexAndTxTypeLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Tx
	globalRpc globalrpc.GlobalRPC
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
}

func NewGetTxsByAccountIndexAndTxTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountIndexAndTxTypeLogic {
	return &GetTxsByAccountIndexAndTxTypeLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		tx:        tx.New(svcCtx.Config),
		globalRpc: globalrpc.New(svcCtx.Config, ctx),
		block:     block.New(svcCtx.Config),
		account:   account.New(svcCtx.Config),
		mempool:   mempool.New(svcCtx.Config),
	}
}

func (l *GetTxsByAccountIndexAndTxTypeLogic) GetTxsByAccountIndexAndTxType(req *types.ReqGetTxsByAccountIndexAndTxType) (resp *types.RespGetTxsByAccountIndexAndTxType, err error) {
	account, err := l.account.GetAccountByPk(req.Pk)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
		return &types.RespGetTxsByAccountIndexAndTxType{}, err
	}
	txCount, err := l.tx.GetTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
		return &types.RespGetTxsByAccountIndexAndTxType{}, err
	}
	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
		return &types.RespGetTxsByAccountIndexAndTxType{}, err
	}
	mempoolTxs, err := l.globalRpc.GetLatestTxsListByAccountIndexAndTxType(uint64(account.AccountIndex), uint64(req.TxType), uint64(req.Limit), uint64(req.Offset))
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountIndexAndTxType] err:%v", err)
		return &types.RespGetTxsByAccountIndexAndTxType{}, err
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
