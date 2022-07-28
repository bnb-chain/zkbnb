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
	"math/big"
	"reflect"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
)

type SendCancelOfferTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendCancelOfferTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCancelOfferTxLogic {
	return &SendCancelOfferTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}
func (l *SendCancelOfferTxLogic) SendCancelOfferTx(in *globalRPCProto.ReqSendTxByRawInfo) (respSendTx *globalRPCProto.RespSendTx, err error) {
	rawTxInfo := in.TxInfo
	respSendTx = &globalRPCProto.RespSendTx{}
	// parse transfer tx info
	txInfo, err := commonTx.ParseCancelOfferTxInfo(rawTxInfo)
	if err != nil {
		errInfo := fmt.Sprintf("[sendCancelOfferTx.ParseCancelOfferTxInfo] %s", err.Error())
		logx.Error(errInfo)
		return respSendTx, errors.New(errInfo)
	}
	/*
		Check Params
	*/
	if err := util.CheckPackedFee(txInfo.GasFeeAssetAmount); err != nil {
		logx.Errorf("[CheckPackedFee] param:%v,err:%v", txInfo.GasFeeAssetAmount, err)
		return respSendTx, err
	}
	// check param: from account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendCancelOfferTx] err: invalid accountIndex %v", txInfo.AccountIndex)
		return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, errors.New(errInfo))
	}
	// check param: to account index
	err = util.CheckRequestParam(util.TypeAccountIndex, reflect.ValueOf(txInfo.GasAccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[sendCancelOfferTx] err: invalid accountIndex %v", txInfo.GasAccountIndex)
		return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, errors.New(errInfo))
	}
	// check gas account index
	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, l.svcCtx.SysConfigModel); err != nil {
		logx.Errorf("[checkGasAccountIndex] err: %v", err)
		return nil, err
	}
	// check expired at
	now := time.Now().UnixMilli()
	if txInfo.ExpiredAt < now {
		logx.Errorf("[sendCancelOfferTx] invalid time stamp")
		return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, errors.New("[sendCancelOfferTx] invalid time stamp"))
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = l.commglobalmap.GetLatestAccountInfo(l.ctx, txInfo.AccountIndex)
	if err != nil {
		logx.Errorf("[sendCancelOfferTx] unable to get account info: %s", err.Error())
		return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, err)
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
			logx.Errorf("[sendCancelOfferTx] the offer is already confirmed or canceled")
			return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, errors.New("[sendCancelOfferTx] the offer is already confirmed or canceled"))
		}
	}
	// get account info by gas index
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		// get account info by gas index
		accountInfoMap[txInfo.GasAccountIndex], err = l.commglobalmap.GetBasicAccountInfo(l.ctx, txInfo.GasAccountIndex)
		if err != nil {
			logx.Errorf("[sendCancelOfferTx] unable to get account info: %s", err.Error())
			return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, err)
		}
	}
	var (
		txDetails []*mempool.MempoolTxDetail
	)
	// verify transfer tx
	txDetails, err = txVerification.VerifyCancelOfferTxInfo(
		accountInfoMap,
		txInfo,
	)
	if err != nil {
		return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, err)
	}

	/*
		Check tx details
	*/

	/*
		Create Mempool Transaction
	*/
	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, err)
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
	respSendTx.TxId = txId
	var isUpdate bool
	offerInfo, err := l.svcCtx.OfferModel.GetOfferByAccountIndexAndOfferId(txInfo.AccountIndex, txInfo.OfferId)
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
	err = CreateMempoolTxForCancelOffer(
		mempoolTx,
		offerInfo,
		isUpdate,
		l.svcCtx.RedisConnection,
		l.svcCtx.MempoolModel,
	)
	if err != nil {
		return respSendTx, l.HandleCreateFailCancelOfferTx(txInfo, err)
	}
	return respSendTx, nil
}

func (l *SendCancelOfferTxLogic) HandleCreateFailCancelOfferTx(txInfo *commonTx.CancelOfferTxInfo, err error) error {
	errCreate := l.CreateFailCancelOfferTx(txInfo, err.Error())
	if errCreate != nil {
		logx.Errorf("[sendtransfertxlogic.HandleCreateFailCancelOfferTx] %s", errCreate.Error())
		return errCreate
	} else {
		errInfo := fmt.Sprintf("[sendtransfertxlogic.HandleCreateFailCancelOfferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
}

func (l *SendCancelOfferTxLogic) CreateFailCancelOfferTx(info *commonTx.CancelOfferTxInfo, extraInfo string) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txInfo, err := json.Marshal(info)
	if err != nil {
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailCancelOfferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: commonTx.TxTypeCancelOffer,
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
		errInfo := fmt.Sprintf("[sendtxlogic.CreateFailCancelOfferTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}

func CreateMempoolTxForCancelOffer(
	nMempoolTx *mempool.MempoolTx,
	offer *nft.Offer,
	isUpdate bool,
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
	err = mempoolModel.CreateMempoolTxAndUpdateOffer(
		nMempoolTx,
		offer,
		isUpdate,
	)
	if err != nil {
		errInfo := fmt.Sprintf("[CreateMempoolTx] %s", err.Error())
		logx.Error(errInfo)
		return errors.New(errInfo)
	}
	return nil
}
