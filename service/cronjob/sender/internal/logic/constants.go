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
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	"github.com/zecrey-labs/zecrey-eth-rpc/zecreyContract/core/zecrey/basic"
	"github.com/zecrey-labs/zecrey/common/commonTx"
	"github.com/zecrey-labs/zecrey/common/model/block"
	"github.com/zecrey-labs/zecrey/common/model/blockForProver"
	"github.com/zecrey-labs/zecrey/common/model/l1TxSender"
	"github.com/zecrey-labs/zecrey/common/model/l1asset"
	"github.com/zecrey-labs/zecrey/common/model/l2asset"
	"github.com/zecrey-labs/zecrey/common/model/proofSender"
	"github.com/zecrey-labs/zecrey/common/model/tx"
	"github.com/zecrey-labs/zecrey/common/utils"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"math/big"
)

type (
	Tx              = tx.Tx
	TxDetail        = tx.TxDetail
	Block           = block.Block
	L1TxSenderModel = l1TxSender.L1TxSenderModel
	L1TxSender      = l1TxSender.L1TxSender
	BlockModel      = block.BlockModel
	BlockDetail     = block.BlockDetail

	ProviderClient = _rpc.ProviderClient
	AuthClient     = _rpc.AuthClient
	Zecrey         = basic.Zecrey

	ZecreyCommitBlockInfo = basic.ZecreyCommitBlockInfo
	StorageBlockHeader    = basic.StorageBlockHeader

	L2AssetInfoModel = l2asset.L2AssetInfoModel
	L1AssetInfoModel = l1asset.L1AssetInfoModel

	BlockForProverModel = blockForProver.BlockForProverModel
	ProofSenderModel    = proofSender.ProofSenderModel
)

const (
	StatusPending   = block.StatusPending
	StatusCommitted = block.StatusCommitted
	StatusVerified  = block.StatusVerified
	StatusExecuted  = block.StatusExecuted

	PendingStatus = l1TxSender.PendingStatus
	CommitTxType  = l1TxSender.CommitTxType
	VerifyTxType  = l1TxSender.VerifyTxType
	ExecuteTxType = l1TxSender.ExecuteTxType

	TxTypeDeposit  = commonTx.TxTypeDeposit
	TxTypeLock     = commonTx.TxTypeLock
	TxTypeWithdraw = commonTx.TxTypeWithdraw

	OnChainOpsTreeLevel = 6
)

const (
	MainChain = iota
	StandAlone
)

var (
	AddressType, _ = abi.NewType("address", "", nil)
	Bytes32Type, _ = abi.NewType("bytes32", "", nil)
	Uint8Type, _   = abi.NewType("uint8", "", nil)
	Uint16Type, _  = abi.NewType("uint16", "", nil)
	Uint32Type, _  = abi.NewType("uint32", "", nil)
	Uint128Type, _ = abi.NewType("uint128", "", nil)

	ErrNotFound = sqlx.ErrNotFound
)

type SenderParam struct {
	Cli            *ProviderClient
	AuthCli        *AuthClient
	ZecreyInstance *Zecrey
	ChainId        int64
	Mode           int64
	MaxWaitingTime int64
	MaxBlockCount  int
	MainChainId    int64
	GasPrice       *big.Int
	GasLimit       uint64
	DebugParams    *utils.DebugOptions
}
