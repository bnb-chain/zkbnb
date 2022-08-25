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
	"fmt"
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
	pendingTx, err := s.l1RollupTxModel.GetLatestPendingTx(l1rolluptx.TxTypeCommit)
	if err != nil && err != types.DbErrNotFound {
		return err
	}
	// No need to submit new transaction if there is any pending commit txs.
	if pendingTx != nil {
		return nil
	}

	lastHandledTx, err := s.l1RollupTxModel.GetLatestHandledTx(l1rolluptx.TxTypeCommit)
	if err != nil && err != types.DbErrNotFound {
		return err
	}
	start := int64(1)
	if lastHandledTx != nil {
		start = lastHandledTx.L2BlockHeight + 1
	}
	// commit new blocks
	blocks, err := s.compressedBlockModel.GetCompressedBlockBetween(start,
		start+int64(s.config.ChainConfig.MaxBlockCount))
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("failed to get compress block err: %v", err)
	}
	if len(blocks) == 0 {
		return nil
	}
	pendingCommitBlocks, err := ConvertBlocksForCommitToCommitBlockInfos(blocks)
	if err != nil {
		return fmt.Errorf("failed to get commit block info, err: %v", err)
	}
	// get last block info
	lastStoredBlockInfo := defaultBlockHeader()
	if lastHandledTx != nil {
		lastHandledBlockInfo, err := s.blockModel.GetBlockByHeight(lastHandledTx.L2BlockHeight)
		if err != nil {
			return fmt.Errorf("failed to get block info, err: %v", err)
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
	txHash, err := zkbas.CommitBlocks(
		cli, authCli,
		zkbasInstance,
		lastStoredBlockInfo,
		pendingCommitBlocks,
		gasPrice,
		s.config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send commit tx, errL %v", err)
	}
	newRollupTx := &l1rolluptx.L1RollupTx{
		L1TxHash:      txHash,
		TxStatus:      l1rolluptx.StatusPending,
		TxType:        l1rolluptx.TxTypeCommit,
		L2BlockHeight: int64(pendingCommitBlocks[len(pendingCommitBlocks)-1].BlockNumber),
	}
	err = s.l1RollupTxModel.CreateL1RollupTx(newRollupTx)
	if err != nil {
		return fmt.Errorf("failed to create tx in database, err: %v", err)
	}
	logx.Infof("new blocks have been committed(height): %v", newRollupTx.L2BlockHeight)
	return nil
}

func (s *Sender) UpdateSentTxs() (err error) {
	pendingTxs, err := s.l1RollupTxModel.GetL1RollupTxsByStatus(l1rolluptx.StatusPending)
	if err != nil {
		if err == types.DbErrNotFound {
			return nil
		}
		return fmt.Errorf("failed to get pending txs, err: %v", err)
	}

	latestL1Height, err := s.cli.GetHeight()
	if err != nil {
		return fmt.Errorf("failed to get l1 block height, err: %v", err)
	}

	var (
		pendingUpdateRxs         []*l1rolluptx.L1RollupTx
		pendingUpdateProofStatus = make(map[int64]int)
	)
	for _, pendingTx := range pendingTxs {
		txHash := pendingTx.L1TxHash
		receipt, err := s.cli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("query transaction receipt %s failed, err: %v", txHash, err)
			if time.Now().After(pendingTx.UpdatedAt.Add(time.Duration(s.config.ChainConfig.MaxWaitingTime) * time.Second)) {
				// No need to check the response, do best effort.
				s.l1RollupTxModel.DeleteL1RollupTx(pendingTx)
			}
			continue
		}
		if receipt.Status == 0 {
			// It is critical to have any failed transactions
			panic(fmt.Sprintf("unexpected failed tx: %v", txHash))
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
					return err
				}
				validTx = int64(event.BlockNumber) == pendingTx.L2BlockHeight
			case zkbasLogBlockVerificationSigHash.Hex():
				var event zkbas.ZkbasBlockVerification
				if err = ZkbasContractAbi.UnpackIntoInterface(&event, EventNameBlockVerification, vlog.Data); err != nil {
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
		return fmt.Errorf("failed to updte rollup txs, err:%v", err)
	}
	return nil
}

func (s *Sender) VerifyAndExecuteBlocks() (err error) {
	var (
		cli           = s.cli
		authCli       = s.authCli
		zkbasInstance = s.zkbasInstance
	)
	pendingTx, err := s.l1RollupTxModel.GetLatestPendingTx(l1rolluptx.TxTypeVerifyAndExecute)
	if err != nil && err != types.DbErrNotFound {
		return err
	}
	// No need to submit new transaction if there is any pending verification txs.
	if pendingTx != nil {
		return nil
	}

	lastHandledTx, err := s.l1RollupTxModel.GetLatestHandledTx(l1rolluptx.TxTypeVerifyAndExecute)
	if err != nil && err != types.DbErrNotFound {
		return err
	}

	start := int64(1)
	if lastHandledTx != nil {
		start = lastHandledTx.L2BlockHeight + 1
	}
	blocks, err := s.blockModel.GetCommittedBlocksBetween(start,
		start+int64(s.config.ChainConfig.MaxBlockCount))
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("unable to get blocks to prove, err: %v", err)
	}
	if len(blocks) == 0 {
		return nil
	}
	pendingVerifyAndExecuteBlocks, err := ConvertBlocksToVerifyAndExecuteBlockInfos(blocks)
	if err != nil {
		return fmt.Errorf("unable to convert blocks to commit block infos: %v", err)
	}

	blockProofs, err := s.proofModel.GetProofsBetween(start, start+int64(len(blocks))-1)
	if err != nil {
		return fmt.Errorf("unable to get proofs, err: %v", err)
	}
	if len(blockProofs) != len(blocks) {
		return errors.New("related proofs not ready")
	}
	var proofs []*big.Int
	for _, bProof := range blockProofs {
		var proofInfo *prove.FormattedProof
		err = json.Unmarshal([]byte(bProof.ProofInfo), &proofInfo)
		if err != nil {
			return err
		}
		proofs = append(proofs, proofInfo.A[:]...)
		proofs = append(proofs, proofInfo.B[0][0], proofInfo.B[0][1])
		proofs = append(proofs, proofInfo.B[1][0], proofInfo.B[1][1])
		proofs = append(proofs, proofInfo.C[:]...)
	}
	gasPrice, err := s.cli.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	// Verify blocks on-chain
	txHash, err := zkbas.VerifyAndExecuteBlocks(cli, authCli, zkbasInstance,
		pendingVerifyAndExecuteBlocks, proofs, gasPrice, s.config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send verify tx: %v", err)
	}

	newRollupTx := &l1rolluptx.L1RollupTx{
		L1TxHash:      txHash,
		TxStatus:      l1rolluptx.StatusPending,
		TxType:        l1rolluptx.TxTypeVerifyAndExecute,
		L2BlockHeight: int64(pendingVerifyAndExecuteBlocks[len(pendingVerifyAndExecuteBlocks)-1].BlockHeader.BlockNumber),
	}
	err = s.l1RollupTxModel.CreateL1RollupTx(newRollupTx)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to create rollup tx in db %v", err))
	}
	logx.Infof("new blocks have been verified and executed(height): %d", newRollupTx.L2BlockHeight)
	return nil
}
