package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zecrey-labs/zecrey-legend/utils"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMempoolTxsLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	mempool mempool.Mempool
	block   block.Block
}

func NewGetMempoolTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsLogic {
	return &GetMempoolTxsLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		mempool: mempool.New(svcCtx),
		block:   block.New(svcCtx),
	}
}
func (l *GetMempoolTxsLogic) GetMempoolTxs(req *types.ReqGetMempoolTxs) (*types.RespGetMempoolTxs, error) {
	if utils.CheckTypeLimit(req.Limit) {
		logx.Errorf("[CheckTypeLimit] param:%v", req.Limit)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckTypeOffset(req.Offset) {
		logx.Errorf("[CheckTypeOffset] param:%v", req.Offset)
		return nil, errcode.ErrInvalidParam
	}
	count, err := l.mempool.GetMempoolTxsTotalCount()
	if err != nil {
		logx.Errorf("[GetMempoolTxsTotalCount] err:%v", err)
		return nil, err
	}
	resp := &types.RespGetMempoolTxs{
		MempoolTxs: make([]*types.Tx, 0),
		Total:      uint32(count),
	}
	logx.Errorf("[GetMempoolTxsTotalCount] count:%v", count)

	mempoolTxs, err := l.mempool.GetMempoolTxs(int64(req.Limit), int64(req.Offset))
	if err != nil {
		logx.Errorf("[GetMempoolTxs] err:%v", err)
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range mempoolTx.MempoolDetails {
			txDetails = append(txDetails, &types.TxDetail{
				AssetId:      (txDetail.AssetId),
				AssetType:    (txDetail.AssetType),
				AccountIndex: (txDetail.AccountIndex),
				AccountName:  txDetail.AccountName,
			})
		}
		blockInfo, err := l.block.GetBlockByBlockHeight(l.ctx, mempoolTx.L2BlockHeight)
		if err != nil {
			logx.Error("[GetBlockByBlockHeight] err:%v", err)
			return nil, err
		}
		resp.MempoolTxs = append(resp.MempoolTxs, &types.Tx{
			TxHash:        mempoolTx.TxHash,
			TxType:        (mempoolTx.TxType),
			GasFeeAssetId: (mempoolTx.GasFeeAssetId),
			GasFee:        mempoolTx.GasFee,
			NftIndex:      (mempoolTx.NftIndex),
			PairIndex:     (mempoolTx.PairIndex),
			AssetId:       (mempoolTx.AssetId),
			TxAmount:      mempoolTx.TxAmount,
			NativeAddress: mempoolTx.NativeAddress,
			TxDetails:     txDetails,
			TxInfo:        mempoolTx.TxInfo,
			ExtraInfo:     mempoolTx.ExtraInfo,
			Memo:          mempoolTx.Memo,
			AccountIndex:  (mempoolTx.AccountIndex),
			Nonce:         (mempoolTx.Nonce),
			ExpiredAt:     (mempoolTx.ExpiredAt),
			BlockHeight:   (mempoolTx.L2BlockHeight),
			Status:        int64(mempoolTx.Status),
			CreatedAt:     (mempoolTx.CreatedAt.Unix()),
			BlockId:       int64(blockInfo.ID),
		})
	}
	return resp, nil
}
