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
	blockStatusInfo := &blockModel.BlockStatusInfo{
		BlockStatus: block.BlockStatus,
		CommittedAt: block.CommittedAt,
		VerifiedAt:  block.VerifiedAt,
	}
	txAmount, _ := strconv.Atoi(txMemppol.TxAmount)
	return packGetTxByHashResp(types.Tx{
		TxHash :        txMemppol.TxHash,
		TxType  :        uint32(txMemppol.TxType),
		GasFeeAssetId  : uint32(txMemppol.GasFeeAssetId),
		GasFee         :        txMemppol.GasFee,
		NftIndex       txMemppol.NftIndex,
		PairIndex      int64
		AssetId        :       uint32(txMemppol.AssetId),
		TxAmount       :      uint32(txAmount),
		NativeAddress  : txMemppol.NativeAddress,
		MempoolDetails []*MempoolTxDetail `gorm:"foreignkey:TxId"`
		TxInfo         string
		ExtraInfo      string
		Memo:          txMemppol.Memo,
		AccountIndex   int64
		Nonce          int64
		ExpiredAt      int64
		L2BlockHeight  :   uint32(txMemppol.L2BlockHeight),
		Status         int `gorm:"index"` // 0: pending tx; 1: committed tx; 2: verified tx;

		// Todo: where is executedAt field from?
		// -> gavin
	}, blockStatusInfo.CommittedAt, blockStatusInfo.VerifiedAt, 0), nil
}
