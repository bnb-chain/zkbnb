package committer

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

const (
	MaxCommitterInterval = 60 * 1
)

var (
	priorityOperationMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "priority_operation_process",
		Help:      "Priority operation requestID metrics.",
	})
	priorityOperationHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "priority_operation_process_height",
		Help:      "Priority operation height metrics.",
	})
	commitOperationMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "db_commit_time",
		Help:      "DB commit operation time",
	})
	executeTxOperationMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_time",
		Help:      "execute txs operation time",
	})
	pendingTxNumMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "pending_tx",
		Help:      "number of pending tx",
	})
	stateDBOperationMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "state_db_time",
		Help:      "stateDB commit operation time",
	})
	stateDBSyncOperationMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "state_sync_time",
		Help:      "stateDB sync operation time",
	})
	sqlDBOperationMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "sql_db_time",
		Help:      "sql DB commit operation time",
	})
	l2BlockHeightMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2_block_height",
		Help:      "l2_Block_Height metrics.",
	})
	poolTxL1ErrorCountMetics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "pool_tx_l1_error_count",
		Help:      "pool_tx_l1_error_count metrics.",
	})
	poolTxL2ErrorCountMetics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "pool_tx_l2_error_count",
		Help:      "pool_tx_l2_error_count metrics.",
	})
)

type Config struct {
	core.ChainConfig

	BlockConfig struct {
		OptionalBlockSizes []int
	}
	LogConf logx.LogConf
}

type Committer struct {
	running            bool
	config             *Config
	maxTxsPerBlock     int
	optionalBlockSizes []int

	bc *core.BlockChain
}

func NewCommitter(config *Config) (*Committer, error) {
	if len(config.BlockConfig.OptionalBlockSizes) == 0 {
		return nil, errors.New("nil optional block sizes")
	}

	bc, err := core.NewBlockChain(&config.ChainConfig, "committer")
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	if err := prometheus.Register(priorityOperationMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register priorityOperationMetric error: %v", err)
	}
	if err := prometheus.Register(priorityOperationHeightMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register priorityOperationHeightMetric error: %v", err)
	}
	if err := prometheus.Register(commitOperationMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register commitOperationMetics error: %v", err)
	}
	if err := prometheus.Register(pendingTxNumMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register pendingTxNumMetrics error: %v", err)
	}
	if err := prometheus.Register(executeTxOperationMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxOperationMetrics error: %v", err)
	}
	if err := prometheus.Register(stateDBOperationMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register stateDBOperationMetics error: %v", err)
	}
	if err := prometheus.Register(stateDBSyncOperationMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register stateDBSyncOperationMetics error: %v", err)
	}
	if err := prometheus.Register(sqlDBOperationMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register sqlDBOperationMetics error: %v", err)
	}
	if err := prometheus.Register(l2BlockHeightMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2BlockHeightMetics error: %v", err)
	}
	if err := prometheus.Register(poolTxL1ErrorCountMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register poolTxL1ErrorCountMetics error: %v", err)
	}
	if err := prometheus.Register(poolTxL2ErrorCountMetics); err != nil {
		return nil, fmt.Errorf("prometheus.Register poolTxL2ErrorCountMetics error: %v", err)
	}

	committer := &Committer{
		running:            true,
		config:             config,
		maxTxsPerBlock:     config.BlockConfig.OptionalBlockSizes[len(config.BlockConfig.OptionalBlockSizes)-1],
		optionalBlockSizes: config.BlockConfig.OptionalBlockSizes,

		bc: bc,
	}
	return committer, nil
}

func (c *Committer) Run() {
	curBlock, err := c.restoreExecutedTxs()
	if err != nil {
		panic("restore executed tx failed: " + err.Error())
	}

	latestRequestId, err := c.getLatestExecutedRequestId()
	if err != nil {
		logx.Error("get latest executed request ID failed:", err)
		latestRequestId = -1
	}

	for {
		if !c.running {
			break
		}
		if curBlock.BlockStatus > block.StatusProposing {
			curBlock, err = c.bc.InitNewBlock()
			if err != nil {
				panic("propose new block failed: " + err.Error())
			}
		}

		// Read pending transactions from tx pool.
		pendingTxs, err := c.bc.TxPoolModel.GetTxsByStatus(tx.StatusPending)
		if err != nil {
			logx.Error("get pending transactions from tx pool failed:", err)
			return
		}
		for len(pendingTxs) == 0 {
			if c.shouldCommit(curBlock) {
				break
			}

			time.Sleep(100 * time.Millisecond)
			pendingTxs, err = c.bc.TxPoolModel.GetTxsByStatus(tx.StatusPending)
			if err != nil {
				logx.Error("get pending transactions from tx pool failed:", err)
				return
			}
		}

		pendingTxNumMetrics.Set(float64(len(pendingTxs)))
		pendingUpdatePoolTxs := make([]*tx.Tx, 0, len(pendingTxs))
		pendingDeletePoolTxs := make([]*tx.Tx, 0, len(pendingTxs))
		start := time.Now()
		for _, poolTx := range pendingTxs {
			if c.shouldCommit(curBlock) {
				break
			}
			logx.Infof("apply transaction, txHash=%s", poolTx.TxHash)
			err = c.bc.ApplyTransaction(poolTx)
			if err != nil {
				logx.Errorf("apply pool tx ID: %d failed, err %v ", poolTx.ID, err)
				poolTx.TxStatus = tx.StatusFailed
				pendingDeletePoolTxs = append(pendingDeletePoolTxs, poolTx)
				if types.IsPriorityOperationTx(poolTx.TxType) {
					poolTxL1ErrorCountMetics.Inc()
				} else {
					poolTxL2ErrorCountMetics.Inc()
				}
				continue
			}

			if types.IsPriorityOperationTx(poolTx.TxType) {
				request, err := c.bc.PriorityRequestModel.GetPriorityRequestsByL2TxHash(poolTx.TxHash)
				if err == nil {

					priorityOperationMetric.Set(float64(request.RequestId))
					priorityOperationHeightMetric.Set(float64(request.L1BlockHeight))

					if latestRequestId != -1 && request.RequestId != latestRequestId+1 {
						logx.Errorf("invalid request ID: %d, txHash: %s", request.RequestId, poolTx.TxHash)
						return
					}
					latestRequestId = request.RequestId
				} else {
					logx.Errorf("query txHash: %s in PriorityRequestTable failed, err %v ", poolTx.TxHash, err)
				}
			}

			// Write the proposed block into database when the first transaction executed.
			if len(c.bc.Statedb.Txs) == 1 {
				err = c.createNewBlock(curBlock, poolTx)
				if err != nil {
					panic("create new block failed" + err.Error())
				}
			} else {
				pendingUpdatePoolTxs = append(pendingUpdatePoolTxs, poolTx)
			}
		}
		executeTxOperationMetrics.Set(float64(time.Since(start).Milliseconds()))

		err = c.bc.StateDB().SyncStateCacheToRedis()
		if err != nil {
			panic("sync redis cache failed: " + err.Error())
		}

		err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
			err := c.bc.TxPoolModel.UpdateTxsInTransact(dbTx, pendingUpdatePoolTxs)
			if err != nil {
				return err
			}
			return c.bc.TxPoolModel.DeleteTxsInTransact(dbTx, pendingDeletePoolTxs)
		})
		if err != nil {
			panic("update tx pool failed: " + err.Error())
		}

		if c.shouldCommit(curBlock) {
			start := time.Now()
			logx.Infof("commit new block, height=%d", curBlock.BlockHeight)
			curBlock, err = c.commitNewBlock(curBlock)
			l2BlockHeightMetics.Set(float64(curBlock.BlockHeight))
			logx.Infof("commit new block success, blockSize=%d", curBlock.BlockSize)

			if err != nil {
				logx.Errorf("commit new block error, err=%s", err.Error())
				panic("commit new block failed: " + err.Error())
			}
			commitOperationMetics.Set(float64(time.Since(start).Milliseconds()))
		}
	}
}

func (c *Committer) Shutdown() {
	c.running = false
	c.bc.Statedb.Close()
	c.bc.ChainDB.Close()
}

func (c *Committer) restoreExecutedTxs() (*block.Block, error) {
	bc := c.bc
	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		return nil, err
	}
	curBlock, err := bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, err
	}

	executedTxs, err := c.bc.TxPoolModel.GetTxsByStatus(tx.StatusExecuted)
	if err != nil {
		return nil, err
	}

	if curBlock.BlockStatus > block.StatusProposing {
		if len(executedTxs) != 0 {
			return nil, errors.New("no proposing block but exist executed txs")
		}
		return curBlock, nil
	}

	if err := c.bc.StateDB().MarkGasAccountAsPending(); err != nil {
		return nil, err
	}
	for _, executedTx := range executedTxs {
		err = c.bc.ApplyTransaction(executedTx)
		if err != nil {
			return nil, err
		}
	}

	return curBlock, nil
}

func (c *Committer) createNewBlock(curBlock *block.Block, poolTx *tx.Tx) error {
	return c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
		err := c.bc.TxPoolModel.UpdateTxsInTransact(dbTx, []*tx.Tx{poolTx})
		if err != nil {
			return err
		}

		return c.bc.BlockModel.CreateBlockInTransact(dbTx, curBlock)
	})
}

func (c *Committer) shouldCommit(curBlock *block.Block) bool {
	var now = time.Now()
	if (len(c.bc.Statedb.Txs) > 0 && now.Unix()-curBlock.CreatedAt.Unix() >= MaxCommitterInterval) ||
		len(c.bc.Statedb.Txs) >= c.maxTxsPerBlock {
		return true
	}

	return false
}

func (c *Committer) commitNewBlock(curBlock *block.Block) (*block.Block, error) {
	blockSize := c.computeCurrentBlockSize()
	start := time.Now()
	blockStates, err := c.bc.CommitNewBlock(blockSize, curBlock.CreatedAt.UnixMilli())
	if err != nil {
		return nil, err
	}
	stateDBOperationMetics.Set(float64(time.Since(start).Milliseconds()))

	start = time.Now()
	err = c.bc.Statedb.SyncGasAccountToRedis()
	if err != nil {
		return nil, err
	}
	stateDBSyncOperationMetics.Set(float64(time.Since(start).Milliseconds()))

	start = time.Now()
	// update db
	err = c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
		// create block for commit
		if blockStates.CompressedBlock != nil {
			err = c.bc.DB().CompressedBlockModel.CreateCompressedBlockInTransact(tx, blockStates.CompressedBlock)
			if err != nil {
				return err
			}
		}
		// create or update account
		if len(blockStates.PendingAccount) != 0 {
			err = c.bc.DB().AccountModel.UpdateAccountsInTransact(tx, blockStates.PendingAccount)
			if err != nil {
				return err
			}
		}
		// create account history
		if len(blockStates.PendingAccountHistory) != 0 {
			err = c.bc.DB().AccountHistoryModel.CreateAccountHistoriesInTransact(tx, blockStates.PendingAccountHistory)
			if err != nil {
				return err
			}
		}
		// create or update nft
		if len(blockStates.PendingNft) != 0 {
			err = c.bc.DB().L2NftModel.UpdateNftsInTransact(tx, blockStates.PendingNft)
			if err != nil {
				return err
			}
		}
		// create nft history
		if len(blockStates.PendingNftHistory) != 0 {
			err = c.bc.DB().L2NftHistoryModel.CreateNftHistoriesInTransact(tx, blockStates.PendingNftHistory)
			if err != nil {
				return err
			}
		}
		// delete txs from tx pool
		err := c.bc.DB().TxPoolModel.DeleteTxsInTransact(tx, blockStates.Block.Txs)
		if err != nil {
			return err
		}
		// update block
		blockStates.Block.ClearTxsModel()
		return c.bc.DB().BlockModel.UpdateBlockInTransact(tx, blockStates.Block)
	})
	if err != nil {
		return nil, err
	}
	sqlDBOperationMetics.Set(float64(time.Since(start).Milliseconds()))

	return blockStates.Block, nil
}

func (c *Committer) computeCurrentBlockSize() int {
	var blockSize int
	for i := 0; i < len(c.optionalBlockSizes); i++ {
		if len(c.bc.Statedb.Txs) <= c.optionalBlockSizes[i] {
			blockSize = c.optionalBlockSizes[i]
			break
		}
	}
	return blockSize
}

func (c *Committer) getLatestExecutedRequestId() (int64, error) {

	statuses := []int{
		tx.StatusExecuted,
		tx.StatusPacked,
		tx.StatusCommitted,
		tx.StatusVerified,
	}

	txTypes := []int64{
		types.TxTypeRegisterZns,
		types.TxTypeDeposit,
		types.TxTypeDepositNft,
		types.TxTypeFullExit,
		types.TxTypeFullExitNft,
	}

	latestTx, err := c.bc.TxPoolModel.GetLatestTx(txTypes, statuses)
	if err != nil && err != types.DbErrNotFound {
		logx.Errorf("get latest executed tx failed: %v", err)
		return -1, err
	} else if err == types.DbErrNotFound {
		return -1, nil
	}

	p, err := c.bc.PriorityRequestModel.GetPriorityRequestsByL2TxHash(latestTx.TxHash)
	if err != nil {
		logx.Errorf("get priority request by txhash: %s failed: %v", latestTx.TxHash, err)
		return -1, err
	}

	return p.RequestId, nil
}
