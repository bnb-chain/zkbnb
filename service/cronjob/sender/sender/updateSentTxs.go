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

package sender

import (
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/l1RollupTx"
	"github.com/bnb-chain/zkbas/common/model/proof"
)

func (s *Sender) UpdateSentTxs() (err error) {
	pendingSenders, err := s.l1RollupTxModel.GetL1RollupTxsByStatus(l1RollupTx.StatusPending)
	if err != nil {
		logx.Errorf("unable to get l1 tx senders by tx status: %s", err.Error())
		return err
	}

	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := s.cli.GetHeight()
	if err != nil {
		logx.Errorf("GetHeight err: %s", err.Error())
		return err
	}

	var (
		pendingUpdateSenders           []*l1RollupTx.L1RollupTx
		pendingUpdateProofSenderStatus = make(map[int64]int)
	)
	for _, pendingSender := range pendingSenders {
		txHash := pendingSender.L1TxHash
		// check if the status of tx is success
		_, isPending, err := s.cli.GetTransactionByHash(txHash)
		if err != nil {
			logx.Errorf("GetTransactionByHash err: %s", err.Error())
			continue
		}
		if isPending {
			continue
		}
		receipt, err := s.cli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("GetTransactionReceipt err: %s", err.Error())
			continue
		}

		if latestHeight < receipt.BlockNumber.Uint64()+s.Config.ChainConfig.ConfirmBlocksCount {
			continue
		}
		var (
			isValidSender bool
		)
		for _, vlog := range receipt.Logs {
			switch vlog.Topics[0].Hex() {
			case zkbasLogBlockCommitSigHash.Hex():
				var event zkbas.ZkbasBlockCommit
				if err = ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data); err != nil {
					logx.Errorf("UnpackIntoInterface err: %s", err.Error())
					return err
				}
				blockHeight := int64(event.BlockNumber)
				if blockHeight == pendingSender.L2BlockHeight {
					isValidSender = true
				}
			case zkbasLogBlockVerificationSigHash.Hex():
				var event zkbas.ZkbasBlockVerification
				if err = ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
					logx.Errorf("UnpackIntoInterface err: %s", err.Error())
					return err
				}
				blockHeight := int64(event.BlockNumber)
				if blockHeight == pendingSender.L2BlockHeight {
					isValidSender = true
				}
				pendingUpdateProofSenderStatus[blockHeight] = proof.Confirmed
			case zkbasLogBlocksRevertSigHash.Hex():
				// TODO revert
			default:
			}
		}

		if isValidSender {
			pendingSender.TxStatus = l1RollupTx.StatusHandled
			pendingUpdateSenders = append(pendingUpdateSenders, pendingSender)
		}
	}

	if err = s.l1RollupTxModel.UpdateL1RollupTxs(pendingUpdateSenders,
		pendingUpdateProofSenderStatus); err != nil {
		logx.Errorf("update sent txs error, err: %s", err.Error())
		return err
	}
	return nil
}
