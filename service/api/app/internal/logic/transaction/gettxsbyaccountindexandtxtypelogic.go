package transaction

import (
	"context"
	"strconv"

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

	//err = utils.CheckRequestParam(utils.TypeAccountIndex, reflect.ValueOf(req.AccountIndex))
	//err = utils.CheckRequestParam(utils.TypeTxType, reflect.ValueOf(req.TxType))
	//err = utils.CheckRequestParam(utils.TypeLimit, reflect.ValueOf(req.Limit))

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
		results = append(results, &types.Tx{
			TxHash:        tx.TxHash,
			TxType:        uint32(tx.TxType),
			GasFeeAssetId: uint32(tx.GasFeeAssetId),
			GasFee:        gasFee,
			TxStatus:      tx.Status,
			BlockHeight:   int(tx.L2BlockHeight),
			BlockStatus:   int(blockInfo.BlockStatus),
			BlockId:       int(blockInfo.ID),
			//Todo: still need AssetAId, AssetBId?
			AssetAId:      int(tx.AssetId),
			AssetBId:      int(tx.AssetId),
			TxAmount:      txAmount,
			TxDetails:     txDetails,
			NativeAddress: tx.NativeAddress,
			CreatedAt:     tx.CreatedAt.UnixNano() / 1e6,
			Memo:          tx.Memo,
		})
	}

	return &types.RespGetTxsByAccountIndexAndTxType{Total: uint32(txCount + mempoolTxCount), Txs: results}, nil
}
