package transaction

import (
	"context"
	"strconv"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/accounthistory"
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
	account   accounthistory.AccountHistory
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
		account:   accounthistory.New(svcCtx.Config),
		globalRpc: globalrpc.New(svcCtx.Config, ctx),
		tx:        tx.New(svcCtx.Config),
		mempool:   mempool.New(svcCtx.Config),
		block:     block.New(svcCtx.Config),
	}
}

func (l *GetTxsByAccountNameLogic) GetTxsByAccountName(req *types.ReqGetTxsByAccountName) (resp *types.RespGetTxsByAccountName, err error) {

	//err := utils.CheckRequestParam(utils.TypeAccountName, reflect.ValueOf(req.AccountName))

	//err = utils.CheckRequestParam(utils.TypeAccountNameOmitSpace, reflect.ValueOf(accountName))

	account, err := l.account.GetAccountByAccountName(req.AccountName)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountName] err:%v", err)
		return &types.RespGetTxsByAccountName{}, err
	}
	//ReqGetLatestTxsListByAccountIndex
	txList, _, err := l.globalRpc.GetLatestTxsListByAccountIndex(uint32(account.AccountIndex), req.Limit)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountName] err:%v", err)
		return &types.RespGetTxsByAccountName{}, err
	}

	txCount, err := l.tx.GetTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountName] err:%v", err)
		return &types.RespGetTxsByAccountName{}, err
	}

	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndex(account.AccountIndex)
	if err != nil {
		logx.Error("[transaction.GetTxsByAccountName] err:%v", err)
		return &types.RespGetTxsByAccountName{}, err
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
		gasFee, _ := strconv.ParseInt(tx.GasFee, 10, 64)
		txAmount, _ := strconv.ParseInt(tx.TxAmount, 10, 64)
		blockInfo, err := l.block.GetBlockByBlockHeight(tx.L2BlockHeight)
		if err != nil {
			logx.Error("[transaction.GetTxsByAccountName] err:%v", err)
			return &types.RespGetTxsByAccountName{}, err
		}
		results = append(results, &types.Tx{
			TxHash:        tx.TxHash,
			TxType:        uint32(tx.TxType),
			GasFeeAssetId: uint32(tx.GasFeeAssetId),
			GasFee:        gasFee,
			TxStatus:      uint32(tx.Status),
			BlockHeight:   uint32(tx.L2BlockHeight),
			BlockStatus:   uint32(blockInfo.BlockStatus),
			BlockId:       uint32(blockInfo.ID),
			//Todo: still need assetAId and assetBId?
			AssetAId:      uint32(tx.AssetId),
			AssetBId:      uint32(tx.AssetId),
			TxAmount:      uint32(txAmount),
			TxDetails:     txDetails,
			NativeAddress: tx.NativeAddress,
			CreatedAt:     tx.CreatedAt.UnixNano() / 1e6,
			Memo:          tx.Memo,
		})
	}
	return &types.RespGetTxsByAccountName{Total: uint32(txCount + mempoolTxCount), Txs: results}, nil
}
