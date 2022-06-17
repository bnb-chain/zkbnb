package transaction

import (
	"context"
	"strconv"

	blockModel "github.com/zecrey-labs/zecrey-legend/common/model/block"
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

func (l *GetTxByHashLogic) GetTxByHash(req *types.ReqGetTxByHash) (resp *types.RespGetTxByHash, err error) {
	txMemppol, err := l.mempool.GetMempoolTxByTxHash(req.TxHash)
	if err != nil {
		l.Error("[mempool.GetMempoolTxByTxHash]:%v", err)
		return
	}

	txDetails := make([]*types.TxDetail, 0)
	for _, w := range txMemppol.MempoolDetails {
		txDetails = append(txDetails, &types.TxDetail{
			AssetId:      int(w.AssetId),
			AssetType:    int(w.AssetType),
			AccountIndex: int32(w.AccountIndex),
			AccountName:  w.AccountName,
			AccountDelta: w.BalanceDelta,
		})
	}
	block, err := l.block.GetBlockByBlockHeight(txMemppol.L2BlockHeight)
	if err != nil {
		l.Error("[Block.GetBlockByBlockHeight]:%v", err)
		return
	}

	blockStatusInfo := &blockModel.BlockStatusInfo{
		BlockStatus: block.BlockStatus,
		CommittedAt: block.CommittedAt,
		VerifiedAt:  block.VerifiedAt,
	}

	// Todo: update blockstatus to cache, but not sure if the whole block shall be inserted. which kind of tx? mempoolTx or tx?
	//err = l.svcCtx.BlockModel.UpdateBlockStatusCacheByBlockHeight(tx.BlockHeight, blockStatusInfo)
	//if err != nil {
	//	errInfo := fmt.Sprintf("[appService.tx.GetTxByHash]<=>[BlockModel.UpdateBlockStatusCacheByBlockHeight] %s", err.Error())
	//	logx.Error(errInfo)
	//	return packGetTxByHashResp(types.FailStatus, "fail", errInfo, respResult), nil
	//}
	gasFee, _ := strconv.Atoi(txMemppol.GasFee)
	txAmount, _ := strconv.Atoi(txMemppol.TxAmount)

	resp.Txs = types.Tx{
		TxHash:        txMemppol.TxHash,
		TxType:        int32(txMemppol.TxType),
		GasFee:        int32(gasFee),
		GasFeeAssetId: int32(txMemppol.GasFeeAssetId),
		TxStatus:      int32(txMemppol.Status),
		BlockHeight:   int64(txMemppol.L2BlockHeight),
		BlockStatus:   int32(blockStatusInfo.BlockStatus),
		BlockId:       int32(block.ID),
		//Todo: globalRPC won't return data with 2 asset ids, where are these fields from
		AssetAId:      int32(txMemppol.AssetId),
		AssetBId:      int32(txMemppol.AssetId),
		TxAmount:      int64(txAmount),
		TxDetails:     txDetails,
		NativeAddress: txMemppol.NativeAddress,
		CreatedAt:     txMemppol.CreatedAt.UnixNano() / 1e6,
		Memo:          txMemppol.Memo,

		// Todo: where is executedAt field from?
		// -> gavin
	}

	return
}
