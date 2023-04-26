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
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb/common/monitor"
	"github.com/bnb-chain/zkbnb/dao/dbcache"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (m *Monitor) MonitorPriorityRequests() error {
	pendingRequests, err := m.PriorityRequestModel.GetPriorityRequestsByStatus(monitor.PendingStatus)
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
		case monitor.TxTypeDeposit:
			txInfo, err := chain.ParseDepositPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit pub data: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}
			poolTx.AssetId = txInfo.AssetId
			poolTx.TxAmount = txInfo.AssetAmount.String()
			accountIndex, err := m.GetAccountIndex(txInfo.L1Address)
			if err == nil {
				poolTx.ToAccountIndex = accountIndex
			} else {
				logx.Errorf("unable to get account index : %v", err)
			}
			NativeAccountIndex, err := m.GetAccountIndex(poolTx.NativeAddress)
			if err == nil {
				poolTx.AccountIndex = NativeAccountIndex
				poolTx.FromAccountIndex = NativeAccountIndex
			} else {
				poolTx.AccountIndex = types.NilAccountIndex
				poolTx.FromAccountIndex = types.NilAccountIndex
			}
		case monitor.TxTypeDepositNft:
			txInfo, err := chain.ParseDepositNftPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit nft pub data: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info: %v", err)
			}
			poolTx.NftIndex = txInfo.NftIndex
			poolTx.CollectionId = txInfo.CollectionId
			accountIndex, err := m.GetAccountIndex(txInfo.L1Address)
			if err == nil {
				poolTx.ToAccountIndex = accountIndex
			} else {
				logx.Errorf("unable to get account index : %v", err)
			}
			NativeAccountIndex, err := m.GetAccountIndex(poolTx.NativeAddress)
			if err == nil {
				poolTx.AccountIndex = NativeAccountIndex
				poolTx.FromAccountIndex = NativeAccountIndex
			} else {
				poolTx.AccountIndex = types.NilAccountIndex
				poolTx.FromAccountIndex = types.NilAccountIndex
			}
		case monitor.TxTypeFullExit:
			txInfo, err := chain.ParseFullExitPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit pub data: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}
			poolTx.AssetId = txInfo.AssetId
			poolTx.TxAmount = txInfo.AssetAmount.String()
			accountIndex, err := m.GetAccountIndex(txInfo.L1Address)
			if err == nil {
				poolTx.AccountIndex = accountIndex
				poolTx.FromAccountIndex = accountIndex
				poolTx.ToAccountIndex = accountIndex
			} else {
				logx.Errorf("unable to get account index : %v", err)
			}
		case monitor.TxTypeFullExitNft:
			txInfo, err := chain.ParseFullExitNftPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit nft pub data: %v", err)
			}

			poolTx.TxType = int64(txInfo.TxType)
			txInfoBytes, err = json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}
			poolTx.NftIndex = txInfo.NftIndex
			poolTx.CollectionId = txInfo.CollectionId
			accountIndex, err := m.GetAccountIndex(txInfo.L1Address)
			if err == nil {
				poolTx.AccountIndex = accountIndex
				poolTx.FromAccountIndex = accountIndex
				poolTx.ToAccountIndex = accountIndex
			} else {
				logx.Errorf("unable to get account index : %v", err)
			}
		default:
			return fmt.Errorf("invalid request type")
		}
		poolTx.TxInfo = string(txInfoBytes)
		poolTx.L1RequestId = request.RequestId
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

func (m *Monitor) GetAccountIndex(l1Address string) (int64, error) {
	cached, exist := m.L1AddressCache.Get(l1Address)
	if exist {
		return cached.(int64), nil
	} else {
		var accountIndex interface{}
		redisAccount, err := m.RedisCache.Get(context.Background(), dbcache.AccountKeyByL1Address(l1Address), &accountIndex)
		if err == nil && redisAccount != nil {
			return accountIndex.(int64), nil
		}
		dbAccount, err := m.AccountModel.GetAccountByL1Address(l1Address)
		if err != nil {
			if err == types.DbErrNotFound {
				return types.NilAccountIndex, nil
			}
			return 0, err
		}
		m.L1AddressCache.Add(l1Address, dbAccount.AccountIndex)
		return dbAccount.AccountIndex, nil
	}
}
