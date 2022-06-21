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
	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zecreyLegend "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/l1TxSender"
	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"math/big"
)

type (
	Tx                  = tx.Tx
	TxDetail            = tx.TxDetail
	Block               = block.Block
	BlockForCommit      = blockForCommit.BlockForCommit
	L1TxSenderModel     = l1TxSender.L1TxSenderModel
	L1TxSender          = l1TxSender.L1TxSender
	BlockModel          = block.BlockModel
	BlockForCommitModel = blockForCommit.BlockForCommitModel

	ProviderClient = _rpc.ProviderClient
	AuthClient     = _rpc.AuthClient
	ZecreyLegend   = zecreyLegend.ZecreyLegend

	ZecreyLegendCommitBlockInfo = zecreyLegend.OldZecreyLegendCommitBlockInfo
	ZecreyLegendVerifyBlockInfo = zecreyLegend.OldZecreyLegendVerifyAndExecuteBlockInfo
	StorageStoredBlockInfo      = zecreyLegend.StorageStoredBlockInfo

	L2AssetInfoModel = assetInfo.AssetInfoModel

	ProofSenderModel = proofSender.ProofSenderModel
)

const (
	StatusPending             = block.StatusPending
	StatusCommitted           = block.StatusCommitted
	StatusVerifiedAndExecuted = block.StatusVerifiedAndExecuted

	PendingStatus          = l1TxSender.PendingStatus
	CommitTxType           = l1TxSender.CommitTxType
	VerifyAndExecuteTxType = l1TxSender.VerifyAndExecuteTxType
)

var (
	ErrNotFound = sqlx.ErrNotFound
)

type SenderParam struct {
	Cli                  *ProviderClient
	AuthCli              *AuthClient
	ZecreyLegendInstance *ZecreyLegend
	MaxWaitingTime       int64
	MaxBlocksCount       int
	GasPrice             *big.Int
	GasLimit             uint64
}
