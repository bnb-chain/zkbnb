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
	"errors"
	"fmt"
	"github.com/zecrey-labs/zecrey-eth-rpc/zecrey/core/zecrey/basic"
	"io"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
)

func SendCommittedBlocks(
	param *SenderParam,
	l1TxSenderModel L1TxSenderModel,
	blockModel BlockModel,
) (err error) {

	var (
		cli            = param.Cli
		chainId        = param.ChainId
		mainChainId    = param.MainChainId
		maxBlockCount  = param.MaxBlockCount
		maxWaitingTime = param.MaxWaitingTime
	)

	// scan l1 tx sender table for handled committed height
	lastHandledBlock, err := l1TxSenderModel.GetLatestHandledBlock(chainId, CommitTxType)
	if err != nil {
		logx.Errorf("[SendCommittedBlocks] unable to get latest handled block: %s", err.Error())
		return err
	}

	// mainChain mode: checking mainChain status
	var mainChainLastHandledHeight int64
	if param.Mode == MainChain {
		// if this chain is not main chain, check if main chain has been committed
		if chainId != mainChainId {
			mainChainLastHandledBlock, err := l1TxSenderModel.GetLatestHandledBlock(mainChainId, CommitTxType)
			if err != nil {
				logx.Errorf("[SendCommittedBlocks] unable to get latest handled block: %s", err.Error())
				return err
			}
			if mainChainLastHandledBlock == nil {
				logx.Info("[SendCommittedBlocks] main chain has not been committed, should wait")
				return nil
			}
			// check main chain committed height
			mainChainLastHandledHeight = mainChainLastHandledBlock.L2BlockHeight
		}
	}

	var lastStoredBlockHeader StorageBlockHeader
	var pendingCommittedBlocks []ZecreyCommitBlockInfo
	// if lastHandledBlock == nil, means we haven't committed any blocks, just start from 0
	if lastHandledBlock == nil {
		// scan l1 tx sender table for pending committed height that higher than the latest handled height
		rowsAffected, pendingSenders, err := l1TxSenderModel.GetLatestPendingBlocks(chainId, CommitTxType)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to get latest pending blocks: %s", err.Error())
			return err
		}

		// if rowsAffected == 0, means we haven't committed new blocks, just start to commit
		if rowsAffected == 0 {
			// get blocks from block table
			var blocks []*Block
			if param.Mode == MainChain && (chainId != mainChainId) {
				blocks, err = blockModel.GetBlocksForSenderBetween(0, mainChainLastHandledHeight, StatusPending, maxBlockCount)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to get latest handled block: %s", err.Error())
					return err
				}
			} else {
				blocks, err = blockModel.GetBlocksForSender(StatusPending, maxBlockCount)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to get latest handled block: %s", err.Error())
					return err
				}
			}
			pendingCommittedBlocks, err = ConvertBlocksToCommitBlockInfos(blocks, chainId)
			if err != nil {
				logx.Errorf("[SendCommittedBlocks] unable to convert blocks to commit block infos: %s", err.Error())
				return err
			}
			// set stored block header to default 0
			lastStoredBlockHeader = DefaultBlockHeader()
		} else {
			if rowsAffected != 1 {
				logx.Errorf("[SendCommittedBlocks] a terrible error, need to check!!! ")
				return errors.New("[SendCommittedBlocks] a terrible error, need to check!!! ")
			}
			// if rowsAffected != 0, check if the tx is pending or fail, if fail re-send it.
			pendingSender := pendingSenders[0]
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
						logx.Errorf("[SendCommittedBlocks] unable to delete l1 tx sender: %s", err.Error())
						return err
					}
					return nil
				} else {
					logx.Infof("[SendCommittedBlocks] tx cannot be found, but not exceed time limit %s", pendingSender.L1TxHash)
					return nil
				}
			}
			// if it is pending, still waiting
			if isPending {
				logx.Infof("[SendCommittedBlocks] tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
				return nil
			}
		}
	} else { // if lastHandledBlock != nil
		// scan l1 tx sender table for pending committed height that higher than the latest handled height
		rowsAffected, pendingSenders, err := l1TxSenderModel.GetLatestPendingBlocks(chainId, CommitTxType)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to get latest pending blocks: %s", err.Error())
			return err
		}
		// if rowsAffected == 0, means we haven't committed new blocks, just start to commit
		if rowsAffected == 0 {
			// get blocks higher than last handled blocks
			var blocks []*Block
			if param.Mode == MainChain && (chainId != mainChainId) {
				blocks, err = blockModel.GetBlocksForSenderBetween(lastHandledBlock.L2BlockHeight, mainChainLastHandledHeight, StatusPending, maxBlockCount)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to get sender new blocks: %s", err.Error())
					return err
				}
			} else {
				// if the main chain want to commit again, it should confirm that the last committed blocks has been executed
				logx.Info("lastHandledBlock.L2BlockHeight", lastHandledBlock.L2BlockHeight)
				blockInfo, err := blockModel.GetBlockByBlockHeight(lastHandledBlock.L2BlockHeight)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to get block by height: %s", err.Error())
					return err
				}
				if param.Mode == MainChain && blockInfo.BlockStatus != StatusExecuted {
					logx.Info("[SendCommittedBlocks] last block has not been executed, should wait")
					return nil
				}
				// else commit new blocks
				blocks, err = blockModel.GetBlocksForSenderHigherThanBlockHeight(lastHandledBlock.L2BlockHeight, StatusPending, maxBlockCount)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to get sender new blocks: %s", err.Error())
					return err
				}
			}
			pendingCommittedBlocks, err = ConvertBlocksToCommitBlockInfos(blocks, chainId)
			if err != nil {
				logx.Errorf("[SendCommittedBlocks] unable to convert blocks to commit block infos: %s", err.Error())
				return err
			}
			// get last block info
			lastHandledBlockInfo, err := blockModel.GetBlockByBlockHeight(lastHandledBlock.L2BlockHeight)
			if err != nil && err != ErrNotFound {
				logx.Errorf("[SendCommittedBlocks] unable to get last handled block info: %s", err.Error())
				return err
			}
			// construct last stored block header
			lastStoredBlockHeader = StorageBlockHeader{
				BlockNumber:    uint32(lastHandledBlockInfo.BlockHeight),
				OnchainOpsRoot: basic.SetFixed32Bytes(common.FromHex(lastHandledBlockInfo.OnChainOpsRoot)),
				AccountRoot:    basic.SetFixed32Bytes(common.FromHex(lastHandledBlockInfo.AccountRoot)),
				Timestamp:      big.NewInt(lastHandledBlockInfo.CreatedAt.UnixMilli()),
				Commitment:     basic.SetFixed32Bytes(common.FromHex(lastHandledBlockInfo.BlockCommitment)),
			}
		} else {
			if rowsAffected != 1 {
				logx.Errorf("[SendCommittedBlocks] a terrible error, need to check!!! ")
				return errors.New("[SendCommittedBlocks] a terrible error, need to check!!! ")
			}
			// if rowsAffected != 0, check if the tx is pending or fail, if fail re-send it.
			pendingSender := pendingSenders[0]
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
						logx.Errorf("[SendCommittedBlocks] unable to delete l1 tx sender: %s", err.Error())
						return err
					}
					return nil
				} else {
					logx.Infof("[SendCommittedBlocks] tx cannot be found, but not exceed time limit: %s", pendingSender.L1TxHash)
					return nil
				}
			}
			// if it is pending, still waiting
			if !isSuccess {
				logx.Infof("[SendCommittedBlocks] tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
				return nil
			}
		}
	}
	// commit blocks on-chain
	if len(pendingCommittedBlocks) != 0 {
		if param.DebugParams != nil {
			// file output start
			var (
				f        *os.File
				filename = param.DebugParams.FilePrefix + "/commit.txt"
			)
			if utils.CheckFileIsExist(filename) {
				f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				if err != nil {
					panic("OpenFile error")
				}
				fmt.Println("file exists")
			} else {
				f, err = os.Create(filename)
				fmt.Println("file doesn't exists")
				if err != nil {
					panic("Create file error")
				}
			}
			defer f.Close()

			common.Bytes2Hex(lastStoredBlockHeader.AccountRoot[:])
			common.Bytes2Hex(lastStoredBlockHeader.Commitment[:])
			common.Bytes2Hex(lastStoredBlockHeader.OnchainOpsRoot[:])
			_, err = io.WriteString(
				f,
				fmt.Sprintf("lastStoredBlockHeader: \n"+
					"BlockNumber : %d\n"+
					"Timestamp : %d\n"+
					"AccountRoot : %s\n"+
					"Commitment : %s\n"+
					"OnchainOpsRoot : %s\n",
					lastStoredBlockHeader.BlockNumber,
					lastStoredBlockHeader.Timestamp,
					common.Bytes2Hex(lastStoredBlockHeader.AccountRoot[:]),
					common.Bytes2Hex(lastStoredBlockHeader.Commitment[:]),
					common.Bytes2Hex(lastStoredBlockHeader.OnchainOpsRoot[:])),
			)

			for _, v := range pendingCommittedBlocks {
				_, err = io.WriteString(
					f,
					fmt.Sprintf("pendingCommittedBlocks: \n"+
						"BlockNumber : %d\n"+
						"Timestamp : %d\n"+
						"NewAccountRoot : %s\n"+
						"Commitment : %s\n"+
						"OnchainOpsCount : %d\n"+
						"OnchainOpsRoot : %s\n",

						v.BlockNumber,
						v.Timestamp,
						common.Bytes2Hex(v.NewAccountRoot[:]),
						common.Bytes2Hex(v.Commitment[:]),
						v.OnchainOpsCount,
						common.Bytes2Hex(v.OnchainOpsRoot[:])),
				)
			}

			// file output end
		}

		txHash, err := zecreyLegend.ZecreyCommitBlocks(
			cli, param.AuthCli, param.ZecreyInstance, lastStoredBlockHeader, pendingCommittedBlocks, param.GasPrice, param.GasLimit,
		)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to commit blocks: %s", err.Error())
			return err
		}
		log.Println("stored block header")
		log.Println(lastStoredBlockHeader.BlockNumber)
		log.Println(common.Bytes2Hex(lastStoredBlockHeader.AccountRoot[:]))
		log.Println(lastStoredBlockHeader.Timestamp.String())
		log.Println(common.Bytes2Hex(lastStoredBlockHeader.Commitment[:]))
		log.Println("committed blocks")
		for _, pendingCommittedBlock := range pendingCommittedBlocks {
			log.Println(pendingCommittedBlock.BlockNumber)
		}
		// update l1 tx sender table records
		newSender := &L1TxSender{
			L1TxHash:      txHash,
			TxStatus:      PendingStatus,
			TxType:        CommitTxType,
			L2BlockHeight: int64(pendingCommittedBlocks[len(pendingCommittedBlocks)-1].BlockNumber),
		}
		isValid, err := l1TxSenderModel.CreateL1TxSender(newSender)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to create l1 tx sender")
			return err
		}
		if !isValid {
			logx.Errorf("[SendCommittedBlocks] cannot create new senders")
			return errors.New("[SendCommittedBlocks] cannot create new senders")
		}
		logx.Infof("[SendCommittedBlocks] new blocks have been committed(height): %v", newSender.L2BlockHeight)
		return nil
	} else {
		logx.Infof("[SendCommittedBlocks] no new blocks need to commit")
		return nil
	}
}
