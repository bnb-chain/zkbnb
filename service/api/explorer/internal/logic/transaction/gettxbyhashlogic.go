package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxByHashLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	tx        tx.Tx
	block     block.Block
	account   account.AccountModel
	mempool   mempool.Mempool
	globalRPC globalrpc.GlobalRPC
}

func NewGetTxByHashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxByHashLogic {
	return &GetTxByHashLogic{
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

func (l *GetTxByHashLogic) GetTxByHash(req *types.ReqGetTxByHash) (*types.RespGetTxByHash, error) {
	resp := &types.RespGetTxByHash{}
	txMemppol, err := l.mempool.GetMempoolTxByTxHash(req.TxHash)
	if err != nil {
		l.Error("[mempool.GetMempoolTxByTxHash]:%v", err)
		return nil, err
	}

	txDetails := make([]*types.TxDetail, 0)
	for _, w := range txMemppol.MempoolDetails {
		txDetails = append(txDetails, &types.TxDetail{
			AssetId:      w.AssetId,
			AssetType:    w.AssetType,
			AccountIndex: w.AccountIndex,
			AccountName:  w.AccountName,
			BalanceDelta: w.BalanceDelta,
		})
	}
	block, err := l.block.GetBlockByBlockHeight(txMemppol.L2BlockHeight)
	if err != nil {
		l.Error("[Block.GetBlockByBlockHeight]:%v", err)
		return nil, err
	}

	// Todo: update blockstatus to cache, but not sure if the whole block shall be inserted. which kind of tx? mempoolTx or tx?
	//err = l.svcCtx.BlockModel.UpdateBlockStatusCacheByBlockHeight(tx.BlockHeight, blockStatusInfo)
	//if err != nil {
	//	errInfo := fmt.Sprintf("[appService.tx.GetTxByHash]<=>[BlockModel.UpdateBlockStatusCacheByBlockHeight] %s", err.Error())
	//	logx.Error(errInfo)
	//	return packGetTxByHashResp(types.FailStatus, "fail", errInfo, respResult), nil
	//}
	resp.Txs = types.Tx{
		TxHash:        txMemppol.TxHash,
		TxType:        txMemppol.TxType,
		GasFee:        txMemppol.GasFee,
		GasFeeAssetId: txMemppol.GasFeeAssetId,
		TxStatus:      int64(txMemppol.Status),
		BlockHeight:   txMemppol.L2BlockHeight,
		BlockId:       int64(block.ID),
		TxAmount:      txMemppol.TxAmount,
		TxDetails:     txDetails,
		NativeAddress: txMemppol.NativeAddress,
		Memo:          txMemppol.Memo,

		// Todo: where is executedAt field from?
		// -> gavin
	}

	return resp, nil
}
