package transaction

import (
	"context"

	blockModel "github.com/zecrey-labs/zecrey-legend/common/model/block"

	"strconv"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxByHashLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	mempool mempool.Mempool
	block   block.Block
	tx      tx.Tx
}

func NewGetTxByHashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxByHashLogic {
	return &GetTxByHashLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		mempool: mempool.New(svcCtx.Config),
		block:   block.New(svcCtx.Config),
		tx:      tx.New(svcCtx.Config),
	}
}

func packGetTxByHashResp(tx types.Tx, committedAt int64, verifiedAt int64, executedAt int64) (res *types.RespGetTxByHash) {
	return &types.RespGetTxByHash{
		Tx:          tx,
		CommittedAt: committedAt,
		VerifiedAt:  verifiedAt,
		ExecutedAt:  executedAt,
	}
}

func (l *GetTxByHashLogic) GetTxByHash(req *types.ReqGetTxByHash) (resp *types.RespGetTxByHash, err error) {
	//	err := utils.CheckRequestParam(utils.TypeHash, reflect.ValueOf(req.TxHash))

	txMemppol, err := l.mempool.GetMempoolTxByTxHash(req.TxHash)
	if err != nil {
		logx.Error("[mempool.GetMempoolTxByTxHash]:%v", err)
		return nil, err
	}
	txDetails := make([]*types.TxDetail, 0)
	for _, w := range txMemppol.MempoolDetails {
		txDetails = append(txDetails, &types.TxDetail{
			AssetId:      uint32(w.AssetId),
			AssetType:    uint32(w.AssetType),
			AccountIndex: int32(w.AccountIndex),
			AccountName:  w.AccountName,
			AccountDelta: w.BalanceDelta,
		})
	}
	block, err := l.block.GetBlockByBlockHeight(txMemppol.L2BlockHeight)
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
	return packGetTxByHashResp(types.Tx{
		TxHash:        txMemppol.TxHash,
		TxType:        uint32(txMemppol.TxType),
		GasFee:        int64(gasFee),
		GasFeeAssetId: uint32(txMemppol.GasFeeAssetId),
		TxStatus:      uint32(txMemppol.Status),
		BlockHeight:   uint32(txMemppol.L2BlockHeight),
		BlockStatus:   uint32(blockStatusInfo.BlockStatus),
		BlockId:       uint32(block.ID),
		//Todo: globalRPC won't return data with 2 asset ids, where are these fields from
		AssetAId:      uint32(txMemppol.AssetId),
		AssetBId:      uint32(txMemppol.AssetId),
		TxAmount:      uint32(txAmount),
		TxDetails:     txDetails,
		NativeAddress: txMemppol.NativeAddress,
		CreatedAt:     txMemppol.CreatedAt.UnixNano() / 1e6,
		Memo:          txMemppol.Memo,

		// Todo: where is executedAt field from?
		// -> gavin
	}, blockStatusInfo.CommittedAt, blockStatusInfo.VerifiedAt, 0), nil
}
