/*
 * Copyright © 2021 ZkBAS Protocol
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

	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/types"
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
		pendingNewMempoolTxs []*mempool.MempoolTx
	)
	// get last handled request id
	currentRequestId, err := m.PriorityRequestModel.GetLatestHandledRequestId()
	if err != nil {
		return fmt.Errorf("unable to get last handled request id, err: %v", err)
	}

	for _, request := range pendingRequests {
		// request id must be in order
		if request.RequestId != currentRequestId+1 {
			return fmt.Errorf("invalid request id")
		}
		currentRequestId++

		txHash := ComputeL1TxTxHash(request.RequestId, request.L1TxHash)

		mempoolTx := &mempool.MempoolTx{
			TxHash:        txHash,
			GasFeeAssetId: types.NilAssetId,
			GasFee:        types.NilAssetAmount,
			NftIndex:      types.NilNftIndex,
			PairIndex:     types.NilPairIndex,
			AssetId:       types.NilAssetId,
			TxAmount:      types.NilAssetAmount,
			NativeAddress: request.SenderAddress,
			AccountIndex:  types.NilAccountIndex,
			Nonce:         types.NilNonce,
			ExpiredAt:     types.NilExpiredAt,
			Status:        mempool.PendingTxStatus,
		}
		// handle request based on request type
		switch request.TxType {
		case TxTypeRegisterZns:
			// parse request info
			txInfo, err := chain.ParseRegisterZnsPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse registerZNS pub data, err: %v", err)
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				return err
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeCreatePair:
			txInfo, err := chain.ParseCreatePairPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse registerZNS pub data: %v", err)
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeUpdatePairRate:
			txInfo, err := chain.ParseUpdatePairRatePubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse update pair rate pub data: %v", err)
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeDeposit:
			txInfo, err := chain.ParseDepositPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit pub data: %v", err)
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeDepositNft:
			txInfo, err := chain.ParseDepositNftPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit nft pub data: %v", err)
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info: %v", err)
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeFullExit:
			txInfo, err := chain.ParseFullExitPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit pub data: %v", err)
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		case TxTypeFullExitNft:
			txInfo, err := chain.ParseFullExitNftPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit nft pub data: %v", err)
			}

			txInfoBytes, err := json.Marshal(txInfo)
			if err != nil {
				return fmt.Errorf("unable to serialize request info : %v", err)
			}

			mempoolTx.TxType = int64(txInfo.TxType)
			mempoolTx.TxInfo = string(txInfoBytes)

			pendingNewMempoolTxs = append(pendingNewMempoolTxs, mempoolTx)
		default:
			return fmt.Errorf("invalid request type")
		}
	}

	// update db
	if err = m.PriorityRequestModel.CreateMempoolTxsAndUpdateRequests(pendingNewMempoolTxs, pendingRequests); err != nil {
		return fmt.Errorf("unable to create mempool pendingRequests and update priority requests, error: %v", err)
	}
	return nil
}
