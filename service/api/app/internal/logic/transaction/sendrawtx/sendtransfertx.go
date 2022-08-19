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

type transferTxSender struct {
	txType          int
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	gasChecker      GasChecker
	nonceChecker    NonceChecker
	mempoolTxSender MempoolTxSender
}

func NewTransferTxSender(ctx context.Context, svcCtx *svc.ServiceContext,
	gasChecker *gasChecker, nonceChecker *nonceChecker, sender *mempoolTxSender) *transferTxSender {
	return &transferTxSender{
		txType:          commonTx.TxTypeTransfer,
		ctx:             ctx,
		svcCtx:          svcCtx,
		gasChecker:      gasChecker,
		nonceChecker:    nonceChecker,
		mempoolTxSender: sender,
	}
}

func (s *transferTxSender) SendTx(rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseTransferTxInfo(rawTxInfo)
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

	// check to account
	toAccount, err := s.svcCtx.MemCache.GetAccountWithFallback(txInfo.ToAccountIndex, func() (interface{}, error) {
		return s.svcCtx.AccountModel.GetAccountByIndex(txInfo.ToAccountIndex)
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

	// check from account
	fromAccount, err := s.svcCtx.StateFetcher.GetLatestAccount(txInfo.FromAccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.FromAccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}

	asset, ok := fromAccount.AssetInfo[txInfo.AssetId]
	if !ok || asset.Balance.Cmp(txInfo.AssetAmount) < 0 {
		logx.Errorf("not enough asset in balance: %d, err: %s", fromAccount.AccountIndex, err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError("not enough asset in balance")
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
		AssetId:       txInfo.AssetId,
		TxAmount:      txInfo.AssetAmount.String(),
		Memo:          txInfo.Memo,
		AccountIndex:  txInfo.FromAccountIndex,
		Nonce:         txInfo.Nonce,
		ExpiredAt:     txInfo.ExpiredAt,
	}
	txId, err = s.mempoolTxSender.SendMempoolTx(func(txInfo interface{}) ([]byte, error) {
		return legendTxTypes.ComputeTransferMsgHash(txInfo.(*legendTxTypes.TransferTxInfo), mimc.NewMiMC())
	}, txInfo, mempoolTx)

	return txId, err
}
