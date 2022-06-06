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
	"encoding/json"
	"errors"
	zecreyLegend "github.com/zecrey-labs/zecrey-eth-rpc/zecrey/core/zecrey-legend"
	"math/big"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func SendVerifiedAndExecutedBlocks(
	param *SenderParam,
	l1TxSenderModel L1TxSenderModel,
	blockModel BlockModel,
	proofSenderModel ProofSenderModel,
) (err error) {

	var (
		cli                  = param.Cli
		authCli              = param.AuthCli
		zecreyLegendInstance = param.ZecreyLegendInstance
		gasPrice             = param.GasPrice
		gasLimit             = param.GasLimit
		maxBlockCount        = param.MaxBlocksCount
		maxWaitingTime       = param.MaxWaitingTime
	)

	// scan l1 tx sender table for handled verified and executed height
	lastHandledBlock, err := l1TxSenderModel.GetLatestHandledBlock(VerifyAndExecuteTxType)
	if err != nil {
		if err != ErrNotFound {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get latest handled block: %s", err.Error())
			return err
		}
	}

	var (
		pendingVerifyAndExecuteBlocks []ZecreyLegendVerifyBlockInfo
		proofs                        []*big.Int
	)
	// if lastHandledBlock == nil, means we haven't verified and executed any blocks, just start from 0
	if err == ErrNotFound {
		// scan l1 tx sender table for pending verified and executed height that higher than the latest handled height
		pendingSender, err := l1TxSenderModel.GetLatestPendingBlock(VerifyAndExecuteTxType)
		if err != nil {
			if err != ErrNotFound {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get latest pending blocks: %s", err.Error())
				return err
			}
		}

		// if ErrNotFound, means we haven't verified and executed new blocks, just start to commit
		if err == ErrNotFound {
			// get blocks from block table
			blocks, err := blockModel.GetBlocksForProverBetween(1, int64(maxBlockCount))
			if err != nil {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get blocks: %s", err.Error())
				return err
			}
			end := blocks[len(blocks)-1].BlockHeight
			pendingVerifyAndExecuteBlocks, err = ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
			if err != nil {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to convert blocks to verify and execute block infos: %s", err.Error())
				return err
			}
			// get proofs
			proofSenders, err := proofSenderModel.GetProofsBetween(1, end)
			if err != nil {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get proofs: %s", err.Error())
				return err
			}
			if len(proofSenders) != len(blocks) {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] we haven't generated related proofs, please wait")
				return errors.New("[SendVerifiedAndExecutedBlocks] we haven't generated related proofs, please wait")
			}
			for _, proofSender := range proofSenders {
				var proofInfo []*big.Int
				err = json.Unmarshal([]byte(proofSender.ProofInfo), &proofInfo)
				if err != nil {
					logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to unmarshal proof info: %s", err.Error())
					return err
				}
				proofs = append(proofs, proofInfo...)
			}
		} else {
			_, isPending, err := cli.GetTransactionByHash(pendingSender.L1TxHash)
			// if err != nil, means we cannot get this tx by hash
			if err != nil {
				// if we cannot get it from rpc and the time over 1 min
				lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
				now := time.Now().UnixMilli()
				if now-lastUpdatedAt > maxWaitingTime {
					// drop the record
					err := l1TxSenderModel.DeleteL1TxSender(pendingSender)
					if err != nil {
						logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to delete l1 tx sender: %s", err.Error())
						return err
					}
					return nil
				} else {
					logx.Infof("[SendVerifiedAndExecutedBlocks] tx cannot be found, but not exceed time limit %s", pendingSender.L1TxHash)
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
					logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get transaction receipt: %s", err.Error())
					return err
				}
				if receipt.Status == 0 {
					logx.Infof("[SendVerifiedAndExecutedBlocks] the transaction is failure, please check: %s", pendingSender.L1TxHash)
					return nil
				}
			}
		}
	} else { // if lastHandledBlock != nil
		// scan l1 tx sender table for pending verified and executed height that higher than the latest handled height
		pendingSender, err := l1TxSenderModel.GetLatestPendingBlock(VerifyAndExecuteTxType)
		if err != nil {
			if err != ErrNotFound {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get latest pending blocks: %s", err.Error())
				return err
			}
		}
		// if ErrNotFound, means we haven't verified and executed new blocks, just start to commit
		if err == ErrNotFound {
			// get blocks higher than last handled blocks
			var blocks []*Block
			// commit new blocks
			blocks, err = blockModel.GetBlocksForProverBetween(lastHandledBlock.L2BlockHeight+1, lastHandledBlock.L2BlockHeight+int64(maxBlockCount))
			if err != nil {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get sender new blocks: %s", err.Error())
				return err
			}
			pendingVerifyAndExecuteBlocks, err = ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
			if err != nil {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to convert blocks to commit block infos: %s", err.Error())
				return err
			}
			end := blocks[len(blocks)-1].BlockHeight
			// get proofs
			proofSenders, err := proofSenderModel.GetProofsBetween(lastHandledBlock.L2BlockHeight+1, end)
			if err != nil {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get proofs: %s", err.Error())
				return err
			}
			if len(proofSenders) != len(blocks) {
				logx.Errorf("[SendVerifiedAndExecutedBlocks] we haven't generated related proofs, please wait")
				return errors.New("[SendVerifiedAndExecutedBlocks] we haven't generated related proofs, please wait")
			}
			for _, proofSender := range proofSenders {
				var proofInfo []*big.Int
				err = json.Unmarshal([]byte(proofSender.ProofInfo), &proofInfo)
				if err != nil {
					logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to unmarshal proof info: %s", err.Error())
					return err
				}
				proofs = append(proofs, proofInfo...)
			}
		} else {
			isSuccess, err := cli.WaitingTransactionStatus(pendingSender.L1TxHash)
			// if err != nil, means we cannot get this tx by hash
			if err != nil {
				// if we cannot get it from rpc and the time over 1 min
				lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
				now := time.Now().UnixMilli()
				if now-lastUpdatedAt > maxWaitingTime {
					// drop the record
					err := l1TxSenderModel.DeleteL1TxSender(pendingSender)
					if err != nil {
						logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to delete l1 tx sender: %s", err.Error())
						return err
					}
					return nil
				} else {
					logx.Infof("[SendVerifiedAndExecutedBlocks] tx cannot be found, but not exceed time limit: %s", pendingSender.L1TxHash)
					return nil
				}
			}
			// if it is pending, still waiting
			if !isSuccess {
				logx.Infof("[SendVerifiedAndExecutedBlocks] tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
				return nil
			}
		}
	}
	// commit blocks on-chain
	if len(pendingVerifyAndExecuteBlocks) != 0 {
		txHash, err := zecreyLegend.VerifyAndExecuteBlocks(
			cli, authCli,
			zecreyLegendInstance,
			pendingVerifyAndExecuteBlocks,
			proofs,
			gasPrice,
			gasLimit,
		)
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to commit blocks: %s", err.Error())
			return err
		}
		for _, pendingBlock := range pendingVerifyAndExecuteBlocks {
			logx.Infof("[SendVerifiedAndExecutedBlocks] verified and executed block: %v", pendingBlock.BlockHeader.BlockNumber)
		}
		// update l1 tx sender table records
		newSender := &L1TxSender{
			L1TxHash:      txHash,
			TxStatus:      PendingStatus,
			TxType:        VerifyAndExecuteTxType,
			L2BlockHeight: int64(pendingVerifyAndExecuteBlocks[len(pendingVerifyAndExecuteBlocks)-1].BlockHeader.BlockNumber),
		}
		isValid, err := l1TxSenderModel.CreateL1TxSender(newSender)
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to create l1 tx sender")
			return err
		}
		if !isValid {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] cannot create new senders")
			return errors.New("[SendVerifiedAndExecutedBlocks] cannot create new senders")
		}
		logx.Infof("[SendVerifiedAndExecutedBlocks] new blocks have been verified and executed(height): %v", newSender.L2BlockHeight)
		return nil
	} else {
		logx.Infof("[SendVerifiedAndExecutedBlocks] no new blocks need to commit")
		return nil
	}
}
