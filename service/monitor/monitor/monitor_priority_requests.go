/*
 * Copyright © 2021 ZkBNB Protocol
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
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (m *Monitor) MonitorPriorityRequests() error {
	pendingRequests, err := m.PriorityRequestModel.GetPriorityRequestsByStatus(PendingStatus)
	if err != nil {
		if err != types.DbErrNotFound {
			return err
		}
		return nil
	}
	var (
		pendingNewPoolTxs []*tx.PoolTx
	)
	// get last handled request id
	currentRequestId, err := m.PriorityRequestModel.GetLatestHandledRequestId()
	if err != nil {
		return fmt.Errorf("unable to get last handled request id, err: %v", err)
	}

	for _, request := range pendingRequests {
		logx.Infof("process pending priority request, requestId=%d", request.RequestId)
		// request id must be in order
		if request.RequestId != currentRequestId+1 {
			return fmt.Errorf("invalid request id, requestId=%d, expected=%d", request.RequestId, currentRequestId+1)
		}
		currentRequestId++

		txHash := ComputeL1TxTxHash(request.RequestId, request.L1TxHash)

		poolTx := &tx.PoolTx{BaseTx: tx.BaseTx{
			TxHash:       txHash,
			AccountIndex: types.NilAccountIndex,
			Nonce:        types.NilNonce,
			ExpiredAt:    types.NilExpiredAt,

			GasFeeAssetId: types.NilAssetId,
			GasFee:        types.NilAssetAmount,
			NftIndex:      types.NilNftIndex,
			CollectionId:  types.NilCollectionNonce,
			AssetId:       types.NilAssetId,
			TxAmount:      types.NilAssetAmount,
			NativeAddress: request.SenderAddress,

			BlockHeight: types.NilBlockHeight,
			TxStatus:    tx.StatusPending,
		}}

		request.L2TxHash = txHash

		// handle request based on request type
		var txInfoBytes []byte
		switch request.TxType {
		case TxTypeRegisterZns:
			// parse request info
			txInfo, err := chain.ParseRegisterZnsPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse registerZNS pub data, err: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return err
			}

		case TxTypeDeposit:
			txInfo, err := chain.ParseDepositPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit pub data: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}

		case TxTypeDepositNft:
			txInfo, err := chain.ParseDepositNftPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit nft pub data: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info: %v", err)
			}

		case TxTypeFullExit:
			txInfo, err := chain.ParseFullExitPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit pub data: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}

		case TxTypeFullExitNft:
			txInfo, err := chain.ParseFullExitNftPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit nft pub data: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}

		default:
			return fmt.Errorf("invalid request type")
		}

		poolTx.TxInfo = string(txInfoBytes)
		pendingNewPoolTxs = append(pendingNewPoolTxs, poolTx)
	}

	// update db
	err = m.db.Transaction(func(tx *gorm.DB) error {
		// create pool txs
		err = m.TxPoolModel.CreateTxsInTransact(tx, pendingNewPoolTxs)
		if err != nil {
			return err
		}

		// update priority request status
		err := m.PriorityRequestModel.UpdateHandledPriorityRequestsInTransact(tx, pendingRequests)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to create pool tx and update priority requests, error: %v", err)
	}

	for _, request := range pendingRequests {
		priorityOperationMetric.Set(float64(request.RequestId))
		priorityOperationHeightMetric.Set(float64(request.L1BlockHeight))
	}

	return nil
}
