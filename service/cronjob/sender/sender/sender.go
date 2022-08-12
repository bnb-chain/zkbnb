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

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/l1RollupTx"
	"github.com/bnb-chain/zkbas/common/model/proof"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/sysConfigName"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/config"
)

const (
	PendingStatus          = l1RollupTx.StatusPending
	CommitTxType           = l1RollupTx.TxTypeCommit
	VerifyAndExecuteTxType = l1RollupTx.TxTypeVerifyAndExecute
)

type Sender struct {
	Config config.Config

	// Client
	cli           *_rpc.ProviderClient
	authCli       *_rpc.AuthClient
	zkbasInstance *zkbas.Zkbas

	// Data access objects
	blockModel          block.BlockModel
	blockForCommitModel blockForCommit.BlockForCommitModel
	l1RollupTxModel     l1RollupTx.L1RollupTxModel
	sysConfigModel      sysconfig.SysConfigModel
	proofModel          proof.ProofModel
}

func NewSender(c config.Config) *Sender {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %v", err)
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)

	s := &Sender{
		Config:              c,
		blockModel:          block.NewBlockModel(conn, c.CacheRedis, gormPointer),
		blockForCommitModel: blockForCommit.NewBlockForCommitModel(conn, c.CacheRedis, gormPointer),
		l1RollupTxModel:     l1RollupTx.NewL1RollupTxModel(conn, c.CacheRedis, gormPointer),
		sysConfigModel:      sysconfig.NewSysConfigModel(conn, c.CacheRedis, gormPointer),
		proofModel:          proof.NewProofModel(gormPointer),
	}

	l1RPCEndpoint, err := s.sysConfigModel.GetSysConfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("fatal error, cannot fetch l1RPCEndpoint from sysConfig, err: %v, SysConfigName: %s",
			err, c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	rollupAddress, err := s.sysConfigModel.GetSysConfigByName(sysConfigName.ZkbasContract)
	if err != nil {
		logx.Severef("fatal error, cannot fetch rollupAddress from sysConfig, err: %v, SysConfigName: %s",
			err, sysConfigName.ZkbasContract)
		panic(err)
	}

	s.cli, err = _rpc.NewClient(l1RPCEndpoint.Value)
	if err != nil {
		panic(err)
	}
	var chainId *big.Int
	if c.ChainConfig.L1ChainId == "" {
		chainId, err = s.cli.ChainID(context.Background())
		if err != nil {
			panic(err)
		}
	} else {
		var (
			isValid bool
		)
		chainId, isValid = new(big.Int).SetString(c.ChainConfig.L1ChainId, 10)
		if !isValid {
			panic("invalid l1 chain id")
		}
	}

	s.authCli, err = _rpc.NewAuthClient(s.cli, c.ChainConfig.Sk, chainId)
	if err != nil {
		panic(err)
	}
	s.zkbasInstance, err = zkbas.LoadZkbasInstance(s.cli, rollupAddress.Value)
	if err != nil {
		panic(err)
	}
	return s
}

func (s *Sender) CommitBlocks() (err error) {
	var (
		cli           = s.cli
		authCli       = s.authCli
		zkbasInstance = s.zkbasInstance
	)
	// scan l1 tx sender table for handled committed height
	lastHandledBlock, getHandleErr := s.l1RollupTxModel.GetLatestHandledTx(CommitTxType)
	if getHandleErr != nil && getHandleErr != errorcode.DbErrNotFound {
		logx.Errorf("GetLatestHandledBlock err: %v", getHandleErr)
		return getHandleErr
	}
	// scan l1 tx sender table for pending committed height that higher than the latest handled height
	pendingSender, getPendingErr := s.l1RollupTxModel.GetLatestPendingTx(CommitTxType)
	if getPendingErr != nil {
		if getPendingErr != errorcode.DbErrNotFound {
			logx.Errorf("GetLatestPendingBlock err: %v", getPendingErr)
			return getPendingErr
		}
	}

	// case 1:
	if getHandleErr == errorcode.DbErrNotFound && getPendingErr == nil {
		_, isPending, err := cli.GetTransactionByHash(pendingSender.L1TxHash)
		// if err != nil, means we cannot get this tx by hash
		if err != nil {
			// if we cannot get it from rpc and the time over 1 min
			lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
			now := time.Now().UnixMilli()
			if now-lastUpdatedAt > s.Config.ChainConfig.MaxWaitingTime*time.Second.Milliseconds() {
				err := s.l1RollupTxModel.DeleteL1RollupTx(pendingSender)
				if err != nil {
					logx.Errorf("unable to delete l1 tx sender: %v", err)
					return err
				}
				return nil
			} else {
				return nil
			}
		}
		// if it is pending, still waiting
		if isPending {
			logx.Infof("tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
			return nil
		} else {
			receipt, err := cli.GetTransactionReceipt(pendingSender.L1TxHash)
			if err != nil {
				logx.Errorf("unable to get transaction receipt: %v", err)
				return err
			}
			if receipt.Status == 0 {
				logx.Infof("the transaction is failure, please check: %s", pendingSender.L1TxHash)
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
			now := time.Now().UnixMilli()
			if now-lastUpdatedAt > s.Config.ChainConfig.MaxWaitingTime*time.Second.Milliseconds() {
				// drop the record
				err := s.l1RollupTxModel.DeleteL1RollupTx(pendingSender)
				if err != nil {
					logx.Errorf("unable to delete l1 tx sender: %v", err)
					return err
				}
				return nil
			} else {
				logx.Infof("tx cannot be found, but not exceed time limit: %s", pendingSender.L1TxHash)
				return nil
			}
		}
		// if it is pending, still waiting
		if !isSuccess {
			logx.Infof("tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
			return nil
		}
	}

	// case 3:
	var lastStoredBlockInfo zkbas.StorageStoredBlockInfo
	var pendingCommitBlocks []zkbas.OldZkbasCommitBlockInfo
	// if lastHandledBlock == nil, means we haven't committed any blocks, just start from 0
	// if errorcode.DbErrNotFound, means we haven't committed new blocks, just start to commit
	if getHandleErr == errorcode.DbErrNotFound && getPendingErr == errorcode.DbErrNotFound {
		var blocks []*blockForCommit.BlockForCommit
		blocks, err = s.blockForCommitModel.GetBlockForCommitBetween(1, int64(s.Config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("GetBlockForCommitBetween err: %v, maxBlockCount: %d",
				err, s.Config.ChainConfig.MaxBlockCount)
			return err
		}
		pendingCommitBlocks, err = ConvertBlocksForCommitToCommitBlockInfos(blocks)
		if err != nil {
			logx.Errorf("unable to convert blocks to commit block infos: %vv", err)
			return err
		}
		// set stored block header to default 0
		lastStoredBlockInfo = DefaultBlockHeader()
	}
	if getHandleErr == nil && getPendingErr == errorcode.DbErrNotFound {
		// if errorcode.DbErrNotFound, means we haven't committed new blocks, just start to commit
		// get blocks higher than last handled blocks
		var blocks []*blockForCommit.BlockForCommit
		// commit new blocks
		blocks, err = s.blockForCommitModel.GetBlockForCommitBetween(lastHandledBlock.L2BlockHeight+1,
			lastHandledBlock.L2BlockHeight+int64(s.Config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("unable to get sender new blocks: %v", err)
			return err
		}
		pendingCommitBlocks, err = ConvertBlocksForCommitToCommitBlockInfos(blocks)
		if err != nil {
			logx.Errorf("unable to convert blocks to commit block infos: %v", err)
			return err
		}
		// get last block info
		lastHandledBlockInfo, err := s.blockModel.GetBlockByBlockHeight(lastHandledBlock.L2BlockHeight)
		if err != nil && err != errorcode.DbErrNotFound {
			logx.Errorf("unable to get last handled block info: %v", err)
			return err
		}
		// construct last stored block header
		lastStoredBlockInfo = util.ConstructStoredBlockInfo(lastHandledBlockInfo)
	}
	gasPrice, err := s.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}
	// commit blocks on-chain
	if len(pendingCommitBlocks) != 0 {
		txHash, err := zkbas.CommitBlocks(
			cli, authCli,
			zkbasInstance,
			lastStoredBlockInfo,
			pendingCommitBlocks,
			gasPrice,
			s.Config.ChainConfig.GasLimit)
		if err != nil {
			logx.Errorf("unable to commit blocks: %v", err)
			return err
		}
		for _, pendingCommittedBlock := range pendingCommitBlocks {
			logx.Infof("commit blocks: %v", pendingCommittedBlock.BlockNumber)
		}
		newRollupTx := &l1RollupTx.L1RollupTx{
			L1TxHash:      txHash,
			TxStatus:      PendingStatus,
			TxType:        CommitTxType,
			L2BlockHeight: int64(pendingCommitBlocks[len(pendingCommitBlocks)-1].BlockNumber),
		}
		isValid, err := s.l1RollupTxModel.CreateL1RollupTx(newRollupTx)
		if err != nil {
			logx.Errorf("unable to create l1 tx sender")
			return err
		}
		if !isValid {
			logx.Errorf("cannot create new senders")
			return errors.New("cannot create new senders")
		}
		logx.Infof("new blocks have been committed(height): %v", newRollupTx.L2BlockHeight)
		return nil
	}
	return nil
}

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
	for _, bProof := range blockProofs {
		var proofInfo *util.FormattedProof
		err = json.Unmarshal([]byte(bProof.ProofInfo), &proofInfo)
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
