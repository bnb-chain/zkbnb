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
	zecreyLegend "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/zero/basic"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/l1TxSender"
	"github.com/bnb-chain/zkbas/common/model/l2BlockEventMonitor"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

type (
	MempoolModel             = mempool.MempoolModel
	AccountModel             = account.AccountModel
	AccountHistoryModel      = account.AccountHistoryModel
	L2NftModel               = nft.L2NftModel
	L2NftHistoryModel        = nft.L2NftHistoryModel
	BlockModel               = block.BlockModel
	Block                    = block.Block
	L2BlockEventMonitorModel = l2BlockEventMonitor.L2BlockEventMonitorModel
	L2BlockEventMonitor      = l2BlockEventMonitor.L2BlockEventMonitor
	L1TxSenderModel          = l1TxSender.L1TxSenderModel
	L1TxSender               = l1TxSender.L1TxSender

	ZecreyLegendBlockCommit       = zecreyLegend.ZecreyLegendBlockCommit
	ZecreyLegendBlockVerification = zecreyLegend.ZecreyLegendBlockVerification

	MempoolTx = mempool.MempoolTx

	ProviderClient = basic.ProviderClient
)

const (
	// block status
	BlockPendingStatus   = block.StatusPending
	BlockCommittedStatus = block.StatusCommitted
	BlockVerifiedStatus  = block.StatusVerifiedAndExecuted

	L1TxSenderPendingStatus = l1TxSender.PendingStatus
	L1TxSenderHandledStatus = l1TxSender.HandledStatus

	BlockCommitEventName       = "BlockCommit"
	BlockVerificationEventName = "BlockVerification"
	BlocksRevertEventName      = "BlocksRevert"
)

var (

	// Zecrey contract logs sig
	zecreyLogBlockCommitSig       = []byte("BlockCommit(uint32)")
	zecreyLogBlockVerificationSig = []byte("BlockVerification(uint32)")
	zecreyLogBlocksRevertSig      = []byte("BlocksRevert(uint32,uint32)")

	ZecreyLogBlockCommitSigHash       = crypto.Keccak256Hash(zecreyLogBlockCommitSig)
	ZecreyLogBlockVerificationSigHash = crypto.Keccak256Hash(zecreyLogBlockVerificationSig)
	ZecreyLogBlocksRevertSigHash      = crypto.Keccak256Hash(zecreyLogBlocksRevertSig)

	ZecreyContractAbi, _ = abi.JSON(strings.NewReader(zecreyLegend.ZecreyLegendABI))
)
