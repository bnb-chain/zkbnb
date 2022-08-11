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

package monitor

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
)

func (m *Monitor) MonitorPriorityRequests() error {
	pendingRequests, err := m.PriorityRequestModel.GetPriorityRequestsByStatus(PendingStatus)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			logx.Info("no pending requests")
			return err
		} else {
			logx.Error("unable to get pending requests")
			return err
		}
	}
	var (
		pendingNewMempoolTxs []*mempool.MempoolTx
	)
	// get last handled request id
	currentRequestId, err := m.PriorityRequestModel.GetLatestHandledRequestId()
	if err != nil {
		logx.Errorf("unable to get last handled request id: %s", err.Error())
		return err
	}

	for _, request := range pendingRequests {
		// request id must be in order
		if request.RequestId != currentRequestId+1 {
			logx.Errorf("invalid request id")
			return errors.New("invalid request id")
		}
		currentRequestId++

		txHash := ComputeL1TxTxHash(request.RequestId, request.L1TxHash)

		mempoolTx := &mempool.MempoolTx{
			TxHash:        txHash,
			GasFeeAssetId: commonConstant.NilAssetId,
			GasFee:        commonConstant.NilAssetAmountStr,
			NftIndex:      commonConstant.NilTxNftIndex,
			PairIndex:     commonConstant.NilPairIndex,
			AssetId:       commonConstant.NilAssetId,
			TxAmount:      commonConstant.NilAssetAmountStr,
			NativeAddress: request.SenderAddress,
			AccountIndex:  commonConstant.NilAccountIndex,
			Nonce:         commonConstant.NilNonce,
			ExpiredAt:     commonConstant.NilExpiredAt,
			L2BlockHeight: commonConstant.NilBlockHeight,
			Status:        mempool.PendingTxStatus,
		}
		// handle request based on request type
		switch request.TxType {
		case TxTypeRegisterZns:
			// parse request info
			txInfo, err := util.ParseRegisterZnsPubData(common.FromHex(request.Pubdata))
			if err != nil {
				logx.Errorf("unable to parse registerZNS pub data: %s", err.Error())
				return err
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("unable to serialize request info : %s", err.Error())
				return err
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.AccountIndex = txInfo.AccountIndex
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeCreatePair:
			txInfo, err := util.ParseCreatePairPubData(common.FromHex(request.Pubdata))
			if err != nil {
				logx.Errorf("unable to parse registerZNS pub data: %s", err.Error())
				return err
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("unable to serialize request info : %s", err.Error())
				return err
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.PairIndex = txInfo.PairIndex
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeUpdatePairRate:
			txInfo, err := util.ParseUpdatePairRatePubData(common.FromHex(request.Pubdata))
			if err != nil {
				logx.Errorf("unable to parse update pair rate pub data: %s", err.Error())
				return err
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("unable to serialize request info : %s", err.Error())
				return err
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeDeposit:
			txInfo, err := util.ParseDepositPubData(common.FromHex(request.Pubdata))
			if err != nil {
				logx.Errorf("unable to parse deposit pub data: %s", err.Error())
				return err
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("unable to serialize request info : %s", err.Error())
				return err
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.AccountIndex = txInfo.AccountIndex
			mempoolTx.AssetId = txInfo.AssetId
			mempoolTx.TxAmount = txInfo.AssetAmount.String()
			mempoolTx.AccountIndex = txInfo.AccountIndex
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeDepositNft:
			txInfo, err := util.ParseDepositNftPubData(common.FromHex(request.Pubdata))
			if err != nil {
				logx.Errorf("unable to parse deposit nft pub data: %s", err.Error())
				return err
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("unable to serialize request info: %s", err.Error())
				return err
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.AccountIndex = txInfo.AccountIndex
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeFullExit:
			txInfo, err := util.ParseFullExitPubData(common.FromHex(request.Pubdata))
			if err != nil {
				logx.Errorf("unable to parse deposit pub data: %s", err.Error())
				return err
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("unable to serialize request info : %s", err.Error())
				return err
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.AssetId = txInfo.AssetId
			mempoolTx.TxAmount = txInfo.AssetAmount.String()
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeFullExitNft:
			txInfo, err := util.ParseFullExitNftPubData(common.FromHex(request.Pubdata))
			if err != nil {
				logx.Errorf("unable to parse deposit nft pub data: %s", err.Error())
				return err
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				logx.Errorf("unable to serialize request info : %s", err.Error())
				return err
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.NftIndex = txInfo.NftIndex
			mempoolTx.TxInfo = string(txInfoBytes)
			mempoolTx.AccountIndex = txInfo.AccountIndex

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		default:
			logx.Errorf("invalid request type")
			return errors.New("invalid request type")
		}
	}

	// update db
	if err = m.PriorityRequestModel.CreateMempoolTxsAndUpdateRequests(pendingNewMempoolTxs, pendingRequests); err != nil {
		logx.Errorf("unable to create mempool pendingRequests and update priority requests, error: %s", err.Error())
		return err
	}
	return nil
}
