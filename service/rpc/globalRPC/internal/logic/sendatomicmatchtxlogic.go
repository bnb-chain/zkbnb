/*
* Copyright © 2021 Zkbas Protocol
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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type SendAtomicMatchTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendAtomicMatchTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAtomicMatchTxLogic {
	return &SendAtomicMatchTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *SendAtomicMatchTxLogic) SendAtomicMatchTx(reqSendTx *globalRPCProto.ReqSendTxByRawInfo) (respSendTx *globalRPCProto.RespSendTx, err error) {
	respSendTx = &globalRPCProto.RespSendTx{}
	rawTxInfo := reqSendTx.TxInfo
	// parse transfer tx info
	txInfo, err := commonTx.ParseAtomicMatchTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx.ParseAtomicMatchTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return respSendTx, errors.New(errInfo)
	}

	/*
		Check Params
	*/
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param: %v,err: %s", txInfo.GasFeeAssetAmount, err.Error())
		return respSendTx, err
	}
	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.AccountIndex)
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.BuyOffer.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.BuyOffer.AccountIndex)
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, errors.New(errInfo))
	}
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.SellOffer.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendAtomicMatchTx] err: invalid accountIndex %v", txInfo.SellOffer.AccountIndex)
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, errors.New(errInfo))
	}
	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %s", err.Error())
		return nil, err
	}
	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now || txInfo.BuyOffer.ExpiredAt < now || txInfo.SellOffer.ExpiredAt < now {
		logx.Errorf("[sendAtomicMatchTx] invalid time stamp")
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, errors.New("[sendAtomicMatchTx] invalid time stamp"))
	}

	if txInfo.BuyOffer.NftIndex != txInfo.SellOffer.NftIndex ||
		txInfo.BuyOffer.AssetId != txInfo.SellOffer.AssetId ||
		txInfo.BuyOffer.AssetAmount.String() != txInfo.SellOffer.AssetAmount.String() ||
		txInfo.BuyOffer.TreasuryRate != txInfo.SellOffer.TreasuryRate {
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, errors.New("[sendAtomicMatchTx] invalid params"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	nftInfo, err := l.commglobalmap.GetLatestNftInfoForRead(l.ctx, txInfo.BuyOffer.NftIndex)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to get latest nft index: %s", err.Error())
		return respSendTx, err
	}
	accountInfoMap[txInfo.AccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
	}
	// get account info by to index
	if accountInfoMap[txInfo.BuyOffer.AccountIndex] == nil {
		accountInfoMap[txInfo.BuyOffer.AccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.BuyOffer.AccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
		}
	}
	if accountInfoMap[nftInfo.CreatorAccountIndex] == nil {
		accountInfoMap[nftInfo.CreatorAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, nftInfo.CreatorAccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
		}
	}
	if accountInfoMap[txInfo.SellOffer.AccountIndex] == nil {
		accountInfoMap[txInfo.SellOffer.AccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.SellOffer.AccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
		}
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendAtomicMatchTx] unable to get account info: %s", err.Error())
			return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
		}
	}
	if nftInfo.OwnerAccountIndex != txInfo.SellOffer.AccountIndex {
		logx.Errorf("[sendAtomicMatchTx] you're not owner")
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, errors.New("[sendAtomicMatchTx] you're not owner"))
	}
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify transfer tx
	txDetails, err = txVerification.VerifyAtomicMatchTxInfo(
		accountInfoMap,
		nftInfo,
		txInfo,
	)
	if err != nil {
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
	}

	/*
		Check tx details
	*/

	/*
		Create Mempool Transaction
	*/
	// delete key
	key := util.GetNftKeyForRead(txInfo.BuyOffer.NftIndex)
	_, err = l.svcCtx.RedisConnection.Del(key)
	if err != nil {
		logx.Errorf("[sendAtomicMatchTx] unable to delete key from redis: %s", err.Error())
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
	}
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
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
	err = createMempoolTxForAtomicMatch(
		nftExchange,
		mempoolTx,
		offers,
		l.svcCtx.RedisConnection,
		l.svcCtx.MempoolModel,
	)
	if err != nil {
		return respSendTx, l.HandleCreateFailAtomicMatchTx(txInfo, err)
	}
	respSendTx.TxId = txId
	// update redis
	var formatNftInfo *commonAsset.NftInfo
	for _, txDetail := range mempoolTx.MempoolDetails {
		if txDetail.AssetType == commonAsset.NftAssetType {
			formatNftInfo, err = commonAsset.ParseNftInfo(txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[sendMintNftTx] unable to parse nft info: %s", err.Error())
				return respSendTx, nil
			}
		}
	}
	nftInfoBytes, err := json.Marshal(formatNftInfo)
	if err != nil {
		logx.Errorf("[sendMintNftTx] unable to marshal: %s", err.Error())
		return respSendTx, nil
	}
	_ = l.svcCtx.RedisConnection.Setex(key, string(nftInfoBytes), globalmapHandler.NftExpiryTime)
	return respSendTx, nil
}

func (l *SendAtomicMatchTxLogic) HandleCreateFailAtomicMatchTx(txInfo *commonTx.AtomicMatchTxInfo, err error) error {
	errCreate := l.CreateFailAtomicMatchTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Errorf("[sendtransfertxlogic.HandleCreateFailAtomicMatchTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendAtomicMatchTxLogic) CreateFailAtomicMatchTx(info *commonTx.AtomicMatchTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
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

	err = l.svcCtx.FailTxModel.CreateFailTx(failTx)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailAtomicMatchTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}

func createMempoolTxForAtomicMatch(
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
