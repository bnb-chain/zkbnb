package sendrawtx

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

func SendCancelOfferTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseCancelOfferTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.AppErrInvalidTx
	}

	if err := legendTxTypes.ValidateCancelOfferTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return "", err
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid AccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.AccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, txInfo.GasAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("invalid GasAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.GasAccountIndex, err.Error())
			return "", errorcode.AppErrInternal
		}
	}

	offerAssetId := txInfo.OfferId / 128
	offerIndex := txInfo.OfferId % 128
	if accountInfoMap[txInfo.AccountIndex].AssetInfo[offerAssetId] == nil {
		accountInfoMap[txInfo.AccountIndex].AssetInfo[offerAssetId] = &commonAsset.AccountAsset{
			AssetId:                  offerAssetId,
			Balance:                  big.NewInt(0),
			LpAmount:                 big.NewInt(0),
			OfferCanceledOrFinalized: big.NewInt(0),
		}
	} else {
		offerInfo := accountInfoMap[txInfo.AccountIndex].AssetInfo[offerAssetId].OfferCanceledOrFinalized
		xBit := offerInfo.Bit(int(offerIndex))
		if xBit == 1 {
			logx.Errorf("offer is already confirmed or canceled")
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid OfferId, already confirmed or canceled")
		}
	}

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify tx
	txDetails, err = txVerification.VerifyCancelOfferTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return "", errorcode.AppErrVerification.RefineError(err)
	}

	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.AppErrInternal
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeCancelOffer,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		commonConstant.NilPairIndex,
		commonConstant.NilAssetId,
		accountInfoMap[txInfo.AccountIndex].AccountName,
		commonConstant.NilL1Address,
		string(txInfoBytes),
		"",
		txInfo.AccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	var isUpdate bool
	offerInfo, err := svcCtx.OfferModel.GetOfferByAccountIndexAndOfferId(txInfo.AccountIndex, txInfo.OfferId)
	if err == errorcode.DbErrNotFound {
		offerInfo = &nft.Offer{
			OfferType:    0,
			OfferId:      txInfo.OfferId,
			AccountIndex: txInfo.AccountIndex,
			NftIndex:     0,
			AssetId:      0,
			AssetAmount:  "0",
			ListedAt:     0,
			ExpiredAt:    0,
			TreasuryRate: 0,
			Sig:          "",
			Status:       nft.OfferFinishedStatus,
		}
	} else {
		offerInfo.Status = nft.OfferFinishedStatus
		isUpdate = true
	}

	if err := svcCtx.MempoolModel.CreateMempoolTxAndUpdateOffer(mempoolTx, offerInfo, isUpdate); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeCancelOffer, txInfo, err)
		return "", errorcode.AppErrInternal
	}
	return txId, nil
}
