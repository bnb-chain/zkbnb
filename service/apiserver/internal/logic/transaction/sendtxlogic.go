package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/core"
	"github.com/bnb-chain/zkbas/core/executor"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbas/types"
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
	resp = &types.TxHash{}
	executor, err := s.getExecutor(int(req.TxType), req.TxInfo)
	if err != nil {
		return resp, types2.AppErrInvalidTx
	}
	if err := executor.Prepare(); err != nil {
		return resp, err
	}
	if err := executor.VerifyInputs(); err != nil {
		return resp, types2.AppErrInvalidTxField.RefineError(err.Error())
	}

	mempoolTx, err := executor.GenerateMempoolTx()
	if err != nil {
		return resp, types2.AppErrInternal
	}
	if err := s.svcCtx.MempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{mempoolTx}); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		failTx := &tx.FailTx{
			TxHash:    mempoolTx.TxHash,
			TxType:    mempoolTx.TxType,
			TxStatus:  tx.StatusFail,
			AssetAId:  types2.NilAssetId,
			AssetBId:  types2.NilAssetId,
			TxAmount:  types2.NilAssetAmountStr,
			TxInfo:    req.TxInfo,
			ExtraInfo: err.Error(),
			Memo:      "",
		}
		_ = s.svcCtx.FailTxModel.CreateFailTx(failTx)
		return resp, types2.AppErrInternal
	}

	resp.TxHash = mempoolTx.TxHash
	return resp, nil
}

func (s *SendTxLogic) getExecutor(txType int, txInfo string) (executor.TxExecutor, error) {
	bc := core.NewBlockChainForDryRun(s.svcCtx.AccountModel, s.svcCtx.LiquidityModel, s.svcCtx.NftModel, s.svcCtx.MempoolModel,
		s.svcCtx.RedisCache)
	t := &tx.Tx{TxType: int64(txType), TxInfo: txInfo}

	switch txType {
	case types2.TxTypeTransfer:
		return executor.NewTransferExecutor(bc, t)
	case types2.TxTypeSwap:
		return executor.NewSwapExecutor(bc, t)
	case types2.TxTypeAddLiquidity:
		return executor.NewAddLiquidityExecutor(bc, t)
	case types2.TxTypeRemoveLiquidity:
		return executor.NewRemoveLiquidityExecutor(bc, t)
	case types2.TxTypeWithdraw:
		return executor.NewWithdrawExecutor(bc, t)
	case types2.TxTypeTransferNft:
		return executor.NewTransferNftExecutor(bc, t)
	case types2.TxTypeAtomicMatch:
		return executor.NewAtomicMatchExecutor(bc, t)
	case types2.TxTypeCancelOffer:
		return executor.NewCancelOfferExecutor(bc, t)
	case types2.TxTypeWithdrawNft:
		return executor.NewWithdrawNftExecutor(bc, t)
	default:
		logx.Errorf("invalid tx type: %s", txType)
		return nil, types2.AppErrInvalidTxType
	}
}
