package sendrawtx

import (
	"context"
	"math/big"
	"time"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type swapTxSender struct {
	txType          int
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	gasChecker      GasChecker
	nonceChecker    NonceChecker
	mempoolTxSender MempoolTxSender
}

func NewSwapTxSender(ctx context.Context, svcCtx *svc.ServiceContext,
	gasChecker *gasChecker, nonceChecker *nonceChecker, sender *mempoolTxSender) *swapTxSender {
	return &swapTxSender{
		txType:          commonTx.TxTypeSwap,
		ctx:             ctx,
		svcCtx:          svcCtx,
		gasChecker:      gasChecker,
		nonceChecker:    nonceChecker,
		mempoolTxSender: sender,
	}
}
func (s *swapTxSender) SendTx(rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseSwapTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.AppErrInvalidTx
	}

	if err := legendTxTypes.ValidateSwapTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	//check expire time
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("invalid ExpiredAt")
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid ExpiredAt")
	}

	//TODO: check signature

	liquidity, err := s.svcCtx.StateFetcher.GetLatestLiquidity(s.ctx, txInfo.PairIndex)
	if err != nil {
		logx.Errorf(" unable to get latest liquidity info for write: %s", err.Error())
		return "", errorcode.AppErrInternal
	}

	// check params
	if liquidity.AssetA == nil || liquidity.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidity.AssetB == nil || liquidity.AssetB.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("liquidity pool %d is empty", liquidity.PairIndex)
		return "", errorcode.AppErrInvalidParam.RefineError("liquidity pool is empty")
	}

	// compute delta
	var (
		toDelta *big.Int
	)
	if liquidity.AssetAId == txInfo.AssetAId &&
		liquidity.AssetBId == txInfo.AssetBId {
		toDelta, _, err = util.ComputeDelta(
			liquidity.AssetA,
			liquidity.AssetB,
			liquidity.AssetAId,
			liquidity.AssetBId,
			txInfo.AssetAId,
			true,
			txInfo.AssetAAmount,
			liquidity.FeeRate,
		)
	} else if liquidity.AssetAId == txInfo.AssetBId &&
		liquidity.AssetBId == txInfo.AssetAId {
		toDelta, _, err = util.ComputeDelta(
			liquidity.AssetA,
			liquidity.AssetB,
			liquidity.AssetAId,
			liquidity.AssetBId,
			txInfo.AssetBId,
			true,
			txInfo.AssetAAmount,
			liquidity.FeeRate,
		)
	} else {
		logx.Errorf("invalid AssetIds: %d %d %d, err: %s",
			txInfo.AssetAId,
			uint32(liquidity.AssetAId),
			uint32(liquidity.AssetBId),
			err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid AssetAId/AssetBId")
	}

	// check amount
	if toDelta.Cmp(txInfo.AssetBMinAmount) < 0 {
		logx.Errorf("received amount: %s is less than min amount: %s", txInfo.AssetBMinAmount.String(), toDelta.String())
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid AssetBMinAmount")
	}

	//check from account
	fromAccount, err := s.svcCtx.StateFetcher.GetLatestAccount(s.ctx, txInfo.FromAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.FromAccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}

	//check gas
	if err := s.gasChecker.CheckGas(fromAccount, txInfo.GasAccountIndex, txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError(err.Error())
	}

	//check nonce
	if err := s.nonceChecker.CheckNonce(fromAccount, txInfo.Nonce); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError(err.Error())
	}

	//send mempool tx
	mempoolTx := &mempool.MempoolTx{
		TxType:        int64(s.txType),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     txInfo.PairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      txInfo.AssetAAmount.String(),
		Memo:          "",
		AccountIndex:  txInfo.FromAccountIndex,
		Nonce:         txInfo.Nonce,
		ExpiredAt:     txInfo.ExpiredAt,
	}
	txId, err = s.mempoolTxSender.SendMempoolTx(func(txInfo interface{}) ([]byte, error) {
		return legendTxTypes.ComputeSwapMsgHash(txInfo.(*legendTxTypes.SwapTxInfo), mimc.NewMiMC())
	}, txInfo, mempoolTx)

	return txId, err
}
