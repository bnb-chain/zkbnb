package transaction

import (
	"context"
	"strconv"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMempoolTxsListLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Tx
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
}

func NewGetMempoolTxsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsListLogic {
	return &GetMempoolTxsListLogic{
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

func (l *GetMempoolTxsListLogic) GetMempoolTxsList(req *types.ReqGetMempoolTxsList) (resp *types.RespGetMempoolTxsList, err error) {
	//	err = utils.CheckRequestParam(utils.TypeLimit, reflect.ValueOf(req.Limit))
	mempoolTxs, err := l.mempool.GetMempoolTxs(int64(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Error("[GetMempoolTxs] err:%v", err)
		return
	}

	// Todo: why not do total=len(mempoolTxs)
	total, err := l.mempool.GetMempoolTxsTotalCount()
	if err != nil {
		logx.Error("[GetMempoolTxs] err:%v", err)
		return
	}

	for _, mempoolTx := range mempoolTxs {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range mempoolTx.MempoolDetails {

			if txDetail.AssetType == commonAsset.LiquidityAssetType {
				//Todo: add json string of liquidity transaction to the list
			} else {
				txDetails = append(txDetails, &types.TxDetail{
					//Todo: verify if accountBalance is still needed, since its no longer a field of table TxDetail
					//Todo: int64 or int?
					//Todo: need balance or not?  no need
					AssetId:      int(txDetail.AssetId),
					AssetType:    int(txDetail.AssetType),
					AccountIndex: int32(txDetail.AccountIndex),
					AccountName:  txDetail.AccountName,
					AccountDelta: txDetail.BalanceDelta,
				})
			}
		}
		//Todo: int64 or int?
		var txAmount, gasFee int
		txAmount, err = strconv.Atoi(mempoolTx.TxAmount)
		if err != nil {
			logx.Error("[GetMempoolTxs] err:%v", err)
			return
		}
		// Todo: why is the field in db string?
		gasFee, err = strconv.Atoi(mempoolTx.GasFee)
		resp.Txs = append(resp.Txs, &types.Tx{
			TxHash:        mempoolTx.TxHash,
			TxType:        int32(mempoolTx.TxType),
			AssetAId:      int32(mempoolTx.AssetId),
			AssetBId:      int32(mempoolTx.AssetId),
			TxDetails:     txDetails,
			TxAmount:      int64(txAmount),
			NativeAddress: mempoolTx.NativeAddress,
			TxStatus:      1, //pending
			GasFeeAssetId: int32(mempoolTx.GasFeeAssetId),
			GasFee:        int32(gasFee),
			CreatedAt:     mempoolTx.CreatedAt.Unix(),
			Memo:          mempoolTx.Memo,
		})
	}
	resp.Total = uint32(total)
	return
}
