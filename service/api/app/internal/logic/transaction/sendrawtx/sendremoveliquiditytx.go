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

type removeLiquidityTxSender struct {
	txType          int
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	gasChecker      GasChecker
	nonceChecker    NonceChecker
	mempoolTxSender MempoolTxSender
}

func NewRemoveLiquidityTxSender(ctx context.Context, svcCtx *svc.ServiceContext,
	gasChecker *gasChecker, nonceChecker *nonceChecker, sender *mempoolTxSender) *removeLiquidityTxSender {
	return &removeLiquidityTxSender{
		txType:          commonTx.TxTypeRemoveLiquidity,
		ctx:             ctx,
		svcCtx:          svcCtx,
		gasChecker:      gasChecker,
		nonceChecker:    nonceChecker,
		mempoolTxSender: sender,
	}
}
func (s *removeLiquidityTxSender) SendTx(rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseRemoveLiquidityTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.AppErrInvalidTx
	}

	if err := txInfo.Validate(); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	//check expire time
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("invalid ExpiredAt")
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid ExpiredAt")
	}

	//check signature
	accountPk, err := s.svcCtx.MemCache.GetAccountPkByIndex(txInfo.FromAccountIndex)
	if err != nil {
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("unknown FromAccountIndex")
			}
			return "", errorcode.AppErrInternal
		}
	}
	if err := txInfo.VerifySignature(accountPk); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid Signature")
	}

	//check liquidity
	liquidity, err := s.svcCtx.StateFetcher.GetLatestLiquidity(s.ctx, txInfo.PairIndex)
	if err != nil {
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("invalid PairIndex")
			}
			logx.Errorf(" unable to get latest liquidity: %d, err: %s", txInfo.PairIndex, err.Error())
			return "", errorcode.AppErrInternal
		}
	}

	// check params
	if liquidity.AssetA == nil ||
		liquidity.AssetA.Cmp(big.NewInt(0)) == 0 ||
		liquidity.AssetB == nil ||
		liquidity.AssetB.Cmp(big.NewInt(0)) == 0 ||
		liquidity.LpAmount == nil ||
		liquidity.LpAmount.Cmp(big.NewInt(0)) == 0 {
		logx.Errorf("invalid asset or lp amount")
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid asset or lp amount")
	}

	var (
		assetAAmount, assetBAmount *big.Int
	)
	assetAAmount, assetBAmount, err = util.ComputeRemoveLiquidityAmount(liquidity, txInfo.LpAmount)
	if err != nil {
		logx.Errorf("fail to compute remove liquidity amount, err: %s", err.Error())
		return "", err
	}
	if assetAAmount.Cmp(txInfo.AssetAMinAmount) < 0 || assetBAmount.Cmp(txInfo.AssetBMinAmount) < 0 {
		logx.Errorf("remove liqudity asset amounts is less than min amount: %s:%s/%s:%s",
			txInfo.AssetAMinAmount.String(), txInfo.AssetBMinAmount.String(), assetAAmount.String(), assetBAmount.String())
		return "", errorcode.AppErrInvalidTxField.RefineError("AssetAMinAmount or AssetBMinAmount cannot by satisfied")
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
		TxAmount:      "",
		Memo:          "",
		AccountIndex:  txInfo.FromAccountIndex,
		Nonce:         txInfo.Nonce,
		ExpiredAt:     txInfo.ExpiredAt,
	}
	txId, err = s.mempoolTxSender.SendMempoolTx(func(txInfo interface{}) ([]byte, error) {
		return legendTxTypes.ComputeRemoveLiquidityMsgHash(txInfo.(*legendTxTypes.RemoveLiquidityTxInfo), mimc.NewMiMC())
	}, txInfo, mempoolTx)

	return txId, err
}
