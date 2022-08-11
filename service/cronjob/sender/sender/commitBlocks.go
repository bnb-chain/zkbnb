package sender

import (
	"context"
	"errors"
	"time"

	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/l1RollupTx"
	"github.com/bnb-chain/zkbas/common/util"
)

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
