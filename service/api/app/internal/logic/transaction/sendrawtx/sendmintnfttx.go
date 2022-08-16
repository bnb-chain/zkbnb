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

type mintNftTxSender struct {
	txType          int
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	gasChecker      GasChecker
	nonceChecker    NonceChecker
	mempoolTxSender MempoolTxSender
}

func NewMintNftTxSender(ctx context.Context, svcCtx *svc.ServiceContext,
	gasChecker *gasChecker, nonceChecker *nonceChecker, sender *mempoolTxSender) *mintNftTxSender {
	return &mintNftTxSender{
		txType:          commonTx.TxTypeMintNft,
		ctx:             ctx,
		svcCtx:          svcCtx,
		gasChecker:      gasChecker,
		nonceChecker:    nonceChecker,
		mempoolTxSender: sender,
	}
}
func (s *mintNftTxSender) SendTx(rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseMintNftTxInfo(rawTxInfo)
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
	accountPk, err := s.svcCtx.MemCache.GetAccountPkByIndex(txInfo.CreatorAccountIndex)
	if err != nil {
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("unknown CreatorAccountIndex")
			}
			return "", errorcode.AppErrInternal
		}
	}
	if err := txInfo.VerifySignature(accountPk); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid Signature")
	}

	// check to account
	toAccount, err := s.svcCtx.MemCache.GetAccountWithFallback(txInfo.ToAccountIndex, func() (interface{}, error) {
		return s.svcCtx.AccountModel.GetAccountByAccountIndex(txInfo.ToAccountIndex)
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid ToAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.ToAccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}
	if toAccount.AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("invalid account name hash, expected: %s, actual: %s", toAccount.AccountNameHash, txInfo.ToAccountNameHash)
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid ToAccountNameHash")
	}

	//check creator account
	creatorAccount, err := s.svcCtx.StateFetcher.GetLatestAccount(s.ctx, txInfo.CreatorAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.CreatorAccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}
	if creatorAccount.CollectionNonce < txInfo.NftCollectionId {
		logx.Errorf("collection id %d is greater than collection nonce %d",
			txInfo.NftCollectionId, creatorAccount.CollectionNonce)
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid NftCollectionId")
	}

	//check gas
	if err := s.gasChecker.CheckGas(creatorAccount, txInfo.GasAccountIndex, txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount); err != nil {
		return "", errorcode.AppErrInvalidTxField.RefineError(err.Error())
	}

	//check nonce
	if err := s.nonceChecker.CheckNonce(creatorAccount, txInfo.Nonce); err != nil {
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
		AccountIndex:  txInfo.CreatorAccountIndex,
		Nonce:         txInfo.Nonce,
		ExpiredAt:     txInfo.ExpiredAt,
	}
	txId, err = s.mempoolTxSender.SendMempoolTx(func(txInfo interface{}) ([]byte, error) {
		return legendTxTypes.ComputeMintNftMsgHash(txInfo.(*legendTxTypes.MintNftTxInfo), mimc.NewMiMC())
	}, txInfo, mempoolTx)

	return txId, err
}
