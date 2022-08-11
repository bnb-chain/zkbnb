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
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/l1RollupTx"
	"github.com/bnb-chain/zkbas/common/util"
)

func (s *Sender) VerifyAndExecuteBlocks() (err error) {
	var (
		cli           = s.cli
		authCli       = s.authCli
		zkbasInstance = s.zkbasInstance
	)
	// scan l1 tx sender table for handled verified and executed height
	lastHandledBlock, getHandleErr := s.l1RollupTxModel.GetLatestHandledTx(VerifyAndExecuteTxType)
	if getHandleErr != nil && getHandleErr != errorcode.DbErrNotFound {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get latest handled block: %v", getHandleErr)
		return getHandleErr
	}
	// scan l1 tx sender table for pending verified and executed height that higher than the latest handled height
	pendingSender, getPendingErr := s.l1RollupTxModel.GetLatestPendingTx(VerifyAndExecuteTxType)
	if getPendingErr != nil && getPendingErr != errorcode.DbErrNotFound {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get latest pending blocks: %v", getPendingErr)
		return getPendingErr
	}
	// case 1: check tx status on L1
	if getHandleErr == errorcode.DbErrNotFound && getPendingErr == nil {
		_, isPending, err := cli.GetTransactionByHash(pendingSender.L1TxHash)
		// if err != nil, means we cannot get this tx by hash
		if err != nil {
			// if we cannot get it from rpc and the time over 1 min
			lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
			now := time.Now().UnixMilli()
			if now-lastUpdatedAt > s.Config.ChainConfig.MaxWaitingTime*time.Second.Milliseconds() {
				// drop the record
				err := s.l1RollupTxModel.DeleteL1RollupTx(pendingSender)
				if err != nil {
					logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to delete l1 tx sender: %v", err)
					return err
				}
				return nil
			} else {
				return nil
			}
		}
		// if it is pending, still waiting
		if isPending {
			logx.Infof("[SendVerifiedAndExecutedBlocks] tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
			return nil
		} else {
			receipt, err := cli.GetTransactionReceipt(pendingSender.L1TxHash)
			if err != nil {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get transaction receipt: %v", err)
				return err
			}
			if receipt.Status == 0 {
				logx.Infof("[SendVerifiedAndExecutedBlocks] the transaction is failure, please check: %s", pendingSender.L1TxHash)
				return nil
			}
		}
	}
	// case 2:
	if getHandleErr == nil && getPendingErr == nil {
		isSuccess, err := cli.WaitingTransactionStatus(pendingSender.L1TxHash)
		// if err != nil, means we cannot get this tx by hash
		if err != nil {
			// if we cannot get it from rpc and the time over 1 min
			lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
			if time.Now().UnixMilli()-lastUpdatedAt > s.Config.ChainConfig.MaxWaitingTime*time.Second.Milliseconds() {
				// drop the record
				if err := s.l1RollupTxModel.DeleteL1RollupTx(pendingSender); err != nil {
					logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to delete l1 tx sender: %v", err)
					return err
				}
			}
			return nil
		}
		// if it is pending, still waiting
		if !isSuccess {
			return nil
		}
	}
	// case 3:  means we haven't verified and executed new blocks, just start to commit
	var (
		start                         int64
		blocks                        []*block.Block
		pendingVerifyAndExecuteBlocks []zkbas.OldZkbasVerifyAndExecuteBlockInfo
	)
	if getHandleErr == errorcode.DbErrNotFound && getPendingErr == errorcode.DbErrNotFound {
		// get blocks from block table
		blocks, err = s.blockModel.GetBlocksForProverBetween(1, int64(s.Config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] GetBlocksForProverBetween err: %v, maxBlockCount: %d",
				err, s.Config.ChainConfig.MaxBlockCount)
			return err
		}
		pendingVerifyAndExecuteBlocks, err = ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to convert blocks to verify and execute block infos: %v", err)
			return err
		}
		start = int64(1)
	}
	if getHandleErr == nil && getPendingErr == errorcode.DbErrNotFound {
		blocks, err = s.blockModel.GetBlocksForProverBetween(lastHandledBlock.L2BlockHeight+1,
			lastHandledBlock.L2BlockHeight+int64(s.Config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get sender new blocks: %v", err)
			return err
		}
		pendingVerifyAndExecuteBlocks, err = ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to convert blocks to commit block infos: %v", err)
			return err
		}
		start = lastHandledBlock.L2BlockHeight + 1
	}

	blockProofs, err := s.proofModel.GetProofsByBlockRange(start, blocks[len(blocks)-1].BlockHeight,
		s.Config.ChainConfig.MaxBlockCount)
	if err != nil {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get proofs: %v", err)
		return err
	}
	if len(blockProofs) != len(blocks) {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] we haven't generated related proofs, please wait")
		return errors.New("[SendVerifiedAndExecutedBlocks] we haven't generated related proofs, please wait")
	}
	var proofs []*big.Int
	for _, proof := range blockProofs {
		var proofInfo *util.FormattedProof
		err = json.Unmarshal([]byte(proof.ProofInfo), &proofInfo)
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to unmarshal proof info: %v", err)
			return err
		}
		proofs = append(proofs, proofInfo.A[:]...)
		proofs = append(proofs, proofInfo.B[0][0], proofInfo.B[0][1])
		proofs = append(proofs, proofInfo.B[1][0], proofInfo.B[1][1])
		proofs = append(proofs, proofInfo.C[:]...)
	}
	gasPrice, err := s.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] failed to fetch gas price err: %v", err)
		return err
	}
	// commit blocks on-chain
	if len(pendingVerifyAndExecuteBlocks) != 0 {
		txHash, err := zkbas.VerifyAndExecuteBlocks(cli, authCli, zkbasInstance,
			pendingVerifyAndExecuteBlocks, proofs, gasPrice, s.Config.ChainConfig.GasLimit)
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] VerifyAndExecuteBlocks err: %v", err)
			return err
		}

		newRollupTx := &l1RollupTx.L1RollupTx{
			L1TxHash:      txHash,
			TxStatus:      PendingStatus,
			TxType:        VerifyAndExecuteTxType,
			L2BlockHeight: int64(pendingVerifyAndExecuteBlocks[len(pendingVerifyAndExecuteBlocks)-1].BlockHeader.BlockNumber),
		}
		isValid, err := s.l1RollupTxModel.CreateL1RollupTx(newRollupTx)
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] CreateL1TxSender err: %v", err)
			return err
		}
		if !isValid {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] cannot create new senders")
			return errors.New("[SendVerifiedAndExecutedBlocks] cannot create new senders")
		}
		logx.Errorf("[SendVerifiedAndExecutedBlocks] new blocks have been verified and executed(height): %d", newRollupTx.L2BlockHeight)
		return nil
	}
	return nil
}
