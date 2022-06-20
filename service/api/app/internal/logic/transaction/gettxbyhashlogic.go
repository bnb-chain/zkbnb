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
			AssetId:      uint32(w.AssetId),
			AssetType:    uint32(w.AssetType),
			AccountIndex: int32(w.AccountIndex),
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
		TxType:        uint32(txMemppol.TxType),
		GasFeeAssetId: uint32(txMemppol.GasFeeAssetId),
		GasFee:        txMemppol.GasFee,
		NftIndex:      uint32(txMemppol.NftIndex),
		PairIndex:     uint32(txMemppol.PairIndex),
		AssetId:       uint32(txMemppol.AssetId),
		TxAmount:      txMemppol.TxAmount,
		NativeAddress: txMemppol.NativeAddress,
		TxDetails:     txDetails,
		TxInfo:        txMemppol.TxInfo,
		ExtraInfo:     txMemppol.ExtraInfo,
		Memo:          txMemppol.Memo,
		AccountIndex:  uint32(txMemppol.AccountIndex),
		Nonce:         uint32(txMemppol.Nonce),
		ExpiredAt:     uint32(txMemppol.ExpiredAt),
		L2BlockHeight: uint32(txMemppol.L2BlockHeight),
		Status:        uint32(txMemppol.Status),
		CreatedAt:     uint32(txMemppol.CreatedAt.Unix()),
		BlockID:       uint32(block.ID),
	}
	return &types.RespGetTxByHash{
		Tx:          tx,
		CommittedAt: block.CommittedAt,
		VerifiedAt:  block.VerifiedAt,
		ExecutedAt:  0,
	}, nil
}
