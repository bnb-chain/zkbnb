package transaction

import (
	"context"
	"github.com/bnb-chain/zkbnb/dao/dbcache"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type SendTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (s *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.TxHash, err error) {
	pendingTxCount, err := s.svcCtx.MemCache.GetTxPendingCountKeyPrefix(func() (interface{}, error) {
		txStatuses := []int64{tx.StatusPending}
		return s.svcCtx.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
	})
	if err != nil {
		return nil, types2.AppErrInternal
	}

	if s.svcCtx.Config.TxPool.MaxPendingTxCount > 0 && pendingTxCount >= int64(s.svcCtx.Config.TxPool.MaxPendingTxCount) {
		return nil, types2.AppErrTooManyTxs
	}

	resp = &types.TxHash{}
	bc, err := core.NewBlockChainForDryRun(s.svcCtx.AccountModel, s.svcCtx.NftModel, s.svcCtx.TxPoolModel,
		s.svcCtx.AssetModel, s.svcCtx.SysConfigModel, s.svcCtx.RedisCache)
	if err != nil {
		logx.Error("fail to init blockchain runner:", err)
		return nil, types2.AppErrInternal
	}
	newPoolTx := tx.PoolTx{
		TxHash: types2.EmptyTxHash, // Would be computed in prepare method of executors.
		TxType: int64(req.TxType),
		TxInfo: req.TxInfo,

		GasFeeAssetId: types2.NilAssetId,
		GasFee:        types2.NilAssetAmount,
		NftIndex:      types2.NilNftIndex,
		CollectionId:  types2.NilCollectionNonce,
		AssetId:       types2.NilAssetId,
		TxAmount:      types2.NilAssetAmount,
		NativeAddress: types2.EmptyL1Address,

		BlockHeight: types2.NilBlockHeight,
		TxStatus:    tx.StatusPending,
	}
	newTx := &tx.Tx{PoolTx: newPoolTx}
	err = bc.ApplyTransaction(newTx)
	if err != nil {
		return resp, err
	}
	newTx.PoolTx.TxType = int64(req.TxType)
	newTx.PoolTx.TxInfo = req.TxInfo
	newTx.PoolTx.BlockHeight = types2.NilBlockHeight
	newTx.PoolTx.TxStatus = tx.StatusPending
	if newTx.PoolTx.TxType == types2.TxTypeMintNft {
		newTx.PoolTx.NftIndex = types2.NilNftIndex
	}
	if newTx.PoolTx.TxType == types2.TxTypeCreateCollection {
		newTx.PoolTx.CollectionId = types2.NilCollectionNonce
	}
	if err := s.svcCtx.TxPoolModel.CreateTxs([]*tx.PoolTx{&newTx.PoolTx}); err != nil {
		logx.Errorf("fail to create pool tx: %v, err: %s", newTx, err.Error())
		return resp, types2.AppErrInternal
	}
	s.svcCtx.RedisCache.Set(context.Background(), dbcache.AccountNonceKeyByIndex(newTx.AccountIndex), newTx.Nonce)
	resp.TxHash = newTx.TxHash
	return resp, nil
}
