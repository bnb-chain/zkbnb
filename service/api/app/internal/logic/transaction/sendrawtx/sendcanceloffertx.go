package sendrawtx

import (
	"context"
	"math/big"
	"time"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type cancelOfferTxSender struct {
	txType          int
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	gasChecker      GasChecker
	nonceChecker    NonceChecker
	mempoolTxSender MempoolTxSender
}

func NewCancelTxSender(ctx context.Context, svcCtx *svc.ServiceContext,
	gasChecker *gasChecker, nonceChecker *nonceChecker, sender *mempoolTxSender) *cancelOfferTxSender {
	return &cancelOfferTxSender{
		txType:          commonTx.TxTypeCancelOffer,
		ctx:             ctx,
		svcCtx:          svcCtx,
		gasChecker:      gasChecker,
		nonceChecker:    nonceChecker,
		mempoolTxSender: sender,
	}
}
func (s *cancelOfferTxSender) SendTx(rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseCancelOfferTxInfo(rawTxInfo)
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
	accountPk, err := s.svcCtx.MemCache.GetAccountPkByIndex(txInfo.AccountIndex)
	if err != nil {
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("unknown AccountIndex")
			}
			return "", errorcode.AppErrInternal
		}
	}
	if err := txInfo.VerifySignature(accountPk); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid Signature")
	}

	fromAccount, err := s.svcCtx.StateFetcher.GetLatestAccount(txInfo.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid AccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.AccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}

	offerAssetId := txInfo.OfferId / 128
	offerIndex := txInfo.OfferId % 128
	if fromAccount.AssetInfo[offerAssetId] == nil {
		fromAccount.AssetInfo[offerAssetId] = &commonAsset.AccountAsset{
			AssetId:                  offerAssetId,
			Balance:                  big.NewInt(0),
			LpAmount:                 big.NewInt(0),
			OfferCanceledOrFinalized: big.NewInt(0),
		}
	} else {
		offer := fromAccount.AssetInfo[offerAssetId].OfferCanceledOrFinalized
		if offer.Bit(int(offerIndex)) == 1 {
			logx.Errorf("offer is already confirmed or canceled")
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid OfferId, already confirmed or canceled")
		}
	}

	//check gas
	if err := s.gasChecker.CheckGas(fromAccount, txInfo.GasAccountIndex, txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError(err.Error())
	}

	//check nonce
	if err := s.nonceChecker.CheckNonce(fromAccount.AccountIndex, txInfo.Nonce); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError(err.Error())
	}

	//send mempool tx
	mempoolTx := &mempool.MempoolTx{
		TxType:        int64(s.txType),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      "",
		Memo:          "",
		AccountIndex:  txInfo.AccountIndex,
		Nonce:         txInfo.Nonce,
		ExpiredAt:     txInfo.ExpiredAt,
	}
	txId, err = s.mempoolTxSender.SendMempoolTx(func(txInfo interface{}) ([]byte, error) {
		return legendTxTypes.ComputeAtomicMatchMsgHash(txInfo.(*legendTxTypes.AtomicMatchTxInfo), mimc.NewMiMC())
	}, txInfo, mempoolTx)

	return txId, err
}
