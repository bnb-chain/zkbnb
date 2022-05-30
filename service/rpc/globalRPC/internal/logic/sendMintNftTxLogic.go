/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logic

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/zeromicro/go-zero/core/stores/redis"
	"reflect"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *SendTxLogic) sendMintNftTx(rawTxInfo string) (txId string, err error) {
	// parse transfer tx info
	txInfo, err := commonTx.ParseMintNftTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendMintNftTx.ParseMintNftTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	/*
		Check Params
	*/
	if txInfo.NftCollectionId == commonConstant.NilCollectionId {
		errInfo := fmt.Sprintf("[sendMintNftTx] err: invalid collection id %v", txInfo.NftCollectionId)
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New(errInfo))
	}
	accountInfo, err := globalmapHandler.GetLatestAccountInfo(
		l.svcCtx.AccountModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.RedisConnection,
		txInfo.CreatorAccountIndex,
	)
	if err != nil {
		errInfo := fmt.Sprintf("[sendMintNftTx] err: invalid accountIndex %v", txInfo.CreatorAccountIndex)
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New(errInfo))
	}
	if accountInfo.CollectionNonce < txInfo.NftCollectionId {
		errInfo := fmt.Sprintf("[sendMintNftTx] err: invalid collection id %v", txInfo.NftCollectionId)
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New(errInfo))
	}

	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.CreatorAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendMintNftTx] err: invalid accountIndex %v", txInfo.CreatorAccountIndex)
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.ToAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendMintNftTx] err: invalid accountIndex %v", txInfo.ToAccountIndex)
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New(errInfo))
	}
	// check gas account index
	gasAccountIndexConfig, err := l.svcCtx.SysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to get sysconfig by name: %s", err.Error())
		return "", l.HandleCreateFailMintNftTx(txInfo, err)
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New("[sendMintNftTx] unable to parse big int"))
	}
	if gasAccountIndex != txInfo.GasAccountIndex {
		logx.Errorf("[sendMintNftTx] invalid gas account index")
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New("[sendMintNftTx] invalid gas account index"))
	}

	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendMintNftTx] invalid time stamp")
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New("[sendMintNftTx] invalid time stamp"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
		nftIndex       int64
		redisLock      *redis.RedisLock
	)
	redisLock, nftIndex, err = globalmapHandler.GetLatestNftIndexForWrite(l.svcCtx.NftModel, l.svcCtx.RedisConnection)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to get latest nft index: %s", err.Error())
		return "", err
	}
	defer redisLock.Release()
	accountInfoMap[txInfo.CreatorAccountIndex], err = globalmapHandler.GetLatestAccountInfo(
		l.svcCtx.AccountModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.RedisConnection,
		txInfo.CreatorAccountIndex,
	)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to get account info: %s", err.Error())
		return "", l.HandleCreateFailMintNftTx(txInfo, err)
	}
	// get account info by to index
	if accountInfoMap[txInfo.ToAccountIndex] == nil {
		accountInfoMap[txInfo.ToAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.ToAccountIndex)
		if err != nil {
			logx.Errorf("[sendMintNftTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateFailMintNftTx(txInfo, err)
		}
	}
	if accountInfoMap[txInfo.ToAccountIndex].AccountNameHash != txInfo.ToAccountNameHash {
		logx.Errorf("[sendMintNftTx] invalid account name")
		return "", l.HandleCreateFailMintNftTx(txInfo, errors.New("[sendMintNftTx] invalid account name"))
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = globalmapHandler.GetBasicAccountInfo(
			l.svcCtx.AccountModel,
			l.svcCtx.RedisConnection,
			txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendMintNftTx] unable to get account info: %s", err.Error())
			return "", l.HandleCreateFailMintNftTx(txInfo, err)
		}
	}
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// set tx info
	txInfo.NftIndex = nftIndex
	// verify transfer tx
	txDetails, err = txVerification.VerifyMintNftTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return "", l.HandleCreateFailMintNftTx(txInfo, err)
	}

	/*
		Check tx details
	*/

	/*
		Create Mempool Transaction
	*/
	// construct nft info
	nftInfo := &nft.L2Nft{
		NftIndex:            nftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		NftL1Address:        commonConstant.NilL1Address,
		NftL1TokenId:        commonConstant.NilL1TokenId,
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.NftCollectionId,
	}
	// delete key
	key := util.GetNftKeyForRead(nftIndex)
	_, err = l.svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to delete key from redis: %s", err.Error())
		return "", l.HandleCreateFailMintNftTx(txInfo, err)
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return "", l.HandleCreateFailMintNftTx(txInfo, err)
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeMintNft,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		nftIndex,
		commonConstant.NilPairIndex,
		commonConstant.NilAssetId,
		commonConstant.NilAssetAmountStr,
		"",
		string(txInfoBytes),
		"",
		txInfo.CreatorAccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	err = CreateMempoolTxForMintNft(nftInfo, mempoolTx, l.svcCtx.RedisConnection, l.svcCtx.MempoolModel)
	if err != nil {
		return "", l.HandleCreateFailMintNftTx(txInfo, err)
	}
	// update redis
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
	_ = l.svcCtx.RedisConnection.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)
	return txId, nil
}

func (l *SendTxLogic) HandleCreateFailMintNftTx(txInfo *commonTx.MintNftTxInfo, err error) error {
	errCreate := l.CreateFailMintNftTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Error("[sendtransfertxlogic.HandleCreateFailMintNftTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailMintNftTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendTxLogic) CreateFailMintNftTx(info *commonTx.MintNftTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailMintNftTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeMintNft,
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

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailMintNftTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}

func CreateMempoolTxForMintNft(
	nftInfo *nft.L2Nft,
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
	err = mempoolModel.CreateMempoolTxAndL2Nft(nMempoolTx, nftInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[CreateMempoolTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
