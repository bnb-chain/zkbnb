package sendrawtx

import (
	"context"
	"time"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

type atomicMatchTxSender struct {
	txType          int
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	gasChecker      GasChecker
	nonceChecker    NonceChecker
	mempoolTxSender MempoolTxSender
}

func NewAtomicMatchTxSender(ctx context.Context, svcCtx *svc.ServiceContext,
	gasChecker *gasChecker, nonceChecker *nonceChecker, sender *mempoolTxSender) *atomicMatchTxSender {
	return &atomicMatchTxSender{
		txType:          commonTx.TxTypeAtomicMatch,
		ctx:             ctx,
		svcCtx:          svcCtx,
		gasChecker:      gasChecker,
		nonceChecker:    nonceChecker,
		mempoolTxSender: sender,
	}
}
func (s *atomicMatchTxSender) SendTx(rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(rawTxInfo)
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
				return "", errorcode.AppErrInvalidTxField.RefineError("unknown FromAccountIndex")
			}
			return "", errorcode.AppErrInternal
		}
	}
	if err := txInfo.VerifySignature(accountPk); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	buyerPk, err := s.svcCtx.MemCache.GetAccountPkByIndex(txInfo.BuyOffer.AccountIndex)
	if err != nil {
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("unknown AccountIndex of BuyOffer")
			}
			return "", errorcode.AppErrInternal
		}
	}
	if err := txInfo.BuyOffer.VerifySignature(buyerPk); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid Signature for BuyOffer")
	}

	sellerPk, err := s.svcCtx.MemCache.GetAccountPkByIndex(txInfo.SellOffer.AccountIndex)
	if err != nil {
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("unknown AccountIndex of SellOffer")
			}
			return "", errorcode.AppErrInternal
		}
	}
	if err := txInfo.SellOffer.VerifySignature(sellerPk); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid Signature for SellOffer")
	}

	//check buy offer and sell offer
	if txInfo.BuyOffer.ExpiredAt < now || txInfo.SellOffer.ExpiredAt < now {
		logx.Errorf("invalid ExpiredAt of BuyOffer or SellOffer")
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid ExpiredAt of BuyOffer or SellOffer")
	}
	if txInfo.BuyOffer.NftIndex != txInfo.SellOffer.NftIndex ||
		txInfo.BuyOffer.AssetId != txInfo.SellOffer.AssetId ||
		txInfo.BuyOffer.AssetAmount.String() != txInfo.SellOffer.AssetAmount.String() ||
		txInfo.BuyOffer.TreasuryRate != txInfo.SellOffer.TreasuryRate {
		return "", errorcode.AppErrInvalidTxField.RefineError("mismatch between BuyOffer and SellOffer")
	}

	// check buyer and seller
	_, err = s.svcCtx.MemCache.GetAccountWithFallback(txInfo.BuyOffer.AccountIndex, func() (interface{}, error) {
		return s.svcCtx.AccountModel.GetAccountByAccountIndex(txInfo.BuyOffer.AccountIndex)
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid BuyOffer.AccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.BuyOffer.AccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}

	_, err = s.svcCtx.MemCache.GetAccountWithFallback(txInfo.SellOffer.AccountIndex, func() (interface{}, error) {
		return s.svcCtx.AccountModel.GetAccountByAccountIndex(txInfo.SellOffer.AccountIndex)
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid SellOffer.AccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.SellOffer.AccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}

	//check from account
	fromAccount, err := s.svcCtx.StateFetcher.GetLatestAccount(s.ctx, txInfo.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.AccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}

	nftInfo, err := s.svcCtx.StateFetcher.GetLatestNft(s.ctx, txInfo.BuyOffer.NftIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid BuyOffer.NftIndex")
		}
		logx.Errorf("fail to get nft info: %d, err: %s", txInfo.BuyOffer.NftIndex, err.Error())
		return "", err
	}
	if nftInfo.OwnerAccountIndex != txInfo.SellOffer.AccountIndex {
		logx.Errorf("not owner, owner: %d, seller: %d", nftInfo.OwnerAccountIndex, txInfo.SellOffer.AccountIndex)
		return "", errorcode.AppErrInvalidTxField.RefineError("seller is not nft owner")
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
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       txInfo.BuyOffer.AssetId,
		TxAmount:      txInfo.BuyOffer.AssetAmount.String(),
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
