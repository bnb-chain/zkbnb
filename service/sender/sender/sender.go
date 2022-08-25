/*
 * Copyright © 2021 ZkBAS Protocol
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
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/common/prove"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/compressedblock"
	"github.com/bnb-chain/zkbas/dao/l1rolluptx"
	"github.com/bnb-chain/zkbas/dao/proof"
	"github.com/bnb-chain/zkbas/dao/sysconfig"
	sconfig "github.com/bnb-chain/zkbas/service/sender/config"
	"github.com/bnb-chain/zkbas/types"
)

type Sender struct {
	config sconfig.Config

	// Client
	cli           *_rpc.ProviderClient
	authCli       *_rpc.AuthClient
	zkbasInstance *zkbas.Zkbas

	// Data access objects
	blockModel           block.BlockModel
	compressedBlockModel compressedblock.CompressedBlockModel
	l1RollupTxModel      l1rolluptx.L1RollupTxModel
	sysConfigModel       sysconfig.SysConfigModel
	proofModel           proof.ProofModel
}

func NewSender(c sconfig.Config) *Sender {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %v", err)
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)

	s := &Sender{
		config:               c,
		blockModel:           block.NewBlockModel(conn, c.CacheRedis, gormPointer),
		compressedBlockModel: compressedblock.NewCompressedBlockModel(conn, c.CacheRedis, gormPointer),
		l1RollupTxModel:      l1rolluptx.NewL1RollupTxModel(conn, c.CacheRedis, gormPointer),
		sysConfigModel:       sysconfig.NewSysConfigModel(conn, c.CacheRedis, gormPointer),
		proofModel:           proof.NewProofModel(gormPointer),
	}

	l1RPCEndpoint, err := s.sysConfigModel.GetSysConfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("fatal error, cannot fetch l1RPCEndpoint from sysconfig, err: %v, SysConfigName: %s",
			err, c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	rollupAddress, err := s.sysConfigModel.GetSysConfigByName(types.ZkbasContract)
	if err != nil {
		logx.Severef("fatal error, cannot fetch rollupAddress from sysconfig, err: %v, SysConfigName: %s",
			err, types.ZkbasContract)
		panic(err)
	}

	s.cli, err = _rpc.NewClient(l1RPCEndpoint.Value)
	if err != nil {
		panic(err)
	}
	chainId, err := s.cli.ChainID(context.Background())
	if err != nil {
		panic(err)
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
	// scan rollup tx table for handled committed height
	lastHandledTx, handledErr := s.l1RollupTxModel.GetLatestHandledTx(l1rolluptx.TxTypeCommit)
	if handledErr != nil && handledErr != types.DbErrNotFound {
		logx.Errorf("GetLatestHandledTx err: %v", handledErr)
		return handledErr
	}
	// scan rollup tx table for pending committed height that higher than the latest handled height
	latestPendingTx, pendingErr := s.l1RollupTxModel.GetLatestPendingTx(l1rolluptx.TxTypeCommit)
	if pendingErr != nil && pendingErr != types.DbErrNotFound {
		logx.Errorf("GetLatestPendingTx err: %v", pendingErr)
		return pendingErr
	}

	// case 1:
	if handledErr == types.DbErrNotFound && pendingErr == nil {
		_, isPending, err := cli.GetTransactionByHash(latestPendingTx.L1TxHash)
		if err != nil {
			// if we cannot get it from rpc and the time over 1 min
			lastUpdatedAt := latestPendingTx.UpdatedAt
			if time.Now().After(lastUpdatedAt.Add(time.Duration(s.config.ChainConfig.MaxWaitingTime) * time.Second)) {
				err := s.l1RollupTxModel.DeleteL1RollupTx(latestPendingTx)
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
			logx.Infof("tx is still pending, no need to work for anything tx hash: %s", latestPendingTx.L1TxHash)
			return nil
		} else {
			receipt, err := cli.GetTransactionReceipt(latestPendingTx.L1TxHash)
			if err != nil {
				logx.Errorf("unable to get transaction receipt: %v", err)
				return err
			}
			if receipt.Status == 0 {
				logx.Infof("the transaction is failure, please check: %s", latestPendingTx.L1TxHash)
				return nil
			}
		}
	}
	// case 2:
	if handledErr == nil && pendingErr == nil {
		isSuccess, err := cli.WaitingTransactionStatus(latestPendingTx.L1TxHash)
		// if err != nil, means we cannot get this tx by hash
		if err != nil {
			// if we cannot get it from rpc and the time over 1 min
			lastUpdatedAt := latestPendingTx.UpdatedAt
			if time.Now().After(lastUpdatedAt.Add(time.Duration(s.config.ChainConfig.MaxWaitingTime) * time.Second)) {
				// drop the record
				err := s.l1RollupTxModel.DeleteL1RollupTx(latestPendingTx)
				if err != nil {
					logx.Errorf("unable to delete l1 tx sender: %v", err)
					return err
				}
				return nil
			} else {
				logx.Infof("tx cannot be found, but not exceed time limit: %s", latestPendingTx.L1TxHash)
				return nil
			}
		}
		// if it is pending, still waiting
		if !isSuccess {
			logx.Infof("tx is still pending, no need to work for anything tx hash: %s", latestPendingTx.L1TxHash)
			return nil
		}
	}

	// case 3:
	var lastStoredBlockInfo zkbas.StorageStoredBlockInfo
	var pendingCommitBlocks []zkbas.OldZkbasCommitBlockInfo
	// if lastHandledTx == nil, means we haven't committed any blocks, just start from 0
	// if errorcode.DbErrNotFound, means we haven't committed new blocks, just start to commit
	if handledErr == types.DbErrNotFound && pendingErr == types.DbErrNotFound {
		var blocks []*compressedblock.CompressedBlock
		blocks, err = s.compressedBlockModel.GetCompressedBlockBetween(1, int64(s.config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("GetCompressedBlockBetween err: %v, maxBlockCount: %d",
				err, s.config.ChainConfig.MaxBlockCount)
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
	if handledErr == nil && pendingErr == types.DbErrNotFound {
		// if errorcode.DbErrNotFound, means we haven't committed new blocks, just start to commit
		// get blocks higher than last handled blocks
		var blocks []*compressedblock.CompressedBlock
		// commit new blocks
		blocks, err := s.compressedBlockModel.GetCompressedBlockBetween(lastHandledTx.L2BlockHeight+1,
			lastHandledTx.L2BlockHeight+int64(s.config.ChainConfig.MaxBlockCount))
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
		lastHandledBlockInfo, err := s.blockModel.GetBlockByHeight(lastHandledTx.L2BlockHeight)
		if err != nil && err != types.DbErrNotFound {
			logx.Errorf("unable to get last handled block info: %v", err)
			return err
		}
		// construct last stored block header
		lastStoredBlockInfo = chain.ConstructStoredBlockInfo(lastHandledBlockInfo)
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
			s.config.ChainConfig.GasLimit)
		if err != nil {
			logx.Errorf("unable to commit blocks: %v", err)
			return err
		}
		for _, pendingCommittedBlock := range pendingCommitBlocks {
			logx.Infof("commit blocks: %v", pendingCommittedBlock.BlockNumber)
		}
		newRollupTx := &l1rolluptx.L1RollupTx{
			L1TxHash:      txHash,
			TxStatus:      l1rolluptx.StatusPending,
			TxType:        l1rolluptx.TxTypeCommit,
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
	pendingTxs, err := s.l1RollupTxModel.GetL1RollupTxsByStatus(l1rolluptx.StatusPending)
	if err != nil {
		logx.Errorf("unable to get l1 tx senders by tx status: %v", err)
		return err
	}

	latestL1Height, err := s.cli.GetHeight()
	if err != nil {
		logx.Errorf("Get L1 height err: %v", err)
		return err
	}

	var (
		pendingUpdateRxs         []*l1rolluptx.L1RollupTx
		pendingUpdateProofStatus = make(map[int64]int)
	)
	for _, pendingTx := range pendingTxs {
		txHash := pendingTx.L1TxHash
		receipt, err := s.cli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("GetTransactionReceipt %s err: %v", txHash, err)
			continue
		}

		// not finalized yet
		if latestL1Height < receipt.BlockNumber.Uint64()+s.config.ChainConfig.ConfirmBlocksCount {
			continue
		}
		var validTx bool
		for _, vlog := range receipt.Logs {
			switch vlog.Topics[0].Hex() {
			case zkbasLogBlockCommitSigHash.Hex():
				var event zkbas.ZkbasBlockCommit
				if err = ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockCommit, vlog.Data); err != nil {
					logx.Errorf("UnpackIntoInterface err: %v", err)
					return err
				}
				validTx = int64(event.BlockNumber) == pendingTx.L2BlockHeight
			case zkbasLogBlockVerificationSigHash.Hex():
				var event zkbas.ZkbasBlockVerification
				if err = ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
					logx.Errorf("UnpackIntoInterface err: %v", err)
					return err
				}
				validTx = int64(event.BlockNumber) == pendingTx.L2BlockHeight
				pendingUpdateProofStatus[pendingTx.L2BlockHeight] = proof.Confirmed
			case zkbasLogBlocksRevertSigHash.Hex():
				// TODO revert
			default:
			}
		}

		if validTx {
			pendingTx.TxStatus = l1rolluptx.StatusHandled
			pendingUpdateRxs = append(pendingUpdateRxs, pendingTx)
		}
	}

	if err = s.l1RollupTxModel.UpdateL1RollupTxs(pendingUpdateRxs,
		pendingUpdateProofStatus); err != nil {
		logx.Errorf("update sent txs error, err: %v", err)
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
	lastHandledBlock, getHandleErr := s.l1RollupTxModel.GetLatestHandledTx(l1rolluptx.TxTypeVerifyAndExecute)
	if getHandleErr != nil && getHandleErr != types.DbErrNotFound {
		logx.Errorf("unable to get latest handled block: %v", getHandleErr)
		return getHandleErr
	}
	// scan l1 tx sender table for pending verified and executed height that higher than the latest handled height
	pendingSender, getPendingErr := s.l1RollupTxModel.GetLatestPendingTx(l1rolluptx.TxTypeVerifyAndExecute)
	if getPendingErr != nil && getPendingErr != types.DbErrNotFound {
		logx.Errorf("unable to get latest pending blocks: %v", getPendingErr)
		return getPendingErr
	}
	// case 1: check tx status on L1
	if getHandleErr == types.DbErrNotFound && getPendingErr == nil {
		_, isPending, err := cli.GetTransactionByHash(pendingSender.L1TxHash)
		// if err != nil, means we cannot get this tx by hash
		if err != nil {
			// if we cannot get it from rpc and the time over 1 min
			lastUpdatedAt := pendingSender.UpdatedAt
			if time.Now().After(lastUpdatedAt.Add(time.Duration(s.config.ChainConfig.MaxWaitingTime) * time.Second)) {
				// drop the record
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
			lastUpdatedAt := pendingSender.UpdatedAt
			if time.Now().After(lastUpdatedAt.Add(time.Duration(s.config.ChainConfig.MaxWaitingTime) * time.Second)) {
				// drop the record
				if err := s.l1RollupTxModel.DeleteL1RollupTx(pendingSender); err != nil {
					logx.Errorf("unable to delete l1 tx sender: %v", err)
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
	if getHandleErr == types.DbErrNotFound && getPendingErr == types.DbErrNotFound {
		// get blocks from block table
		blocks, err = s.blockModel.GetBlocksForProverBetween(1, int64(s.config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("GetBlocksForProverBetween err: %v, maxBlockCount: %d",
				err, s.config.ChainConfig.MaxBlockCount)
			return err
		}
		pendingVerifyAndExecuteBlocks, err = ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
		if err != nil {
			logx.Errorf("unable to convert blocks to verify and execute block infos: %v", err)
			return err
		}
		start = int64(1)
	}
	if getHandleErr == nil && getPendingErr == types.DbErrNotFound {
		blocks, err = s.blockModel.GetBlocksForProverBetween(lastHandledBlock.L2BlockHeight+1,
			lastHandledBlock.L2BlockHeight+int64(s.config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("unable to get sender new blocks: %v", err)
			return err
		}
		pendingVerifyAndExecuteBlocks, err = ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
		if err != nil {
			logx.Errorf("unable to convert blocks to commit block infos: %v", err)
			return err
		}
		start = lastHandledBlock.L2BlockHeight + 1
	}

	blockProofs, err := s.proofModel.GetProofsByBlockRange(start, blocks[len(blocks)-1].BlockHeight,
		s.config.ChainConfig.MaxBlockCount)
	if err != nil {
		logx.Errorf("unable to get proofs: %v", err)
		return err
	}
	if len(blockProofs) != len(blocks) {
		logx.Errorf("we haven't generated related proofs, please wait")
		return errors.New("we haven't generated related proofs, please wait")
	}
	var proofs []*big.Int
	for _, bProof := range blockProofs {
		var proofInfo *prove.FormattedProof
		err = json.Unmarshal([]byte(bProof.ProofInfo), &proofInfo)
		if err != nil {
			logx.Errorf("unable to unmarshal proof info: %v", err)
			return err
		}
		proofs = append(proofs, proofInfo.A[:]...)
		proofs = append(proofs, proofInfo.B[0][0], proofInfo.B[0][1])
		proofs = append(proofs, proofInfo.B[1][0], proofInfo.B[1][1])
		proofs = append(proofs, proofInfo.C[:]...)
	}
	gasPrice, err := s.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price err: %v", err)
		return err
	}
	// commit blocks on-chain
	if len(pendingVerifyAndExecuteBlocks) != 0 {
		txHash, err := zkbas.VerifyAndExecuteBlocks(cli, authCli, zkbasInstance,
			pendingVerifyAndExecuteBlocks, proofs, gasPrice, s.config.ChainConfig.GasLimit)
		if err != nil {
			logx.Errorf("VerifyAndExecuteBlocks err: %v", err)
			return err
		}

		newRollupTx := &l1rolluptx.L1RollupTx{
			L1TxHash:      txHash,
			TxStatus:      l1rolluptx.StatusPending,
			TxType:        l1rolluptx.TxTypeVerifyAndExecute,
			L2BlockHeight: int64(pendingVerifyAndExecuteBlocks[len(pendingVerifyAndExecuteBlocks)-1].BlockHeader.BlockNumber),
		}
		isValid, err := s.l1RollupTxModel.CreateL1RollupTx(newRollupTx)
		if err != nil {
			logx.Errorf("CreateL1TxSender err: %v", err)
			return err
		}
		if !isValid {
			logx.Errorf("cannot create new senders")
			return errors.New("cannot create new senders")
		}
		logx.Errorf("new blocks have been verified and executed(height): %d", newRollupTx.L2BlockHeight)
		return nil
	}
	return nil
}
