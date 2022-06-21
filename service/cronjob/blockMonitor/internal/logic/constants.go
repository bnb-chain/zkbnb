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
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zecreyLegend "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas/common/model/l1BlockMonitor"
	"github.com/bnb-chain/zkbas/common/model/l2BlockEventMonitor"
	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
)

type (
	ProviderClient = _rpc.ProviderClient
	AuthClient     = _rpc.AuthClient

	// event monitor
	L2TxEventMonitor    = l2TxEventMonitor.L2TxEventMonitor
	L2BlockEventMonitor = l2BlockEventMonitor.L2BlockEventMonitor

	// model
	L1BlockMonitorModel      = l1BlockMonitor.L1BlockMonitorModel
	L2TxEventMonitorModel    = l2TxEventMonitor.L2TxEventMonitorModel
	L2BlockEventMonitorModel = l2BlockEventMonitor.L2BlockEventMonitorModel
)

const (
	// zecrey event name
	EventNameNewPriorityRequest = "NewPriorityRequest"
	EventNameBlockCommit        = "BlockCommit"
	EventNameBlockVerification  = "BlockVerification"
	EventNameBlocksRevert       = "BlocksRevert"

	// tx type for l2 block event monitors
	EventTypeNewPriorityRequest = 0
	EventTypeCommittedBlock     = l2BlockEventMonitor.CommittedBlockEventType
	EventTypeVerifiedBlock      = l2BlockEventMonitor.VerifiedBlockEventType
	EventTypeRevertedBlock      = l2BlockEventMonitor.RevertedBlockEventType
	// status
	PendingStatusL2BlockEventMonitor = l2BlockEventMonitor.PendingStatus
)

var (
	// err
	ErrNotFound = sqlx.ErrNotFound
	// Zecrey contract logs sig
	zecreyLogNewPriorityRequestSig = []byte("NewPriorityRequest(address,uint64,uint8,bytes,uint256)")
	zecreyLogWithdrawalSig         = []byte("Withdrawal(uint16,uint128)")
	zecreyLogWithdrawalPendingSig  = []byte("WithdrawalPending(uint16,uint128)")
	zecreyLogBlockCommitSig        = []byte("BlockCommit(uint32)")
	zecreyLogBlockVerificationSig  = []byte("BlockVerification(uint32)")
	zecreyLogBlocksRevertSig       = []byte("BlocksRevert(uint32,uint32)")

	zecreyLogNewPriorityRequestSigHash = crypto.Keccak256Hash(zecreyLogNewPriorityRequestSig)
	ZecreyLogWithdrawalSigHash         = crypto.Keccak256Hash(zecreyLogWithdrawalSig)
	ZecreyLogWithdrawalPendingSigHash  = crypto.Keccak256Hash(zecreyLogWithdrawalPendingSig)
	ZecreyLogBlockCommitSigHash        = crypto.Keccak256Hash(zecreyLogBlockCommitSig)
	ZecreyLogBlockVerificationSigHash  = crypto.Keccak256Hash(zecreyLogBlockVerificationSig)
	ZecreyLogBlocksRevertSigHash       = crypto.Keccak256Hash(zecreyLogBlocksRevertSig)

	ZecreyContractAbi, _ = abi.JSON(strings.NewReader(zecreyLegend.ZecreyLegendABI))
)
