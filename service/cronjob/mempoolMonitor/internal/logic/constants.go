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
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	// event status
	PendingStatus = l2TxEventMonitor.PendingStatus
	HandledStatus = l2TxEventMonitor.HandledStatus
	// tx type
	TxTypeRegisterZns    = commonTx.TxTypeRegisterZns
	TxTypeCreatePair     = commonTx.TxTypeCreatePair
	TxTypeUpdatePairRate = commonTx.TxTypeUpdatePairRate
	TxTypeDeposit        = commonTx.TxTypeDeposit
	TxTypeDepositNft     = commonTx.TxTypeDepositNft
	TxTypeFullExit       = commonTx.TxTypeFullExit
	TxTypeFullExitNft    = commonTx.TxTypeFullExitNft

	GeneralAssetType = commonAsset.GeneralAssetType
	NftAssetType     = commonAsset.NftAssetType

	Base = 10
)

var (
	ErrNotFound      = sqlx.ErrNotFound
	ZeroBigIntString = "0"
)
