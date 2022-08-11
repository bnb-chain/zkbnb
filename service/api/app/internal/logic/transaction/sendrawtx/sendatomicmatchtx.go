package sendrawtx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

func SendAtomicMatchTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.AppErrInvalidTx
	}

	if err := legendTxTypes.ValidateAtomicMatchTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return "", err
	}

	now := time.Now().UnixMilli()
	if txInfo.BuyOffer.ExpiredAt < now || txInfo.SellOffer.ExpiredAt < now {
		logx.Errorf("[sendAtomicMatchTx] invalid time stamp")
		return "", errorcode.AppErrInvalidTxField.RefineError("invalid ExpiredAt of BuyOffer or SellOffer")
	}
	if txInfo.BuyOffer.NftIndex != txInfo.SellOffer.NftIndex ||
		txInfo.BuyOffer.AssetId != txInfo.SellOffer.AssetId ||
		txInfo.BuyOffer.AssetAmount.String() != txInfo.SellOffer.AssetAmount.String() ||
		txInfo.BuyOffer.TreasuryRate != txInfo.SellOffer.TreasuryRate {
		return "", errorcode.AppErrInvalidTxField.RefineError("mismatch between BuyOffer and SellOffer")
	}
	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.AccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}
	if accountInfoMap[txInfo.BuyOffer.AccountIndex] == nil {
		accountInfoMap[txInfo.BuyOffer.AccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, txInfo.BuyOffer.AccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("invalid BuyOffer.AccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.BuyOffer.AccountIndex, err.Error())
			return "", errorcode.AppErrInternal
		}
	}
	if accountInfoMap[txInfo.SellOffer.AccountIndex] == nil {
		accountInfoMap[txInfo.SellOffer.AccountIndex], err = commglobalmap.GetBasicAccountInfo(ctx, txInfo.SellOffer.AccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("invalid SellOffer.AccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.SellOffer.AccountIndex, err.Error())
			return "", errorcode.AppErrInternal
		}
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

	nftInfo, err := commglobalmap.GetLatestNftInfoForRead(ctx, txInfo.BuyOffer.NftIndex)
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

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	txDetails, err = txVerification.VerifyAtomicMatchTxInfo(
		accountInfoMap,
		nftInfo,
		txInfo,
	)
	if err != nil {
		return "", errorcode.AppErrVerification.RefineError(err)
	}
	key := util.GetNftKeyForRead(txInfo.BuyOffer.NftIndex)
	_, err = svcCtx.RedisConn.Del(key)
	if err != nil {
		logx.Errorf("unable to delete key from redis: %s", err.Error())
		return "", errorcode.AppErrInternal
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.AppErrInternal
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeAtomicMatch,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		txInfo.BuyOffer.NftIndex,
		commonConstant.NilPairIndex,
		txInfo.BuyOffer.AssetId,
		txInfo.BuyOffer.AssetAmount.String(),
		"",
		string(txInfoBytes),
		"",
		txInfo.AccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	nftExchange := &nft.L2NftExchange{
		BuyerAccountIndex: txInfo.BuyOffer.AccountIndex,
		OwnerAccountIndex: txInfo.SellOffer.AccountIndex,
		NftIndex:          txInfo.BuyOffer.NftIndex,
		AssetId:           txInfo.BuyOffer.AssetId,
		AssetAmount:       txInfo.BuyOffer.AssetAmount.String(),
	}
	var offers []*nft.Offer
	offers = append(offers, &nft.Offer{
		OfferType:    txInfo.BuyOffer.Type,
		OfferId:      txInfo.BuyOffer.OfferId,
		AccountIndex: txInfo.BuyOffer.AccountIndex,
		NftIndex:     txInfo.BuyOffer.NftIndex,
		AssetId:      txInfo.BuyOffer.AssetId,
		AssetAmount:  txInfo.BuyOffer.AssetAmount.String(),
		ListedAt:     txInfo.BuyOffer.ListedAt,
		ExpiredAt:    txInfo.BuyOffer.ExpiredAt,
		TreasuryRate: txInfo.BuyOffer.TreasuryRate,
		Sig:          common.Bytes2Hex(txInfo.BuyOffer.Sig),
		Status:       nft.OfferFinishedStatus,
	})
	offers = append(offers, &nft.Offer{
		OfferType:    txInfo.SellOffer.Type,
		OfferId:      txInfo.SellOffer.OfferId,
		AccountIndex: txInfo.SellOffer.AccountIndex,
		NftIndex:     txInfo.SellOffer.NftIndex,
		AssetId:      txInfo.SellOffer.AssetId,
		AssetAmount:  txInfo.SellOffer.AssetAmount.String(),
		ListedAt:     txInfo.SellOffer.ListedAt,
		ExpiredAt:    txInfo.SellOffer.ExpiredAt,
		TreasuryRate: txInfo.SellOffer.TreasuryRate,
		Sig:          common.Bytes2Hex(txInfo.SellOffer.Sig),
		Status:       nft.OfferFinishedStatus,
	})

	if err := svcCtx.MempoolModel.CreateMempoolTxAndL2NftExchange(mempoolTx, offers, nftExchange); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeAtomicMatch, txInfo, err)
		return "", err
	}
	var formatNftInfo *commonAsset.NftInfo
	for _, txDetail := range mempoolTx.MempoolDetails {
		if txDetail.AssetType == commonAsset.NftAssetType {
			formatNftInfo, err = commonAsset.ParseNftInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("unable to parse nft info: %s", err.Error())
				return txId, nil
			}
		}
	}
	nftInfoBytes, err := json.Marshal(formatNftInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx: %s", err.Error())
		return txId, nil
	}
	_ = svcCtx.RedisConn.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)
	return txId, nil
}
