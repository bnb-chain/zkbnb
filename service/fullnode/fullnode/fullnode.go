package fullnode

import (
	"fmt"
	"time"

	"github.com/bnb-chain/zkbnb-go-sdk/client"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/block"
	tx "github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	DefaultL2EndPoint = "http://localhost:8888"
	SyncInterval      = 100 * time.Millisecond
)

type Config struct {
	core.ChainConfig
	L2EndPoint      string
	SyncBlockStatus int64
	LogConf         logx.LogConf
}

type Fullnode struct {
	config *Config
	client client.ZkBNBClient
	bc     *core.BlockChain

	quitCh chan struct{}
}

func NewFullnode(config *Config) (*Fullnode, error) {
	bc, err := core.NewBlockChain(&config.ChainConfig, "fullnode")
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	l2EndPoint := config.L2EndPoint
	if len(l2EndPoint) == 0 {
		l2EndPoint = DefaultL2EndPoint
	}

	if config.SyncBlockStatus <= block.StatusProposing ||
		config.SyncBlockStatus > block.StatusVerifiedAndExecuted {
		config.SyncBlockStatus = block.StatusVerifiedAndExecuted
	}

	fullnode := &Fullnode{
		config: config,
		client: nil,
		bc:     bc,

		quitCh: make(chan struct{}),
	}
	return fullnode, nil
}

func (c *Fullnode) Run() {
	curHeight, err := c.bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		logx.Severef("get current block height failed, error: %s", err.Error())
		panic(fmt.Sprintf("get current block height failed, error: %v", err.Error()))
	}

	curBlock, err := c.bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		logx.Severef("get current block failed, height: %d, error: %s", curHeight, err.Error())
		panic(fmt.Sprintf("get current block failed, height: %d, error: %v", curHeight, err.Error()))
	}

	ticker := time.NewTicker(SyncInterval)
	defer ticker.Stop()

	syncBlockStatus := c.config.SyncBlockStatus
	for {
		select {
		case <-ticker.C:

			// if the latest block have been created
			if curBlock.BlockStatus > block.StatusProposing {
				// init new block, set curBlock.status to block.StatusProposing
				curBlock, err = c.bc.InitNewBlock()
				if err != nil {
					logx.Severef("init new block failed, block height: %d, error: %s", curHeight, err.Error())
					panic(fmt.Sprintf("init new block failed, block height: %d, error: %v", curHeight, err.Error()))
				}

				curHeight++
			}

			l2Block, err := c.client.GetBlockByHeight(curHeight)
			if err != nil {
				if err != types.DbErrNotFound {
					logx.Errorf("get block failed, height: %d, err %v ", curHeight, err)
				}
				continue
			}

			if l2Block.Status < syncBlockStatus {
				continue
			}

			// create time needs to be set, otherwise tx will fail if expire time is set
			c.bc.CurrentBlock().CreatedAt = time.UnixMilli(l2Block.CommittedAt)
			// set info
			if l2Block.Status >= block.StatusCommitted {
				c.bc.CurrentBlock().CommittedAt = l2Block.CommittedAt
				c.bc.CurrentBlock().CommittedTxHash = l2Block.CommittedTxHash
				c.bc.CurrentBlock().BlockCommitment = l2Block.Commitment
			}
			if l2Block.Status == block.StatusVerifiedAndExecuted {
				c.bc.CurrentBlock().VerifiedAt = l2Block.VerifiedAt
				c.bc.CurrentBlock().VerifiedTxHash = l2Block.VerifiedTxHash
			}

			// clean cache
			c.bc.Statedb.PurgeCache(curBlock.StateRoot)

			for _, blockTx := range l2Block.Txs {
				newTx := &tx.Tx{BaseTx: tx.BaseTx{
					TxHash: blockTx.Hash, // Would be computed in prepare method of executors.
					TxType: blockTx.Type,
					TxInfo: blockTx.Info,
				}}

				err = c.bc.ApplyTransaction(newTx)
				if err != nil {
					logx.Errorf("apply block tx ID: %d failed, err %v ", newTx.ID, err)
					continue
				}
			}

			err = c.bc.Statedb.UpdateAssetTree(true, nil)
			if err != nil {
				logx.Severef("calculate state root failed, err: %v", err)
				panic(fmt.Sprint("calculate state root failed, err", err))
			}

			if c.bc.Statedb.StateRoot != l2Block.StateRoot {
				logx.Severef("state root not matched between statedb and l2block: %d, local: %s, remote: %s", l2Block.Height, c.bc.Statedb.StateRoot, l2Block.StateRoot)
				panic(fmt.Sprintf("state root not matched between statedb and l2block: %d, local: %s, remote: %s", l2Block.Height, c.bc.Statedb.StateRoot, l2Block.StateRoot))
			}

			curBlock, err = c.processNewBlock(int(l2Block.Size))
			if err != nil {
				logx.Severef("new block failed, block height: %d, Error: %s", l2Block.Height, err.Error())
				panic(fmt.Sprintf("new block failed, block height: %d, Error: %s", l2Block.Height, err.Error()))
			}
			logx.Infof("created new block on fullnode, height=%d, blockSize=%d", curBlock.BlockHeight, l2Block.Size)
		case <-c.quitCh:
			return
		}
	}
}

func (c *Fullnode) Shutdown() {
	close(c.quitCh)
	c.bc.Statedb.Close()
	c.bc.ChainDB.Close()
}

func (c *Fullnode) processNewBlock(blockSize int) (*block.Block, error) {
	//blockStates, err := c.bc.CommitNewBlock(blockSize, nil)
	//if err != nil {
	//	return nil, err

	//}
	//blockStates.Block.BlockStatus = c.config.SyncBlockStatus
	//
	//// sync gas account
	//err = c.bc.Statedb.SyncGasAccountToRedis()
	//if err != nil {
	//	return nil, err
	//}
	//
	//// sync pending value to caches
	//err = c.bc.Statedb.SyncStateCacheToRedis()
	//if err != nil {
	//	panic("sync redis cache failed: " + err.Error())
	//}
	//
	//// update db
	//err = c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
	//	// create block for commit
	//	if blockStates.CompressedBlock != nil {
	//		err = c.bc.DB().CompressedBlockModel.CreateCompressedBlockInTransact(tx, blockStates.CompressedBlock)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	// create or update account
	//	if len(blockStates.PendingAccount) != 0 {
	//		err = c.bc.DB().AccountModel.UpdateAccountsInTransact(tx, blockStates.PendingAccount)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	// create account history
	//	if len(blockStates.PendingAccountHistory) != 0 {
	//		err = c.bc.DB().AccountHistoryModel.CreateAccountHistoriesInTransact(tx, blockStates.PendingAccountHistory)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	// create or update nft
	//	if len(blockStates.PendingNft) != 0 {
	//		err = c.bc.DB().L2NftModel.UpdateNftsInTransact(tx, blockStates.PendingNft)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	// create nft history
	//	if len(blockStates.PendingNftHistory) != 0 {
	//		err = c.bc.DB().L2NftHistoryModel.CreateNftHistoriesInTransact(tx, blockStates.PendingNftHistory)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//
	//	return c.bc.DB().BlockModel.CreateBlockInTransact(tx, blockStates.Block)
	//})
	//
	//if err != nil {
	//	return nil, err
	//}
	//
	//return blockStates.Block, nil
	return nil, nil
}
