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
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/common/zcrypto/txVerification"
	"github.com/zeromicro/go-zero/core/logx"
)

func (l *SendTxLogic) sendOfferTx(rawTxInfo string) (txId string, err error) {
	// parse transfer tx info
	txInfo, err := commonTx.ParseOfferTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendOfferTx.ParseOfferTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return "", errors.New(errInfo)
	}

	/*
		Check Params
	*/
	if err := util.CheckPackedAmount(txInfo.AssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAmount, err)
		return "", err
	}

	if err := util.CheckPackedAmount(txInfo.AssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.AssetAmount, err)
		return "", err
	}
	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.AccountIndex))
	if err != nil {
		return "", err
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAssetId, reflect.ValueOf(txInfo.AssetId))
	if err != nil {
		return "", err
	}

	redisLock, offerId, err := globalmapHandler.GetLatestOfferIdForWrite(l.svcCtx.OfferModel, l.svcCtx.RedisConnection, txInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[sendOfferTx] unable to get latest offer id: %s", err.Error())
		return "", err
	}
	defer redisLock.Release()
	if offerId != txInfo.OfferId {
		logx.Errorf("[sendOfferTx] invalid offer id")
		return "", errors.New("[sendOfferTx] invalid offer id")
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	nftInfo, err := globalmapHandler.GetLatestNftInfoForRead(
		l.svcCtx.NftModel,
		l.svcCtx.MempoolModel,
		l.svcCtx.RedisConnection,
		txInfo.NftIndex,
	)
	if err != nil {
		logx.Errorf("[sendOfferTx] unable to get latest nft index: %s", err.Error())
		return "", err
	}
	accountInfoMap[txInfo.AccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[sendOfferTx] unable to get account info: %s", err.Error())
		return "", err
	}
	// verify transfer tx
	err = txVerification.VerifyOfferTxInfo(
		accountInfoMap,
		nftInfo,
		txInfo,
	)
	if err != nil {
		return "", err
	}
	// write into offer table
	offer := &nft.Offer{
		OfferType:    txInfo.Type,
		OfferId:      txInfo.OfferId,
		AccountIndex: txInfo.AccountIndex,
		NftIndex:     txInfo.NftIndex,
		AssetId:      txInfo.AssetId,
		AssetAmount:  txInfo.AssetAmount.String(),
		ListedAt:     txInfo.ListedAt,
		ExpiredAt:    txInfo.ExpiredAt,
		TreasuryRate: txInfo.TreasuryRate,
		Sig:          common.Bytes2Hex(txInfo.Sig),
	}
	err = l.svcCtx.OfferModel.CreateOffer(offer)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(offer.OfferId, 10), nil
}
