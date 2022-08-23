package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/core"
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
		return resp, errorcode.AppErrInvalidTx
	}
	if err := executor.Prepare(); err != nil {
		return resp, err
	}
	if err := executor.VerifyInputs(); err != nil {
		return resp, errorcode.AppErrInvalidTxField.RefineError(err.Error())
	}

	mempoolTx, err := executor.GenerateMempoolTx()
	if err != nil {
		return resp, errorcode.AppErrInternal
	}
	if err := s.svcCtx.MempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{mempoolTx}); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		failTx := &tx.FailTx{
			TxHash:    mempoolTx.TxHash,
			TxType:    mempoolTx.TxType,
			TxStatus:  tx.StatusFail,
			AssetAId:  commonConstant.NilAssetId,
			AssetBId:  commonConstant.NilAssetId,
			TxAmount:  commonConstant.NilAssetAmountStr,
			TxInfo:    req.TxInfo,
			ExtraInfo: err.Error(),
			Memo:      "",
		}
		_ = s.svcCtx.FailTxModel.CreateFailTx(failTx)
		return resp, errorcode.AppErrInternal
	}

	resp.TxHash = mempoolTx.TxHash
	return resp, nil
}

func (s *SendTxLogic) getExecutor(txType int, txInfo string) (core.TxExecutor, error) {
	bc := core.NewBlockChainForDryRun(s.svcCtx.AccountModel, s.svcCtx.LiquidityModel, s.svcCtx.NftModel, s.svcCtx.MempoolModel,
		s.svcCtx.RedisCache)
	t := &tx.Tx{TxType: int64(txType), TxInfo: txInfo}

	switch txType {
	case commonTx.TxTypeTransfer:
		return core.NewTransferExecutor(bc, t)
	case commonTx.TxTypeSwap:
		return core.NewSwapExecutor(bc, t)
	case commonTx.TxTypeAddLiquidity:
		return core.NewAddLiquidityExecutor(bc, t)
	case commonTx.TxTypeRemoveLiquidity:
		return core.NewRemoveLiquidityExecutor(bc, t)
	case commonTx.TxTypeWithdraw:
		return core.NewWithdrawExecutor(bc, t)
	case commonTx.TxTypeTransferNft:
		return core.NewTransferNftExecutor(bc, t)
	case commonTx.TxTypeAtomicMatch:
		return core.NewAtomicMatchExecutor(bc, t)
	case commonTx.TxTypeCancelOffer:
		return core.NewCancelOfferExecutor(bc, t)
	case commonTx.TxTypeWithdrawNft:
		return core.NewWithdrawNftExecutor(bc, t)
	default:
		logx.Errorf("invalid tx type: %s", txType)
		return nil, errorcode.AppErrInvalidTxType
	}
}
