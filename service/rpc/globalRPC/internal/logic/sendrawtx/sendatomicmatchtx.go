package sendrawtx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/sysconfigName"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zeromicro/go-zero/core/logx"
)

func SendAtomicMatchTx(ctx context.Context, svcCtx *svc.ServiceContext, commglobalmap commglobalmap.Commglobalmap, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx.ParseAtomicMatchTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return "", err
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.AccountIndex)
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.BuyOffer.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.BuyOffer.AccountIndex)
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.SellOffer.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.SellOffer.AccountIndex)
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, errors.New(errInfo))
	}
	commglobalmap.DeleteLatestAccountInfoInCache(ctx, txInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[DeleteLatestAccountInfoInCache] err:%v", err)
	}
	gasAccountIndexConfig, err := svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to get sysconfig by name: %s", err.Error())
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, errors.New("[sendAtomicMatchTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendAtomicMatchTx] invalid gas account index")
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, errors.New("[sendAtomicMatchTx] invalid gas account index"))
	}
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now || txInfo.BuyOffer.ExpiredAt < now || txInfo.SellOffer.ExpiredAt < now {
		logx.Errorf("[sendAtomicMatchTx] invalid time stamp")
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, errors.New("[sendAtomicMatchTx] invalid time stamp"))
	}
	if txInfo.BuyOffer.NftIndex != txInfo.SellOffer.NftIndex ||
		txInfo.BuyOffer.AssetId != txInfo.SellOffer.AssetId ||
		txInfo.BuyOffer.AssetAmount.String() != txInfo.SellOffer.AssetAmount.String() ||
		txInfo.BuyOffer.TreasuryRate != txInfo.SellOffer.TreasuryRate {
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, errors.New("[sendAtomicMatchTx] invalid params"))
	}
	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	nftInfo, err := globalmapHandler.GetLatestNftInfoForRead(
		svcCtx.NftModel,
		svcCtx.MempoolModel,
		svcCtx.RedisConnection,
		txInfo.BuyOffer.NftIndex,
	)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to get latest nft index: %s", err.Error())
		return "", err
	}
	accountInfoMap[txInfo.AccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
	}
	if accountInfoMap[txInfo.BuyOffer.AccountIndex] == nil {
		accountInfoMap[txInfo.BuyOffer.AccountIndex], err = commglobalmap.GetLatestAccountInfo(ctx, txInfo.BuyOffer.AccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
		}
	}
	if accountInfoMap[nftInfo.CreatorAccountIndex] == nil {
		accountInfoMap[nftInfo.CreatorAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			nftInfo.CreatorAccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
		}
	}
	if accountInfoMap[txInfo.SellOffer.AccountIndex] == nil {
		accountInfoMap[txInfo.SellOffer.AccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.SellOffer.AccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
		}
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			svcCtx.AccountModel,
			svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
		}
	}
	if nftInfo.OwnerAccountIndex != txInfo.SellOffer.AccountIndex {
		logx.Errorf("[sendAtomicMatchTx] you're not owner")
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, errors.New("[sendAtomicMatchTx] you're not owner"))
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
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
	}
	key := util.GetNftKeyForRead(txInfo.BuyOffer.NftIndex)
	_, err = svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to delete key from redis: %s", err.Error())
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
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
	err = CreateMempoolTxForAtomicMatch(
		nftExchange,
		mempoolTx,
		offers,
		svcCtx.RedisConnection,
		svcCtx.MempoolModel,
	)
	if err != nil {
		return "", handleCreateFailAtomicMatchTx(svcCtx.FailTxModel, txInfo, err)
	}
	var formatNftInfo *commonAsset.NftInfo
	for _, txDetail := range mempoolTx.MempoolDetails {
		if txDetail.AssetType == commonAsset.NftAssetType {
			formatNftInfo, err = commonAsset.ParseNftInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendMintNftTx] unable to parse nft info: %s", err.Error())
				return txId, nil
			}
		}
	}
	nftInfoBytes, err := json.Marshal(formatNftInfo)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to marshal: %s", err.Error())
		return txId, nil
	}
	_ = svcCtx.RedisConnection.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)
	return txId, nil
}

func handleCreateFailAtomicMatchTx(failTxModel tx.FailTxModel, txInfo *commonTx.AtomicMatchTxInfo, err error) error {
	errCreate := createFailAtomicMatchTx(failTxModel, txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendtransfertxlogic.HandleCreateFailAtomicMatchTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func createFailAtomicMatchTx(failTxModel tx.FailTxModel, info *commonTx.AtomicMatchTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeAtomicMatch,
		// tx fee
		GasFee: info.GasFeeAssetAmount.String(),
		// tx fee l1asset id
		GasFeeAssetId: info.GasFeeAssetId,
		// tx status, 1 - success(default), 2 - failure
		TxStatus: tx.StatusFail,
		// l1asset id
		AssetAId: commonConstant.NilAssetId,
		// AssetBId
		AssetBId: commonConstant.NilAssetId,
		// tx amount
		TxAmount: commonConstant.NilAssetAmountStr,
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txInfo),
		// extra info, if tx fails, show the error info
		ExtraInfo: extraInfo,
		// native memo info
		Memo: "",
	}
	err = failTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}

func CreateMempoolTxForAtomicMatch(
	nftExchange *nft.L2NftExchange,
	nMempoolTx *mempool.MempoolTx,
	offers []*nft.Offer,
	redisConnection *redis.Redis,
	mempoolModel mempool.MempoolModel,
) (err error) {
	var keys []string
	for _, mempoolTxDetail := range nMempoolTx.MempoolDetails {
		keys = append(keys, util.GetAccountKey(mempoolTxDetail.AccountIndex))
	}
	_, err = redisConnection.Del(keys...)
	if err != nil {
		logx.Errorf("[CreateMempoolTx] error with redis: %s", err.Error())
		return err
	}
	// write into mempool
	err = mempoolModel.CreateMempoolTxAndL2NftExchange(
		nMempoolTx,
		offers,
		nftExchange,
	)
	if err != nil {
		errInfo := fmt.Sprintf("[CreateMempoolTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}

func ConstructMempoolTx(
	txType int64,
	gasFeeAssetId int64,
	gasFeeAssetAmount string,
	nftIndex int64,
	pairIndex int64,
	assetId int64,
	txAmount string,
	toAddress string,
	txInfo string,
	memo string,
	accountIndex int64,
	nonce int64,
	expiredAt int64,
	txDetails []*mempool.MempoolTxDetail,
) (txId string, mempoolTx *mempool.MempoolTx) {
	txId = util.RandomUUID()
	return txId, &mempool.MempoolTx{
		TxHash:         txId,
		TxType:         txType,
		GasFeeAssetId:  gasFeeAssetId,
		GasFee:         gasFeeAssetAmount,
		NftIndex:       nftIndex,
		PairIndex:      pairIndex,
		AssetId:        assetId,
		TxAmount:       txAmount,
		NativeAddress:  toAddress,
		MempoolDetails: txDetails,
		TxInfo:         txInfo,
		ExtraInfo:      "",
		Memo:           memo,
		AccountIndex:   accountIndex,
		Nonce:          nonce,
		ExpiredAt:      expiredAt,
		L2BlockHeight:  commonConstant.NilBlockHeight,
		Status:         mempool.PendingTxStatus,
	}
}

func CreateMempoolTx(
	nMempoolTx *mempool.MempoolTx,
	redisConnection *redis.Redis,
	mempoolModel mempool.MempoolModel,
) (err error) {
	var keys []string
	for _, mempoolTxDetail := range nMempoolTx.MempoolDetails {
		keys = append(keys, util.GetAccountKey(mempoolTxDetail.AccountIndex))
	}
	_, err = redisConnection.Del(keys...)
	if err != nil {
		logx.Errorf("[CreateMempoolTx] error with redis: %s", err.Error())
		return err
	}
	// write into mempool
	err = mempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{nMempoolTx})
	if err != nil {
		errInfo := fmt.Sprintf("[CreateMempoolTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
