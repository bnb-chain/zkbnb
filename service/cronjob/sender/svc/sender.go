package svc

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/l1TxSender"
	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/config"
)

const (
	PendingStatus          = l1TxSender.PendingStatus
	CommitTxType           = l1TxSender.CommitTxType
	VerifyAndExecuteTxType = l1TxSender.VerifyAndExecuteTxType
)

type Sender struct {
	Config config.Config

	// Client
	cli           *ProviderClient
	authCli       *AuthClient
	zkbasInstance *Zkbas

	// Data access objects
	blockModel          block.BlockModel
	blockForCommitModel blockForCommit.BlockForCommitModel
	l1TxSenderModel     l1TxSender.L1TxSenderModel
	sysConfigModel      sysconfig.SysconfigModel
	proofSenderModel    proofSender.ProofSenderModel
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func NewSender(c config.Config) *Sender {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %v", err)
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))

	s := &Sender{
		Config:              c,
		blockModel:          block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		blockForCommitModel: blockForCommit.NewBlockForCommitModel(conn, c.CacheRedis, gormPointer),
		l1TxSenderModel:     l1TxSender.NewL1TxSenderModel(conn, c.CacheRedis, gormPointer),
		sysConfigModel:      sysconfig.NewSysconfigModel(conn, c.CacheRedis, gormPointer),
		proofSenderModel:    proofSender.NewProofSenderModel(gormPointer),
	}

	l1RPCEndpoint, err := s.sysConfigModel.GetSysconfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch l1RPCEndpoint from sysConfig, err: %v, SysConfigName: %s",
			err, c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	rollupAddress, err := s.sysConfigModel.GetSysconfigByName(c.ChainConfig.ZkbasContractAddrSysConfigName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch rollupAddress from sysConfig, err: %v, SysConfigName: %s",
			err, c.ChainConfig.ZkbasContractAddrSysConfigName)
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

func (s *Sender) CommittedBlocks() (err error) {
	var (
		cli           = s.cli
		authCli       = s.authCli
		zkbasInstance = s.zkbasInstance
	)
	// scan l1 tx sender table for handled committed height
	lastHandledBlock, getHandleErr := s.l1TxSenderModel.GetLatestHandledBlock(CommitTxType)
	if getHandleErr != nil && getHandleErr != errorcode.DbErrNotFound {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] GetLatestHandledBlock err: %v", getHandleErr)
		return getHandleErr
	}
	// scan l1 tx sender table for pending committed height that higher than the latest handled height
	pendingSender, getPendingErr := s.l1TxSenderModel.GetLatestPendingBlock(CommitTxType)
	if getPendingErr != nil {
		if getPendingErr != errorcode.DbErrNotFound {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] GetLatestPendingBlock err: %v", getPendingErr)
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
				err := s.l1TxSenderModel.DeleteL1TxSender(pendingSender)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to delete l1 tx sender: %v", err)
					return err
				}
				return nil
			} else {
				return nil
			}
		}
		// if it is pending, still waiting
		if isPending {
			logx.Infof("[SendCommittedBlocks] tx is still pending, no need to work for anything tx hash: %s", pendingSender.L1TxHash)
			return nil
		} else {
			receipt, err := cli.GetTransactionReceipt(pendingSender.L1TxHash)
			if err != nil {
				logx.Errorf("[SendCommittedBlocks] unable to get transaction receipt: %v", err)
				return err
			}
			if receipt.Status == 0 {
				logx.Infof("[SendCommittedBlocks] the transaction is failure, please check: %s", pendingSender.L1TxHash)
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
				err := s.l1TxSenderModel.DeleteL1TxSender(pendingSender)
				if err != nil {
					logx.Errorf("[SendCommittedBlocks] unable to delete l1 tx sender: %v", err)
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

	// case 3:
	var lastStoredBlockInfo StorageStoredBlockInfo
	var pendingCommitBlocks []ZkbasCommitBlockInfo
	// if lastHandledBlock == nil, means we haven't committed any blocks, just start from 0
	// if errorcode.DbErrNotFound, means we haven't committed new blocks, just start to commit
	if getHandleErr == errorcode.DbErrNotFound && getPendingErr == errorcode.DbErrNotFound {
		var blocks []*BlockForCommit
		blocks, err = s.blockForCommitModel.GetBlockForCommitBetween(1, int64(s.Config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] GetBlockForCommitBetween err: %v, maxBlockCount: %d",
				err, s.Config.ChainConfig.MaxBlockCount)
			return err
		}
		pendingCommitBlocks, err = ConvertBlocksForCommitToCommitBlockInfos(blocks)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to convert blocks to commit block infos: %vv", err)
			return err
		}
		// set stored block header to default 0
		lastStoredBlockInfo = DefaultBlockHeader()
	}
	if getHandleErr == nil && getPendingErr == errorcode.DbErrNotFound {
		// if errorcode.DbErrNotFound, means we haven't committed new blocks, just start to commit
		// get blocks higher than last handled blocks
		var blocks []*BlockForCommit
		// commit new blocks
		blocks, err = s.blockForCommitModel.GetBlockForCommitBetween(lastHandledBlock.L2BlockHeight+1,
			lastHandledBlock.L2BlockHeight+int64(s.Config.ChainConfig.MaxBlockCount))
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to get sender new blocks: %v", err)
			return err
		}
		pendingCommitBlocks, err = ConvertBlocksForCommitToCommitBlockInfos(blocks)
		if err != nil {
			logx.Errorf("[SendCommittedBlocks] unable to convert blocks to commit block infos: %v", err)
			return err
		}
		// get last block info
		lastHandledBlockInfo, err := s.blockModel.GetBlockByBlockHeight(lastHandledBlock.L2BlockHeight)
		if err != nil && err != errorcode.DbErrNotFound {
			logx.Errorf("[SendCommittedBlocks] unable to get last handled block info: %v", err)
			return err
		}
		// construct last stored block header
		lastStoredBlockInfo = util.ConstructStoredBlockInfo(lastHandledBlockInfo)
	}
	gasPrice, err := s.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("[SendCommittedBlocks] failed to fetch gas price: %v", err)
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
			logx.Errorf("[SendCommittedBlocks] unable to commit blocks: %v", err)
			return err
		}
		for _, pendingCommittedBlock := range pendingCommitBlocks {
			logx.Infof("[SendCommittedBlocks] commit blocks: %v", pendingCommittedBlock.BlockNumber)
		}
		// update l1 tx sender table records
		newSender := &L1TxSender{
			L1TxHash:      txHash,
			TxStatus:      PendingStatus,
			TxType:        CommitTxType,
			L2BlockHeight: int64(pendingCommitBlocks[len(pendingCommitBlocks)-1].BlockNumber),
		}
		isValid, err := s.l1TxSenderModel.CreateL1TxSender(newSender)
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
	lastHandledBlock, getHandleErr := s.l1TxSenderModel.GetLatestHandledBlock(VerifyAndExecuteTxType)
	if getHandleErr != nil && getHandleErr != errorcode.DbErrNotFound {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get latest handled block: %v", getHandleErr)
		return getHandleErr
	}
	// scan l1 tx sender table for pending verified and executed height that higher than the latest handled height
	pendingSender, getPendingErr := s.l1TxSenderModel.GetLatestPendingBlock(VerifyAndExecuteTxType)
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
				err := s.l1TxSenderModel.DeleteL1TxSender(pendingSender)
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
				if err := s.l1TxSenderModel.DeleteL1TxSender(pendingSender); err != nil {
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
		pendingVerifyAndExecuteBlocks []ZkbasVerifyBlockInfo
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
	// TODO: for test
	/*
		if len(blocks) < maxBlockCount {
			logx.Errorf("current pending verify blocks %d is less than %d", len(blocks), maxBlockCount)
			return err
		}
	*/
	proofSenders, err := s.proofSenderModel.GetProofsByBlockRange(start, blocks[len(blocks)-1].BlockHeight,
		s.Config.ChainConfig.MaxBlockCount)
	if err != nil {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] unable to get proofs: %v", err)
		return err
	}
	if len(proofSenders) != len(blocks) {
		logx.Errorf("[SendVerifiedAndExecutedBlocks] we haven't generated related proofs, please wait")
		return errors.New("[SendVerifiedAndExecutedBlocks] we haven't generated related proofs, please wait")
	}
	var proofs []*big.Int
	for _, proof := range proofSenders {
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
		// update l1 tx sender table records
		newSender := &L1TxSender{
			L1TxHash:      txHash,
			TxStatus:      PendingStatus,
			TxType:        VerifyAndExecuteTxType,
			L2BlockHeight: int64(pendingVerifyAndExecuteBlocks[len(pendingVerifyAndExecuteBlocks)-1].BlockHeader.BlockNumber),
		}
		isValid, err := s.l1TxSenderModel.CreateL1TxSender(newSender)
		if err != nil {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] CreateL1TxSender err: %v", err)
			return err
		}
		if !isValid {
			logx.Errorf("[SendVerifiedAndExecutedBlocks] cannot create new senders")
			return errors.New("[SendVerifiedAndExecutedBlocks] cannot create new senders")
		}
		logx.Errorf("[SendVerifiedAndExecutedBlocks] new blocks have been verified and executed(height): %d", newSender.L2BlockHeight)
		return nil
	}
	return nil
}
