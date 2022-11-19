package committer

import (
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"strconv"
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

	l2BlockMemoryHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_memory_height",
		Help:      "l2Block_memory_height metrics.",
	})

	l2BlockRedisHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_redis_height",
		Help:      "l2Block_memory_height metrics.",
	})

	l2BlockDbHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_db_height",
		Help:      "l2Block_memory_height metrics.",
	})

	l2BlockCommitToChainHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_commit_to_chain_height",
		Help:      "l2Block_memory_height metrics.",
	})

	l2BlockCommitConfirmByChainHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_commit_confirm_by_chain_height",
		Help:      "l2Block_memory_height metrics.",
	})

	l2BlockSubmitToVerifyHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_submit_to_verify_height",
		Help:      "l2Block_memory_height metrics.",
	})

	l2BlockVerifiedHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_verified_height",
		Help:      "l2Block_memory_height metrics.",
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
	executeTxApply1TxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "exec_tx_apply_1_transaction_time",
		Help:      "execute txs apply 1 transaction operation time",
	})

	updatePoolTxsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_pool_txs_time",
		Help:      "update pool txs time",
	})

	addCompressedBlockMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "add_compressed_block_time",
		Help:      "add compressed block time",
	})

	updateAccountMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_account_time",
		Help:      "update account time",
	})

	addAccountHistoryMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "add_account_history_time",
		Help:      "add account history time",
	})

	deletePoolTxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "delete_pool_tx_time",
		Help:      "delete pool tx time",
	})

	updateBlockMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_block_time",
		Help:      "update block time time",
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

	bc                     *core.BlockChain
	txWorker               *core.TxWorker
	treeWorker             *core.Worker
	saveBlockTxWorker      *core.Worker
	updatePoolTxWorker     *core.Worker
	syncStateToRedisWorker *core.Worker
}

type PendingMap struct {
	PendingAccountMap map[int64]*types.AccountInfo
	PendingNftMap     map[int64]*nft.L2Nft
}
type UpdatePoolTx struct {
	PendingUpdatePoolTxs []*tx.Tx
	PendingDeletePoolTxs []*tx.Tx
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

	if err := prometheus.Register(l2BlockMemoryHeightMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
	}

	if err := prometheus.Register(l2BlockRedisHeightMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
	}

	if err := prometheus.Register(l2BlockDbHeightMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
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
	if err := prometheus.Register(executeTxApply1TxMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register executeTxApply1TxMetrics error: %v", err)
	}
	if err := prometheus.Register(updatePoolTxsMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register updatePoolTxsMetrics error: %v", err)
	}
	if err := prometheus.Register(addCompressedBlockMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register addCompressedBlockMetrics error: %v", err)
	}

	if err := prometheus.Register(updateAccountMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register updateAccountMetrics error: %v", err)
	}

	if err := prometheus.Register(addAccountHistoryMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register addAccountHistoryMetrics error: %v", err)
	}

	if err := prometheus.Register(deletePoolTxMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register deletePoolTxMetrics error: %v", err)
	}

	if err := prometheus.Register(updateBlockMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register updateBlockMetrics error: %v", err)
	}

	committer := &Committer{
		running:            true,
		config:             config,
		maxTxsPerBlock:     config.BlockConfig.OptionalBlockSizes[len(config.BlockConfig.OptionalBlockSizes)-1],
		optionalBlockSizes: config.BlockConfig.OptionalBlockSizes,
		bc:                 bc,
	}

	return committer, nil
}

func (c *Committer) Run() {
	c.txWorker = core.ExecuteTxWorker(6000, func() {
		c.executeTxFunc()
	})
	c.syncStateToRedisWorker = core.SyncStateCacheToRedisWorker(20000, func(item interface{}) {
		c.syncStateCacheToRedisFunc(item.(*PendingMap))
	})
	c.updatePoolTxWorker = core.UpdatePoolTxWorker(6000, func(item interface{}) {
		c.updatePoolTxFunc(item.(*UpdatePoolTx))
	})
	c.treeWorker = core.ExecuteTreeWorker(10, func(item interface{}) {
		c.executeTreeFunc(item.(*statedb.StateDataCopy))
	})
	c.saveBlockTxWorker = core.SaveBlockTransactionWorker(10, func(item interface{}) {
		c.saveBlockTransactionFunc(item.(*block.BlockStates))
	})

	c.txWorker.Start()
	c.syncStateToRedisWorker.Start()
	c.updatePoolTxWorker.Start()
	c.treeWorker.Start()
	c.saveBlockTxWorker.Start()

	c.pullPoolTxs()
}

func (c *Committer) pullPoolTxs() {
	executedTx, err := c.bc.TxPoolModel.GetLatestExecutedTx()
	if err != nil {
		logx.Errorf("get executed tx from tx pool failed:%s", err.Error())
		panic("get executed tx from tx pool failed: " + err.Error())
	}
	var executedTxMaxId uint = 0
	if executedTx != nil {
		executedTxMaxId = executedTx.ID
	}
	for {
		if !c.running {
			break
		}
		logx.Infof("get pool txs executedTxMaxId=%d", executedTxMaxId)
		pendingTxs, err := c.bc.TxPoolModel.GetTxsByStatusAndMaxId(tx.StatusPending, executedTxMaxId, 50)
		if err != nil {
			logx.Errorf("get pending transactions from tx pool failed:%s", err.Error())
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if len(pendingTxs) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		checkMaxId := executedTxMaxId
		success := true
		for _, poolTx := range pendingTxs {
			if int(poolTx.ID)-int(checkMaxId) != 1 {
				logx.Errorf("not equal id=%s", poolTx.ID)
				time.Sleep(50 * time.Millisecond)
				success = false
				break
			}
			checkMaxId = poolTx.ID
		}
		if success {
			for _, poolTx := range pendingTxs {
				executedTxMaxId = poolTx.ID
				c.txWorker.Enqueue(poolTx)
			}
		}
	}
}

func (c *Committer) getPoolTxsFromQueue() []*tx.Tx {
	//todo list
	pendingUpdatePoolTxs := make([]*tx.Tx, 0, 300)
	for {
		select {
		case i := <-c.txWorker.GetJobQueue():
			pendingUpdatePoolTxs = append(pendingUpdatePoolTxs, i.(*tx.Tx))
		default:
			return pendingUpdatePoolTxs
		}
		if len(pendingUpdatePoolTxs) == 300 {
			return pendingUpdatePoolTxs
		}
	}
}

func (c *Committer) executeTxFunc() {
	latestRequestId, err := c.getLatestExecutedRequestId()
	if err != nil {
		logx.Errorf("get latest executed request id failed:%s", err.Error())
		panic("get latest executed request id failed: " + err.Error())
	}
	var subPendingTxs []*tx.Tx
	var pendingTxs []*tx.Tx
	for {
		curBlock := c.bc.CurrentBlock()
		if curBlock.BlockStatus > block.StatusProposing {
			curBlock, err = c.bc.InitNewBlock()
			if err != nil {
				logx.Errorf("propose new block failed:%s", err)
				panic("propose new block failed: " + err.Error())
			}
		}
		if subPendingTxs != nil && len(subPendingTxs) > 0 {
			pendingTxs = subPendingTxs
			subPendingTxs = nil
		} else {
			pendingTxs = c.getPoolTxsFromQueue()
		}
		for len(pendingTxs) == 0 {
			if c.shouldCommit(curBlock) {
				break
			}
			time.Sleep(100 * time.Millisecond)
			pendingTxs = c.getPoolTxsFromQueue()
		}
		pendingTxNumMetrics.Set(float64(len(pendingTxs)))
		pendingUpdatePoolTxs := make([]*tx.Tx, 0, len(pendingTxs))
		pendingDeletePoolTxs := make([]*tx.Tx, 0, len(pendingTxs))
		start := time.Now()
		for _, poolTx := range pendingTxs {
			logx.Error("pendingTxs----: ", poolTx.ID)
			if c.shouldCommit(curBlock) {
				subPendingTxs = append(subPendingTxs, poolTx)
				continue
			}
			logx.Infof("apply transaction, txHash=%s", poolTx.TxHash)
			startApplyTx := time.Now()
			err = c.bc.ApplyTransaction(poolTx)
			executeTxApply1TxMetrics.Set(float64(time.Since(startApplyTx).Milliseconds()))
			if err != nil {
				logx.Errorf("apply pool tx ID: %d failed, err %v ", poolTx.ID, err)
				if types.IsPriorityOperationTx(poolTx.TxType) {
					logx.Errorf("apply priority pool tx failed,id=%s,error=%s", strconv.Itoa(int(poolTx.ID)), err.Error())
					panic("apply priority pool tx failed,id=" + strconv.Itoa(int(poolTx.ID)) + ",error=" + err.Error())
				}
				poolTx.TxStatus = tx.StatusFailed
				pendingDeletePoolTxs = append(pendingDeletePoolTxs, poolTx)
				continue
			}

			if types.IsPriorityOperationTx(poolTx.TxType) {
				request, err := c.bc.PriorityRequestModel.GetPriorityRequestsByL2TxHash(poolTx.TxHash)
				if err == nil {
					priorityOperationMetric.Set(float64(request.RequestId))
					priorityOperationHeightMetric.Set(float64(request.L1BlockHeight))
					if latestRequestId != -1 && request.RequestId != latestRequestId+1 {
						logx.Errorf("invalid request id=%s,txHash=%s", strconv.Itoa(int(request.RequestId)), err.Error())
						panic("invalid request id=" + strconv.Itoa(int(request.RequestId)) + ",txHash=" + poolTx.TxHash)
					}
					latestRequestId = request.RequestId
				} else {
					logx.Errorf("query txHash from priority request txHash=%s,error=%s", poolTx.TxHash, err.Error())
					panic("query txHash from priority request txHash=" + poolTx.TxHash + ",error=" + err.Error())
				}
			}

			// Write the proposed block into database when the first transaction executed.
			if len(c.bc.Statedb.Txs) == 1 {
				err = c.createNewBlock(curBlock, poolTx)
				if err != nil {
					logx.Errorf("create new block failed:%s", err.Error())
					panic("create new block failed" + err.Error())
				}
			} else {
				pendingUpdatePoolTxs = append(pendingUpdatePoolTxs, poolTx)
			}
		}
		executeTxOperationMetrics.Set(float64(time.Since(start).Milliseconds()))

		c.bc.Statedb.SyncPendingAccountToMemoryCache(c.bc.Statedb.PendingAccountMap)
		c.bc.Statedb.SyncPendingNftToMemoryCache(c.bc.Statedb.PendingNftMap)

		c.enqueueSyncStateCacheToRedis(c.bc.Statedb.PendingAccountMap, c.bc.Statedb.PendingNftMap)
		c.enqueueUpdatePoolTx(pendingUpdatePoolTxs, pendingDeletePoolTxs)

		if c.shouldCommit(curBlock) {
			start := time.Now()
			logx.Infof("commit new block, height=%d,blockSize=%s", curBlock.BlockHeight, curBlock.BlockSize)
			stateDataCopy := &statedb.StateDataCopy{
				StateCache:   c.bc.Statedb.StateCache,
				CurrentBlock: curBlock,
			}
			c.treeWorker.Enqueue(stateDataCopy)
			l2BlockMemoryHeightMetric.Set(float64(stateDataCopy.CurrentBlock.BlockHeight))
			previousHeight := stateDataCopy.CurrentBlock.BlockHeight
			curBlock, err = c.bc.InitNewBlock()
			logx.Infof("init new block, current height=%s,previous height=%d", curBlock.BlockHeight, previousHeight)
			if err != nil {
				logx.Errorf("propose new block failed:%s ", err.Error())
				panic("propose new block failed: " + err.Error())
			}
			if err != nil {
				logx.Errorf("commit new block error, err=%s", err.Error())
				panic("commit new block failed: " + err.Error())
			}
			commitOperationMetics.Set(float64(time.Since(start).Milliseconds()))
		}
	}
}

func (c *Committer) enqueueUpdatePoolTx(pendingUpdatePoolTxs []*tx.Tx, pendingDeletePoolTxs []*tx.Tx) {
	updatePoolTxMap := &UpdatePoolTx{
		PendingUpdatePoolTxs: make([]*tx.Tx, 0, len(pendingUpdatePoolTxs)),
		PendingDeletePoolTxs: make([]*tx.Tx, 0, len(pendingDeletePoolTxs)),
	}
	for _, poolTx := range pendingUpdatePoolTxs {
		updatePoolTxMap.PendingUpdatePoolTxs = append(updatePoolTxMap.PendingUpdatePoolTxs, poolTx.DeepCopy())
	}
	for _, poolTx := range pendingDeletePoolTxs {
		updatePoolTxMap.PendingDeletePoolTxs = append(updatePoolTxMap.PendingDeletePoolTxs, poolTx.DeepCopy())
	}
	c.updatePoolTxWorker.Enqueue(updatePoolTxMap)
}

func (c *Committer) updatePoolTxFunc(updatePoolTxMap *UpdatePoolTx) {
	for _, pendingDeletePoolTx := range updatePoolTxMap.PendingDeletePoolTxs {
		updatePoolTxMap.PendingUpdatePoolTxs = append(updatePoolTxMap.PendingUpdatePoolTxs, pendingDeletePoolTx)
	}
	err := c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
		err := c.bc.TxPoolModel.UpdateTxsInTransact(dbTx, updatePoolTxMap.PendingUpdatePoolTxs)
		if err != nil {
			logx.Error("update tx pool failed:", err)
		}
		return c.bc.TxPoolModel.DeleteTxsBatchInTransact(dbTx, updatePoolTxMap.PendingDeletePoolTxs)
	})
	if err != nil {
		logx.Error("update tx pool failed:", err)
	}
}

func (c *Committer) enqueueSyncStateCacheToRedis(originPendingAccountMap map[int64]*types.AccountInfo, originPendingNftMap map[int64]*nft.L2Nft) {
	pendingMap := &PendingMap{
		PendingAccountMap: make(map[int64]*types.AccountInfo, len(originPendingAccountMap)),
		PendingNftMap:     make(map[int64]*nft.L2Nft, len(originPendingNftMap)),
	}
	for _, accountInfo := range originPendingAccountMap {
		pendingMap.PendingAccountMap[accountInfo.AccountIndex] = accountInfo.DeepCopy()
	}
	for _, nftInfo := range originPendingNftMap {
		pendingMap.PendingNftMap[nftInfo.NftIndex] = nftInfo.DeepCopy()
	}
	c.syncStateToRedisWorker.Enqueue(pendingMap)
}

func (c *Committer) syncStateCacheToRedisFunc(pendingMap *PendingMap) {
	c.bc.Statedb.SyncPendingAccountToRedis(pendingMap.PendingAccountMap)
	c.bc.Statedb.SyncPendingNftToRedis(pendingMap.PendingNftMap)
}

func (c *Committer) executeTreeFunc(stateDataCopy *statedb.StateDataCopy) {
	logx.Infof("commit new block start blockHeight:%s", stateDataCopy.CurrentBlock.BlockHeight)
	start := time.Now()
	blockSize := c.computeCurrentBlockSize(stateDataCopy)
	blockStates, err := c.bc.CommitNewBlock(blockSize, stateDataCopy)
	if err != nil {
		logx.Errorf("commit new block failed:%s,blockHeight:%s", err, stateDataCopy.CurrentBlock.BlockHeight)
		panic("commit new block failed: " + err.Error())
	}
	stateDBOperationMetics.Set(float64(time.Since(start).Milliseconds()))

	start = time.Now()
	//todo
	err = c.bc.Statedb.SyncGasAccountToRedis()
	if err != nil {
		logx.Errorf("update pool tx to pending failed:%s", err.Error())
		panic("update pool tx to pending failed: " + err.Error())
	}
	c.saveBlockTxWorker.Enqueue(blockStates)
	stateDBSyncOperationMetics.Set(float64(time.Since(start).Milliseconds()))
	l2BlockRedisHeightMetric.Set(float64(blockStates.Block.BlockHeight))
}

func (c *Committer) saveBlockTransactionFunc(blockStates *block.BlockStates) {
	logx.Infof("save block transaction start, blockHeight:%s", blockStates.Block.BlockHeight)
	start := time.Now()
	// update db
	err := c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
		start := time.Now()
		// create block for commit
		var err error
		if blockStates.CompressedBlock != nil {
			err = c.bc.DB().CompressedBlockModel.CreateCompressedBlockInTransact(tx, blockStates.CompressedBlock)
			if err != nil {
				return err
			}
		}
		addCompressedBlockMetrics.Set(float64(time.Since(start).Milliseconds()))
		start = time.Now()
		// create or update account
		if len(blockStates.PendingAccount) != 0 {
			err = c.bc.DB().AccountModel.UpdateAccountsInTransact(tx, blockStates.PendingAccount)
			if err != nil {
				return err
			}
		}
		updateAccountMetrics.Set(float64(time.Since(start).Milliseconds()))
		start = time.Now()
		// create account history
		if len(blockStates.PendingAccountHistory) != 0 {
			err = c.bc.DB().AccountHistoryModel.CreateAccountHistoriesInTransact(tx, blockStates.PendingAccountHistory)
			if err != nil {
				return err
			}
		}
		addAccountHistoryMetrics.Set(float64(time.Since(start).Milliseconds()))
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
		start = time.Now()
		// delete txs from tx pool
		err = c.bc.DB().TxPoolModel.DeleteTxsBatchInTransact(tx, blockStates.Block.Txs)
		if err != nil {
			return err
		}
		deletePoolTxMetrics.Set(float64(time.Since(start).Milliseconds()))

		// update block
		blockStates.Block.ClearTxsModel()
		start = time.Now()
		err = c.bc.DB().BlockModel.UpdateBlockInTransact(tx, blockStates.Block)
		updateBlockMetrics.Set(float64(time.Since(start).Milliseconds()))
		return err

	})
	if err != nil {
		logx.Errorf("save block transaction failed:%s,blockHeight:%s", err.Error(), blockStates.Block.BlockHeight)
		panic("save block transaction failed: " + err.Error())
		//todo 重试优化
	}
	c.bc.Statedb.UpdatePrunedBlockHeight(blockStates.Block.BlockHeight)
	sqlDBOperationMetics.Set(float64(time.Since(start).Milliseconds()))
	l2BlockDbHeightMetric.Set(float64(blockStates.Block.BlockHeight))

}

func (c *Committer) Shutdown() {
	c.running = false
	c.txWorker.Stop()
	c.treeWorker.Stop()
	c.syncStateToRedisWorker.Stop()
	c.saveBlockTxWorker.Stop()
	c.updatePoolTxWorker.Stop()
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

//func (c *Committer) commitNewBlock(curBlock *block.Block) (*block.Block, error) {
//	blockSize := c.computeCurrentBlockSize()
//	start := time.Now()
//	blockStates, err := c.bc.CommitNewBlock(blockSize, curBlock.CreatedAt.UnixMilli())
//	if err != nil {
//		return nil, err
//	}
//	stateDBOperationMetics.Set(float64(time.Since(start).Milliseconds()))
//
//	start = time.Now()
//	err = c.bc.Statedb.SyncGasAccountToRedis()
//	if err != nil {
//		return nil, err
//	}
//	stateDBSyncOperationMetics.Set(float64(time.Since(start).Milliseconds()))
//
//	start = time.Now()
//	// update db
//	err = c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
//		// create block for commit
//		if blockStates.CompressedBlock != nil {
//			err = c.bc.DB().CompressedBlockModel.CreateCompressedBlockInTransact(tx, blockStates.CompressedBlock)
//			if err != nil {
//				return err
//			}
//		}
//		// create or update account
//		if len(blockStates.PendingAccount) != 0 {
//			err = c.bc.DB().AccountModel.UpdateAccountsInTransact(tx, blockStates.PendingAccount)
//			if err != nil {
//				return err
//			}
//		}
//		// create account history
//		if len(blockStates.PendingAccountHistory) != 0 {
//			err = c.bc.DB().AccountHistoryModel.CreateAccountHistoriesInTransact(tx, blockStates.PendingAccountHistory)
//			if err != nil {
//				return err
//			}
//		}
//		// create or update nft
//		if len(blockStates.PendingNft) != 0 {
//			err = c.bc.DB().L2NftModel.UpdateNftsInTransact(tx, blockStates.PendingNft)
//			if err != nil {
//				return err
//			}
//		}
//		// create nft history
//		if len(blockStates.PendingNftHistory) != 0 {
//			err = c.bc.DB().L2NftHistoryModel.CreateNftHistoriesInTransact(tx, blockStates.PendingNftHistory)
//			if err != nil {
//				return err
//			}
//		}
//		// delete txs from tx pool
//		err := c.bc.DB().TxPoolModel.DeleteTxsInTransact(tx, blockStates.Block.Txs)
//		if err != nil {
//			return err
//		}
//		// update block
//		blockStates.Block.ClearTxsModel()
//		return c.bc.DB().BlockModel.UpdateBlockInTransact(tx, blockStates.Block)
//	})
//	if err != nil {
//		return nil, err
//	}
//	sqlDBOperationMetics.Set(float64(time.Since(start).Milliseconds()))
//
//	return blockStates.Block, nil
//}

func (c *Committer) computeCurrentBlockSize(stateCopy *statedb.StateDataCopy) int {
	var blockSize int
	for i := 0; i < len(c.optionalBlockSizes); i++ {
		if len(stateCopy.StateCache.Txs) <= c.optionalBlockSizes[i] {
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
