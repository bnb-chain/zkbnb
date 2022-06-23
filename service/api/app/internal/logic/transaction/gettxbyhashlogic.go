package transaction

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
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
}

func NewGetTxByHashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxByHashLogic {
	return &GetTxByHashLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		mempool: mempool.New(svcCtx),
		block:   block.New(svcCtx),
	}
}

func (l *GetTxByHashLogic) GetTxByHash(req *types.ReqGetTxByHash) (*types.RespGetTxByHash, error) {
	txMemppol, err := l.mempool.GetMempoolTxByTxHash(req.TxHash)
	if err != nil {
		logx.Errorf("[mempool.GetMempoolTxByTxHash]:%v", err)
		return nil, err
	}
	txDetails := make([]*types.TxDetail, 0)
	for _, w := range txMemppol.MempoolDetails {
		txDetails = append(txDetails, &types.TxDetail{
			AssetId:      (w.AssetId),
			AssetType:    (w.AssetType),
			AccountIndex: (w.AccountIndex),
			AccountName:  w.AccountName,
			AccountDelta: w.BalanceDelta,
		})
	}
	block, err := l.block.GetBlockByBlockHeight(txMemppol.L2BlockHeight)
	if err != nil {
		logx.Errorf("[GetBlockByBlockHeight]:%v", err)
		return nil, err
	}
	tx := types.Tx{
		TxHash:        txMemppol.TxHash,
		TxType:        (txMemppol.TxType),
		GasFeeAssetId: (txMemppol.GasFeeAssetId),
		GasFee:        txMemppol.GasFee,
		NftIndex:      (txMemppol.NftIndex),
		PairIndex:     (txMemppol.PairIndex),
		AssetId:       (txMemppol.AssetId),
		TxAmount:      txMemppol.TxAmount,
		NativeAddress: txMemppol.NativeAddress,
		TxDetails:     txDetails,
		TxInfo:        txMemppol.TxInfo,
		ExtraInfo:     txMemppol.ExtraInfo,
		Memo:          txMemppol.Memo,
		AccountIndex:  (txMemppol.AccountIndex),
		Nonce:         (txMemppol.Nonce),
		ExpiredAt:     (txMemppol.ExpiredAt),
		BlockHeight:   (txMemppol.L2BlockHeight),
		Status:        int64(txMemppol.Status),
		CreatedAt:     (txMemppol.CreatedAt.Unix()),
		BlockId:       int64(block.ID),
	}
	return &types.RespGetTxByHash{
		Tx:          tx,
		CommittedAt: block.CommittedAt,
		VerifiedAt:  block.VerifiedAt,
		ExecutedAt:  0,
	}, nil
}
