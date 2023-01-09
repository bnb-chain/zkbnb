package committer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb/common/gopool"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/panjf2000/ants/v2"
	"sort"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/chain"
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

	AccountLatestVersionTreeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_account_latest_version",
		Help:      "Account latest version metrics.",
	})
	AccountRecentVersionTreeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_account_recent_version",
		Help:      "Account recent version metrics.",
	})
	NftTreeLatestVersionMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_nft_latest_version",
		Help:      "Nft latest version metrics.",
	})
	NftTreeRecentVersionMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "commit_nft_recent_version",
		Help:      "Nft recent version metrics.",
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
	updateAccountAssetTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_account_asset_tree_time",
		Help:      "updateAccountAssetTreeMetrics",
	})
	updateAccountTreeAndNftTreeMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_account_tree_and_nft_tree_time",
		Help:      "updateAccountTreeAndNftTreeMetrics",
	})
	stateDBSyncOperationMetics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "state_sync_time",
		Help:      "stateDB sync operation time",
	})

	preSaveBlockDataMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "pre_save_block_data_time",
		Help:      "pre save block data time",
	}, []string{"type"})

	saveBlockDataMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "save_block_data_time",
		Help:      "save block data time",
	}, []string{"type"})

	finalSaveBlockDataMetrics = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "final_save_block_data_time",
		Help:      "final save block data time",
	}, []string{"type"})

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

	addTxDetailsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "add_tx_details_time",
		Help:      "add tx details time",
	})

	addTxsMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "add_txs_time",
		Help:      "add txs time",
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

	saveAccountsGoroutineMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "save_accounts_goroutine_time",
		Help:      "save accounts goroutine time",
	})

	getPendingPoolTxMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "get_pending_pool_tx_time",
		Help:      "get pending pool tx time",
	})

	updatePoolTxsProcessingMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "update_pool_txs_processing_time",
		Help:      "update pool txs processing time",
	})
	syncAccountToRedisMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "sync_account_to_redis_time",
		Help:      "sync account to redis time",
	})
	getPendingTxsToQueueMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "get_pending_txs_to_queue_count",
		Help:      "get pending txs to queue count",
	})

	txWorkerQueueMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "tx_worker_queue_count",
		Help:      "tx worker queue count",
	})

	executeTxMetrics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "execute_tx_count",
		Help:      "execute tx count",
	})

	updateAccountAssetTreeTxMetrics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "update_account_asset_tree_tx_count",
		Help:      "update_account_asset_tree_tx_count",
	})
	updateAccountTreeAndNftTreeTxMetrics = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "zkbnb",
		Name:      "update_account_tree_and_nft_tree_tx_count",
		Help:      "update_account_tree_and_nft_tree_tx_count",
	})

	accountAssetTreeQueueMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_asset_tree_queue_count",
		Help:      "account asset tree queue count",
	})

	accountTreeAndNftTreeQueueMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "account_tree_and_nft_tree_queue_count",
		Help:      "account tree and nft tree queue count",
	})

	antsPoolGaugeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "ants_pool_count",
		Help:      "ants pool count",
	}, []string{"type"})

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
		OptionalBlockSizes    []int
		SaveBlockDataPoolSize int `json:",optional"`
	}
	LogConf logx.LogConf
}

type Committer struct {
	running            bool
	config             *Config
	maxTxsPerBlock     int
	optionalBlockSizes []int

	bc                                *core.BlockChain
	executeTxWorker                   *core.TxWorker
	updatePoolTxWorker                *core.Worker
	syncAccountToRedisWorker          *core.Worker
	updateAccountAssetTreeWorker      *core.Worker
	updateAccountTreeAndNftTreeWorker *core.Worker
	preSaveBlockDataWorker            *core.Worker
	saveBlockDataWorker               *core.Worker
	finalSaveBlockDataWorker          *core.Worker
	pool                              *ants.Pool
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

	err := initMetrics()
	if err != nil {
		return nil, err
	}

	bc, err := core.NewBlockChain(&config.ChainConfig, "committer")
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	saveBlockDataPoolSize := config.BlockConfig.SaveBlockDataPoolSize
	if saveBlockDataPoolSize == 0 {
		saveBlockDataPoolSize = 100
	}
	pool, err := ants.NewPool(saveBlockDataPoolSize)

	committer := &Committer{
		running:            true,
		config:             config,
		maxTxsPerBlock:     config.BlockConfig.OptionalBlockSizes[len(config.BlockConfig.OptionalBlockSizes)-1],
		optionalBlockSizes: config.BlockConfig.OptionalBlockSizes,
		bc:                 bc,
		pool:               pool,
	}

	return committer, nil
}

func initMetrics() error {
	if err := prometheus.Register(priorityOperationMetric); err != nil {
		return fmt.Errorf("prometheus.Register priorityOperationMetric error: %v", err)
	}
	if err := prometheus.Register(priorityOperationHeightMetric); err != nil {
		return fmt.Errorf("prometheus.Register priorityOperationHeightMetric error: %v", err)
	}
	if err := prometheus.Register(l2BlockMemoryHeightMetric); err != nil {
		return fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
	}
	if err := prometheus.Register(l2BlockRedisHeightMetric); err != nil {
		return fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
	}
	if err := prometheus.Register(l2BlockDbHeightMetric); err != nil {
		return fmt.Errorf("prometheus.Register l2BlockMemoryHeightMetric error: %v", err)
	}
	if err := prometheus.Register(AccountLatestVersionTreeMetric); err != nil {
		return fmt.Errorf("prometheus.Register AccountLatestVersionTreeMetric error: %v", err)
	}
	if err := prometheus.Register(AccountRecentVersionTreeMetric); err != nil {
		return fmt.Errorf("prometheus.Register AccountRecentVersionTreeMetric error: %v", err)
	}
	if err := prometheus.Register(NftTreeLatestVersionMetric); err != nil {
		return fmt.Errorf("prometheus.Register NftTreeLatestVersionMetric error: %v", err)
	}
	if err := prometheus.Register(NftTreeRecentVersionMetric); err != nil {
		return fmt.Errorf("prometheus.Register NftTreeRecentVersionMetric error: %v", err)
	}
	if err := prometheus.Register(commitOperationMetics); err != nil {
		return fmt.Errorf("prometheus.Register commitOperationMetics error: %v", err)
	}
	if err := prometheus.Register(pendingTxNumMetrics); err != nil {
		return fmt.Errorf("prometheus.Register pendingTxNumMetrics error: %v", err)
	}
	if err := prometheus.Register(executeTxOperationMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxOperationMetrics error: %v", err)
	}
	if err := prometheus.Register(updateAccountAssetTreeMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountAssetTreeMetrics error: %v", err)
	}
	if err := prometheus.Register(stateDBSyncOperationMetics); err != nil {
		return fmt.Errorf("prometheus.Register stateDBSyncOperationMetics error: %v", err)
	}
	if err := prometheus.Register(preSaveBlockDataMetrics); err != nil {
		return fmt.Errorf("prometheus.Register preSaveBlockDataMetrics error: %v", err)
	}
	if err := prometheus.Register(saveBlockDataMetrics); err != nil {
		return fmt.Errorf("prometheus.Register saveBlockDataMetrics error: %v", err)
	}
	if err := prometheus.Register(finalSaveBlockDataMetrics); err != nil {
		return fmt.Errorf("prometheus.Register finalSaveBlockDataMetrics error: %v", err)
	}
	if err := prometheus.Register(executeTxApply1TxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxApply1TxMetrics error: %v", err)
	}
	if err := prometheus.Register(updatePoolTxsMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updatePoolTxsMetrics error: %v", err)
	}
	if err := prometheus.Register(addCompressedBlockMetrics); err != nil {
		return fmt.Errorf("prometheus.Register addCompressedBlockMetrics error: %v", err)
	}
	if err := prometheus.Register(updateAccountMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountMetrics error: %v", err)
	}
	if err := prometheus.Register(addAccountHistoryMetrics); err != nil {
		return fmt.Errorf("prometheus.Register addAccountHistoryMetrics error: %v", err)
	}
	if err := prometheus.Register(addTxDetailsMetrics); err != nil {
		return fmt.Errorf("prometheus.Register addTxDetailsMetrics error: %v", err)
	}
	if err := prometheus.Register(addTxsMetrics); err != nil {
		return fmt.Errorf("prometheus.Register addTxsMetrics error: %v", err)
	}
	if err := prometheus.Register(deletePoolTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register deletePoolTxMetrics error: %v", err)
	}
	if err := prometheus.Register(updateBlockMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateBlockMetrics error: %v", err)
	}
	if err := prometheus.Register(saveAccountsGoroutineMetrics); err != nil {
		return fmt.Errorf("prometheus.Register saveAccountsGoroutineMetrics error: %v", err)
	}
	if err := prometheus.Register(getPendingPoolTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register getPendingPoolTxMetrics error: %v", err)
	}
	if err := prometheus.Register(updatePoolTxsProcessingMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updatePoolTxsProcessingMetrics error: %v", err)
	}
	if err := prometheus.Register(getPendingTxsToQueueMetric); err != nil {
		return fmt.Errorf("prometheus.Register getPendingTxsToQueueMetric error: %v", err)
	}
	if err := prometheus.Register(txWorkerQueueMetric); err != nil {
		return fmt.Errorf("prometheus.Register txWorkerQueueMetric error: %v", err)
	}
	if err := prometheus.Register(executeTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register executeTxMetrics error: %v", err)
	}
	if err := prometheus.Register(updateAccountAssetTreeTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountAssetTreeTxMetrics error: %v", err)
	}
	if err := prometheus.Register(updateAccountTreeAndNftTreeTxMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountTreeAndNftTreeTxMetrics error: %v", err)
	}
	if err := prometheus.Register(updateAccountTreeAndNftTreeMetrics); err != nil {
		return fmt.Errorf("prometheus.Register updateAccountTreeAndNftTreeMetrics error: %v", err)
	}
	if err := prometheus.Register(accountTreeAndNftTreeQueueMetric); err != nil {
		return fmt.Errorf("prometheus.Register accountTreeAndNftTreeQueueMetric error: %v", err)
	}
	if err := prometheus.Register(accountAssetTreeQueueMetric); err != nil {
		return fmt.Errorf("prometheus.Register accountAssetTreeQueueMetric error: %v", err)
	}
	if err := prometheus.Register(antsPoolGaugeMetric); err != nil {
		return fmt.Errorf("prometheus.Register antsPoolGaugeMetric error: %v", err)
	}
	if err := prometheus.Register(l2BlockHeightMetics); err != nil {
		return fmt.Errorf("prometheus.Register l2BlockHeightMetics error: %v", err)
	}
	if err := prometheus.Register(poolTxL1ErrorCountMetics); err != nil {
		return fmt.Errorf("prometheus.Register poolTxL1ErrorCountMetics error: %v", err)
	}
	if err := prometheus.Register(poolTxL2ErrorCountMetics); err != nil {
		return fmt.Errorf("prometheus.Register poolTxL2ErrorCountMetics error: %v", err)
	}
	return nil
}

func (c *Committer) Run() {
	c.executeTxWorker = core.ExecuteTxWorker(10000, func() {
		c.executeTxFunc()
	})
	c.updatePoolTxWorker = core.UpdatePoolTxWorker(100000, func(item interface{}) {
		c.updatePoolTxFunc(item.(*UpdatePoolTx))
	})
	c.syncAccountToRedisWorker = core.SyncAccountToRedisWorker(200000, func(item interface{}) {
		c.syncAccountToRedisFunc(item.(*PendingMap))
	})
	c.preSaveBlockDataWorker = core.PreSaveBlockDataWorker(10, func(item interface{}) {
		c.preSaveBlockDataFunc(item.(*statedb.StateDataCopy))
	})
	c.updateAccountAssetTreeWorker = core.UpdateAccountAssetTreeWorker(10, func(item interface{}) {
		c.updateAccountAssetTreeFunc(item.(*statedb.StateDataCopy))
	})
	c.updateAccountTreeAndNftTreeWorker = core.UpdateAccountTreeAndNftTreeWorker(10, func(item interface{}) {
		c.updateAccountTreeAndNftTreeFunc(item.(*statedb.StateDataCopy))
	})
	c.saveBlockDataWorker = core.SaveBlockDataWorker(10, func(item interface{}) {
		c.saveBlockDataFunc(item.(*block.BlockStates))
	})
	c.finalSaveBlockDataWorker = core.FinalSaveBlockDataWorker(10, func(item interface{}) {
		c.finalSaveBlockDataFunc(item.(*block.BlockStates))
	})

	c.executeTxWorker.Start()
	c.syncAccountToRedisWorker.Start()
	c.updatePoolTxWorker.Start()
	c.preSaveBlockDataWorker.Start()
	c.updateAccountAssetTreeWorker.Start()
	c.updateAccountTreeAndNftTreeWorker.Start()
	c.saveBlockDataWorker.Start()
	c.finalSaveBlockDataWorker.Start()
	c.loadAllAccounts()
	c.loadAllNfts()
	c.pullPoolTxs()
}

func (c *Committer) PendingTxNum() {
	txStatuses := []int64{tx.StatusPending}
	pendingTxCount, _ := c.bc.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
	pendingTxNumMetrics.Set(float64(pendingTxCount))
}

func (c *Committer) pullPoolTxs() {
	executedTx, err := c.bc.TxPoolModel.GetLatestExecutedTx()
	if err != nil && err != types.DbErrNotFound {
		logx.Errorf("get executed tx from tx pool failed:%s", err.Error())
		panic("get executed tx from tx pool failed: " + err.Error())
	}
	var executedTxMaxId uint = 0
	if executedTx != nil {
		executedTxMaxId = executedTx.ID
	}
	limit := 1000
	for {
		if !c.running {
			break
		}
		start := time.Now()
		//logx.Infof("get pool txs executedTxMaxId=%d", executedTxMaxId)
		pendingTxs, err := c.bc.TxPoolModel.GetTxsByStatusAndMaxId(tx.StatusPending, executedTxMaxId, int64(limit))
		getPendingPoolTxMetrics.Set(float64(time.Since(start).Milliseconds()))
		getPendingTxsToQueueMetric.Set(float64(len(pendingTxs)))
		txWorkerQueueMetric.Set(float64(c.executeTxWorker.GetQueueSize()))
		if err != nil {
			logx.Errorf("get pending transactions from tx pool failed:%s", err.Error())
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if len(pendingTxs) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		limit = 1000
		for _, poolTx := range pendingTxs {
			if int(poolTx.ID)-int(executedTxMaxId) != 1 {
				if time.Now().Sub(poolTx.CreatedAt).Seconds() < 5 {
					limit = 10
					time.Sleep(50 * time.Millisecond)
					logx.Infof("not equal id=%s,but delay seconds<5,break it", poolTx.ID)
					break
				} else {
					logx.Infof("not equal id=%s,but delay seconds>5,do it", poolTx.ID)
				}
			}
			executedTxMaxId = poolTx.ID
			c.executeTxWorker.Enqueue(poolTx)
		}
	}
}

func (c *Committer) getPoolTxsFromQueue() []*tx.Tx {
	pendingUpdatePoolTxs := make([]*tx.Tx, 0, 300)
	for {
		select {
		case i := <-c.executeTxWorker.GetJobQueue():
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
	pendingUpdatePoolTxs := make([]*tx.Tx, 0, c.maxTxsPerBlock)
	for {
		curBlock := c.bc.CurrentBlock()
		if curBlock.BlockStatus > block.StatusProposing {
			previousHeight := curBlock.BlockHeight
			curBlock, err = c.bc.InitNewBlock()
			logx.Infof("1 init new block, current height=%s,previous height=%s,blockId=%s", curBlock.BlockHeight, previousHeight, curBlock.ID)
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
			if len(pendingUpdatePoolTxs) > 0 {
				c.enqueueUpdatePoolTx(pendingUpdatePoolTxs, nil)
				pendingUpdatePoolTxs = make([]*tx.Tx, 0, c.maxTxsPerBlock)
			}

			time.Sleep(100 * time.Millisecond)
			pendingTxs = c.getPoolTxsFromQueue()
		}

		pendingDeletePoolTxs := make([]*tx.Tx, 0, len(pendingTxs))
		start := time.Now()

		for _, poolTx := range pendingTxs {
			if c.shouldCommit(curBlock) {
				subPendingTxs = append(subPendingTxs, poolTx)
				continue
			}
			executeTxMetrics.Inc()

			//logx.Infof("apply transaction, txHash=%s", poolTx.TxHash)
			startApplyTx := time.Now()
			err = c.bc.ApplyTransaction(poolTx)
			executeTxApply1TxMetrics.Set(float64(time.Since(startApplyTx).Milliseconds()))
			if err != nil {
				logx.Errorf("apply pool tx ID: %d failed, err %v ", poolTx.ID, err)
				if types.IsPriorityOperationTx(poolTx.TxType) {
					poolTxL1ErrorCountMetics.Inc()
					logx.Errorf("apply priority pool tx failed,id=%s,error=%s", strconv.Itoa(int(poolTx.ID)), err.Error())
					panic("apply priority pool tx failed,id=" + strconv.Itoa(int(poolTx.ID)) + ",error=" + err.Error())
				} else {
					poolTxL2ErrorCountMetics.Inc()
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
					//todo get requestId from pool tx
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
				previousHeight := curBlock.BlockHeight
				if curBlock.ID != 0 {
					err = c.createNewBlock(curBlock)
					logx.Infof("create new block, current height=%s,previous height=%d,blockId=%s", curBlock.BlockHeight, previousHeight, curBlock.ID)
					if err != nil {
						logx.Severef("create new block failed:%s", err.Error())
						panic("create new block failed" + err.Error())
					}
				} else {
					logx.Infof("not create new block,use old block data, current height=%s,previous height=%d,blockId=%s", curBlock.BlockHeight, previousHeight, curBlock.ID)
				}
			}
			pendingUpdatePoolTxs = append(pendingUpdatePoolTxs, poolTx)
		}
		executeTxOperationMetrics.Set(float64(time.Since(start).Milliseconds()))

		for _, formatAccount := range c.bc.Statedb.StateCache.PendingAccountMap {
			if formatAccount.AccountIndex != types.GasAccount {
				continue
			}
			for assetId, delta := range c.bc.Statedb.StateCache.PendingGasMap {
				if asset, ok := formatAccount.AssetInfo[assetId]; ok {
					formatAccount.AssetInfo[assetId].Balance = ffmath.Add(asset.Balance, delta)
				} else {
					formatAccount.AssetInfo[assetId] = &types.AccountAsset{
						Balance:                  delta,
						OfferCanceledOrFinalized: types.ZeroBigInt,
					}
				}
				c.bc.Statedb.StateCache.PendingGasMap[assetId] = types.ZeroBigInt
			}
		}
		c.bc.Statedb.SyncPendingAccountToMemoryCache(c.bc.Statedb.PendingAccountMap)
		c.bc.Statedb.SyncPendingNftToMemoryCache(c.bc.Statedb.PendingNftMap)

		c.enqueueSyncAccountToRedis(c.bc.Statedb.PendingAccountMap, c.bc.Statedb.PendingNftMap)
		c.enqueueUpdatePoolTx(nil, pendingDeletePoolTxs)

		if c.shouldCommit(curBlock) {
			start := time.Now()
			logx.Infof("commit new block, height=%d,blockSize=%d", curBlock.BlockHeight, curBlock.BlockSize)
			for accountIndex, _ := range c.bc.Statedb.GetDirtyAccountsAndAssetsMap() {
				_, exist := c.bc.Statedb.StateCache.GetPendingAccount(accountIndex)
				if !exist {
					accountInfo, err := c.bc.Statedb.GetFormatAccount(accountIndex)
					if err != nil {
						logx.Errorf("get account info failed,accountIndex=%s,err=%s ", accountIndex, err.Error())
						panic("get account info failed: " + err.Error())
					}
					c.bc.Statedb.PendingAccountMap[accountIndex] = accountInfo
				}
			}
			for nftIndex, _ := range c.bc.Statedb.StateCache.GetDirtyNftMap() {
				_, exist := c.bc.Statedb.StateCache.GetPendingNft(nftIndex)
				if !exist {
					nftInfo, err := c.bc.Statedb.GetNft(nftIndex)
					if err != nil {
						logx.Errorf("get nft info failed,nftIndex=%s,err=%s ", nftIndex, err.Error())
						panic("get nft info failed: " + err.Error())
					}
					c.bc.Statedb.PendingNftMap[nftIndex] = nftInfo
				}
			}

			addPendingAccounts := make([]*account.Account, 0)
			for _, accountInfo := range c.bc.Statedb.StateCache.PendingAccountMap {
				if accountInfo.AccountId != 0 {
					continue
				}
				newAccount, err := chain.FromFormatAccountInfo(accountInfo)
				if err != nil {
					logx.Errorf("account info format failed:%s ", err.Error())
					panic("account info format failed: " + err.Error())
				}
				newAccount.L2BlockHeight = curBlock.BlockHeight
				addPendingAccounts = append(addPendingAccounts, newAccount)
			}
			if len(addPendingAccounts) != 0 {
				err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
					return c.bc.DB().AccountModel.BatchInsertOrUpdateInTransact(dbTx, addPendingAccounts)
				})
				if err != nil {
					logx.Errorf("account batch insert or update failed:%s ", err.Error())
					panic("account batch insert or update failed: " + err.Error())
				}
				for _, accountInfo := range addPendingAccounts {
					c.bc.Statedb.StateCache.PendingAccountMap[accountInfo.AccountIndex].AccountId = accountInfo.ID
				}
			}

			addPendingNfts := make([]*nft.L2Nft, 0)
			for _, nftInfo := range c.bc.Statedb.StateCache.PendingNftMap {
				nftInfo.L2BlockHeight = curBlock.BlockHeight
				if nftInfo.ID != 0 {
					continue
				}
				addPendingNfts = append(addPendingNfts, nftInfo)
			}
			if len(addPendingNfts) != 0 {
				err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
					return c.bc.DB().L2NftModel.BatchInsertOrUpdateInTransact(dbTx, addPendingNfts)
				})
				if err != nil {
					logx.Errorf("l2nft batch insert or update failed:%s ", err.Error())
					panic("l2nft batch insert or update failed: " + err.Error())
				}
				for _, nftInfo := range addPendingNfts {
					c.bc.Statedb.StateCache.PendingNftMap[nftInfo.NftIndex].ID = nftInfo.ID
				}
			}

			if len(addPendingAccounts) > 0 || len(addPendingNfts) > 0 {
				pendingAccountMap := make(map[int64]*types.AccountInfo, len(addPendingAccounts))
				pendingNftMap := make(map[int64]*nft.L2Nft, len(addPendingNfts))
				for _, accountInfo := range addPendingAccounts {
					pendingAccountMap[accountInfo.AccountIndex] = c.bc.Statedb.StateCache.PendingAccountMap[accountInfo.AccountIndex]
				}
				for _, nftInfo := range addPendingNfts {
					pendingNftMap[nftInfo.NftIndex] = c.bc.Statedb.StateCache.PendingNftMap[nftInfo.NftIndex]
				}
				c.bc.Statedb.SyncPendingAccountToMemoryCache(pendingAccountMap)
				c.bc.Statedb.SyncPendingNftToMemoryCache(pendingNftMap)
				c.enqueueSyncAccountToRedis(pendingAccountMap, pendingNftMap)
			}

			pendingUpdatePoolTxs = make([]*tx.Tx, 0, c.maxTxsPerBlock)
			pendingAccountMap := make(map[int64]*types.AccountInfo, len(c.bc.Statedb.StateCache.PendingAccountMap))
			pendingNftMap := make(map[int64]*nft.L2Nft, len(c.bc.Statedb.StateCache.PendingNftMap))
			for _, accountInfo := range c.bc.Statedb.StateCache.PendingAccountMap {
				pendingAccountMap[accountInfo.AccountIndex] = accountInfo.DeepCopy()
			}
			for _, nftInfo := range c.bc.Statedb.StateCache.PendingNftMap {
				pendingNftMap[nftInfo.NftIndex] = nftInfo.DeepCopy()
			}
			c.bc.Statedb.StateCache.PendingAccountMap = pendingAccountMap
			c.bc.Statedb.StateCache.PendingNftMap = pendingNftMap

			stateDataCopy := &statedb.StateDataCopy{
				StateCache:   c.bc.Statedb.StateCache,
				CurrentBlock: curBlock,
			}
			c.preSaveBlockDataWorker.Enqueue(stateDataCopy)
			accountAssetTreeQueueMetric.Set(float64(c.updateAccountAssetTreeWorker.GetQueueSize()))

			l2BlockMemoryHeightMetric.Set(float64(stateDataCopy.CurrentBlock.BlockHeight))
			previousHeight := stateDataCopy.CurrentBlock.BlockHeight
			curBlock, err = c.bc.InitNewBlock()

			logx.Infof("2 init new block, current height=%d,previous height=%d,blockId=%d", curBlock.BlockHeight, previousHeight, curBlock.ID)
			if err != nil {
				logx.Errorf("propose new block failed:%s ", err.Error())
				panic("propose new block failed: " + err.Error())
			}

			antsPoolGaugeMetric.WithLabelValues("smt-pool-cap").Set(float64(c.bc.Statedb.TreeCtx.RoutinePool().Cap()))
			antsPoolGaugeMetric.WithLabelValues("smt-pool-free").Set(float64(c.bc.Statedb.TreeCtx.RoutinePool().Free()))
			antsPoolGaugeMetric.WithLabelValues("smt-pool-running").Set(float64(c.bc.Statedb.TreeCtx.RoutinePool().Running()))

			antsPoolGaugeMetric.WithLabelValues("committer-pool-cap").Set(float64(gopool.Cap()))
			antsPoolGaugeMetric.WithLabelValues("committer-pool-free").Set(float64(gopool.Free()))
			antsPoolGaugeMetric.WithLabelValues("committer-pool-running").Set(float64(gopool.Running()))

			commitOperationMetics.Set(float64(time.Since(start).Milliseconds()))
		}
	}
}

func (c *Committer) enqueueUpdatePoolTx(pendingUpdatePoolTxs []*tx.Tx, pendingDeletePoolTxs []*tx.Tx) {
	updatePoolTxMap := &UpdatePoolTx{}
	if pendingUpdatePoolTxs != nil {
		updatePoolTxMap.PendingUpdatePoolTxs = make([]*tx.Tx, 0, len(pendingUpdatePoolTxs))
		for _, poolTx := range pendingUpdatePoolTxs {
			updatePoolTxMap.PendingUpdatePoolTxs = append(updatePoolTxMap.PendingUpdatePoolTxs, poolTx.DeepCopy())
		}
	}
	if pendingDeletePoolTxs != nil {
		updatePoolTxMap.PendingDeletePoolTxs = make([]*tx.Tx, 0, len(pendingDeletePoolTxs))
		for _, poolTx := range pendingDeletePoolTxs {
			updatePoolTxMap.PendingDeletePoolTxs = append(updatePoolTxMap.PendingDeletePoolTxs, poolTx.DeepCopy())
		}
	}
	c.updatePoolTxWorker.Enqueue(updatePoolTxMap)
}

func (c *Committer) updatePoolTxFunc(updatePoolTxMap *UpdatePoolTx) {
	start := time.Now()
	if updatePoolTxMap.PendingUpdatePoolTxs != nil {
		length := len(updatePoolTxMap.PendingUpdatePoolTxs)
		if length > 0 {
			ids := make([]uint, 0, length)
			updateNftIndexOrCollectionIdList := make([]*tx.PoolTx, 0)
			for _, pendingUpdatePoolTx := range updatePoolTxMap.PendingUpdatePoolTxs {
				ids = append(ids, pendingUpdatePoolTx.ID)
				if pendingUpdatePoolTx.TxType == types.TxTypeCreateCollection || pendingUpdatePoolTx.TxType == types.TxTypeMintNft {
					updateNftIndexOrCollectionIdList = append(updateNftIndexOrCollectionIdList, &tx.PoolTx{
						Model:        gorm.Model{ID: pendingUpdatePoolTx.ID},
						NftIndex:     pendingUpdatePoolTx.NftIndex,
						CollectionId: pendingUpdatePoolTx.CollectionId,
					})
				}
			}
			if len(updateNftIndexOrCollectionIdList) > 0 {
				err := c.bc.TxPoolModel.BatchUpdateNftIndexOrCollectionId(updateNftIndexOrCollectionIdList)
				if err != nil {
					logx.Error("update tx pool failed:", err)
					return
				}
			}
			err := c.bc.TxPoolModel.UpdateTxsStatusAndHeightByIds(ids, tx.StatusExecuted, updatePoolTxMap.PendingUpdatePoolTxs[0].BlockHeight)
			if err != nil {
				logx.Error("update tx pool failed:", err)
			}
		}
	}
	if updatePoolTxMap.PendingDeletePoolTxs != nil {
		length := len(updatePoolTxMap.PendingDeletePoolTxs)
		if length > 0 {
			poolTxIds := make([]uint, 0, len(updatePoolTxMap.PendingDeletePoolTxs))
			for _, poolTx := range updatePoolTxMap.PendingDeletePoolTxs {
				poolTxIds = append(poolTxIds, poolTx.ID)
			}
			err := c.bc.TxPoolModel.DeleteTxsBatch(poolTxIds, tx.StatusFailed, -1)
			if err != nil {
				logx.Error("update tx pool failed:", err)
			}
		}
	}
	updatePoolTxsMetrics.Set(float64(time.Since(start).Milliseconds()))
}

func (c *Committer) enqueueSyncAccountToRedis(originPendingAccountMap map[int64]*types.AccountInfo, originPendingNftMap map[int64]*nft.L2Nft) {
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
	c.syncAccountToRedisWorker.Enqueue(pendingMap)
}

func (c *Committer) syncAccountToRedisFunc(pendingMap *PendingMap) {
	start := time.Now()
	c.bc.Statedb.SyncPendingAccountToRedis(pendingMap.PendingAccountMap)
	c.bc.Statedb.SyncPendingNftToRedis(pendingMap.PendingNftMap)
	syncAccountToRedisMetrics.Set(float64(time.Since(start).Milliseconds()))
}

func (c *Committer) preSaveBlockDataFunc(stateDataCopy *statedb.StateDataCopy) {
	start := time.Now()
	logx.Infof("preSaveBlockDataFunc start, blockHeight:%d", stateDataCopy.CurrentBlock.BlockHeight)
	accountIndexes := make([]int64, 0, len(stateDataCopy.StateCache.PendingAccountMap))

	for _, accountInfo := range stateDataCopy.StateCache.PendingAccountMap {
		accountIndexes = append(accountIndexes, accountInfo.AccountIndex)
	}
	nftIndexes := make([]int64, 0, len(stateDataCopy.StateCache.PendingNftMap))
	for _, nftInfo := range stateDataCopy.StateCache.PendingNftMap {
		nftIndexes = append(nftIndexes, nftInfo.NftIndex)
	}
	accountIndexesJson, err := json.Marshal(accountIndexes)
	if err != nil {
		logx.Errorf("marshal accountIndexes failed:%s,blockHeight:%s", err, stateDataCopy.CurrentBlock.BlockHeight)
		panic("marshal accountIndexes failed: " + err.Error())
	}
	nftIndexesJson, err := json.Marshal(nftIndexes)
	if err != nil {
		logx.Errorf("marshal nftIndexesJson failed:%s,blockHeight:%s", err, stateDataCopy.CurrentBlock.BlockHeight)
		panic("marshal nftIndexesJson failed: " + err.Error())
	}
	stateDataCopy.CurrentBlock.AccountIndexes = string(accountIndexesJson)
	stateDataCopy.CurrentBlock.NftIndexes = string(nftIndexesJson)
	stateDataCopy.CurrentBlock.BlockStatus = block.StatusPacked
	err = c.bc.DB().BlockModel.PreSaveBlockData(stateDataCopy.CurrentBlock)
	if err != nil {
		logx.Errorf("preSaveBlockDataFunc failed:%s,blockHeight:%s", err, stateDataCopy.CurrentBlock.BlockHeight)
		panic("preSaveBlockDataFunc failed: " + err.Error())
	}
	latestVerifiedBlockNr, err := c.bc.BlockModel.GetLatestVerifiedHeight()
	if err != nil {
		logx.Error("get latest verified height failed: ", err)
		panic("get latest verified height failed:" + err.Error())
	}
	c.bc.Statedb.UpdatePrunedBlockHeight(latestVerifiedBlockNr)

	preSaveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
	c.updateAccountAssetTreeWorker.Enqueue(stateDataCopy)
}

func (c *Committer) updateAccountAssetTreeFunc(stateDataCopy *statedb.StateDataCopy) {
	start := time.Now()
	updateAccountAssetTreeTxMetrics.Add(float64(len(stateDataCopy.StateCache.Txs)))
	logx.Infof("updateAccountAssetTreeFunc blockHeight:%s,blockId:%s", stateDataCopy.CurrentBlock.BlockHeight, stateDataCopy.CurrentBlock.ID)
	blockSize := c.computeCurrentBlockSize(stateDataCopy)
	if blockSize < len(stateDataCopy.StateCache.Txs) {
		panic("block size too small")
	}
	err := c.bc.UpdateAccountAssetTree(stateDataCopy)
	if err != nil {
		logx.Errorf("updateAccountAssetTreeFunc failed:%s,blockHeight:%s", err, stateDataCopy.CurrentBlock.BlockHeight)
		panic("updateAccountAssetTreeFunc failed: " + err.Error())
	}
	c.updateAccountTreeAndNftTreeWorker.Enqueue(stateDataCopy)
	accountTreeAndNftTreeQueueMetric.Set(float64(c.updateAccountTreeAndNftTreeWorker.GetQueueSize()))
	updateAccountAssetTreeMetrics.Set(float64(time.Since(start).Milliseconds()))

}

func (c *Committer) updateAccountTreeAndNftTreeFunc(stateDataCopy *statedb.StateDataCopy) {
	start := time.Now()
	updateAccountTreeAndNftTreeTxMetrics.Add(float64(len(stateDataCopy.StateCache.Txs)))
	logx.Infof("updateAccountTreeAndNftTreeFunc blockHeight:%s,blockId:%s", stateDataCopy.CurrentBlock.BlockHeight, stateDataCopy.CurrentBlock.ID)
	blockSize := c.computeCurrentBlockSize(stateDataCopy)
	blockStates, err := c.bc.UpdateAccountTreeAndNftTree(blockSize, stateDataCopy)
	if err != nil {
		logx.Errorf("updateAccountTreeAndNftTreeFunc failed:%s,blockHeight:%s", err, stateDataCopy.CurrentBlock.BlockHeight)
		panic("updateAccountTreeAndNftTreeFunc failed: " + err.Error())
	}
	c.saveBlockDataWorker.Enqueue(blockStates)
	l2BlockRedisHeightMetric.Set(float64(blockStates.Block.BlockHeight))
	AccountLatestVersionTreeMetric.Set(float64(c.bc.StateDB().AccountTree.LatestVersion()))
	AccountRecentVersionTreeMetric.Set(float64(c.bc.StateDB().AccountTree.RecentVersion()))
	NftTreeLatestVersionMetric.Set(float64(c.bc.StateDB().NftTree.LatestVersion()))
	NftTreeRecentVersionMetric.Set(float64(c.bc.StateDB().NftTree.RecentVersion()))
	updateAccountTreeAndNftTreeMetrics.Set(float64(time.Since(start).Milliseconds()))
}

func (c *Committer) saveBlockDataFunc(blockStates *block.BlockStates) {
	start := time.Now()
	logx.Infof("saveBlockDataFunc start, blockHeight:%d", blockStates.Block.BlockHeight)
	totalTask := 0
	errChan := make(chan error, 1)
	defer close(errChan)
	var err error

	poolTxIds := make([]uint, 0, len(blockStates.Block.Txs))
	updateNftIndexOrCollectionIdList := make([]*tx.PoolTx, 0)

	for _, poolTx := range blockStates.Block.Txs {
		poolTxIds = append(poolTxIds, poolTx.ID)
		if poolTx.TxType == types.TxTypeCreateCollection || poolTx.TxType == types.TxTypeMintNft {
			updateNftIndexOrCollectionIdList = append(updateNftIndexOrCollectionIdList, &tx.PoolTx{
				Model:        gorm.Model{ID: poolTx.ID},
				NftIndex:     poolTx.NftIndex,
				CollectionId: poolTx.CollectionId,
			})
		}
	}

	blockStates.Block.ClearTxsModel()
	totalTask++
	err = func(poolTxIds []uint, blockHeight int64, updateNftIndexOrCollectionIdList []*tx.PoolTx) error {
		return c.pool.Submit(func() {
			start := time.Now()
			if len(updateNftIndexOrCollectionIdList) > 0 {
				err := c.bc.TxPoolModel.BatchUpdateNftIndexOrCollectionId(updateNftIndexOrCollectionIdList)
				if err != nil {
					logx.Error("update tx pool failed:", err)
					errChan <- err
					return
				}
			}

			err = c.bc.DB().TxPoolModel.DeleteTxsBatch(poolTxIds, tx.StatusExecuted, blockHeight)
			deletePoolTxMetrics.Set(float64(time.Since(start).Milliseconds()))
			if err != nil {
				errChan <- err
				return
			}
			errChan <- nil
		})
	}(poolTxIds, blockStates.Block.BlockHeight, updateNftIndexOrCollectionIdList)
	if err != nil {
		panic("DeleteTxsBatch failed: " + err.Error())
	}

	pendingAccountLen := len(blockStates.PendingAccount)
	if pendingAccountLen > 0 {
		sort.SliceStable(blockStates.PendingAccount, func(i, j int) bool {
			return blockStates.PendingAccount[i].AccountIndex < blockStates.PendingAccount[j].AccountIndex
		})
		fromIndex := 0
		limit := 100
		toIndex := limit
		for {
			if fromIndex >= pendingAccountLen {
				break
			}
			if toIndex > pendingAccountLen {
				toIndex = pendingAccountLen
			}
			accounts := blockStates.PendingAccount[fromIndex:toIndex]
			fromIndex = toIndex
			toIndex += limit

			totalTask++
			err := func(accounts []*account.Account) error {
				return c.pool.Submit(func() {
					start := time.Now()
					err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
						return c.bc.DB().AccountModel.BatchInsertOrUpdateInTransact(dbTx, accounts)
					})
					saveAccountsGoroutineMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(accounts)
			if err != nil {
				panic("batchInsertOrUpdate accounts failed: " + err.Error())
			}
		}
	}
	pendingAccountHistoryLen := len(blockStates.PendingAccountHistory)
	if pendingAccountHistoryLen > 0 {
		fromIndex := 0
		limit := 100
		toIndex := limit
		for {
			if fromIndex >= pendingAccountHistoryLen {
				break
			}
			if toIndex > pendingAccountHistoryLen {
				toIndex = pendingAccountHistoryLen
			}
			accountHistories := blockStates.PendingAccountHistory[fromIndex:toIndex]
			fromIndex = toIndex
			toIndex += limit

			totalTask++
			err := func(accountHistories []*account.AccountHistory) error {
				return c.pool.Submit(func() {
					start := time.Now()
					err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
						return c.bc.DB().AccountHistoryModel.CreateAccountHistoriesInTransact(dbTx, accountHistories)
					})
					addAccountHistoryMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(accountHistories)
			if err != nil {
				panic("createAccountHistories failed: " + err.Error())
			}
		}
	}

	pendingNftLen := len(blockStates.PendingNft)
	if pendingNftLen > 0 {
		sort.SliceStable(blockStates.PendingNft, func(i, j int) bool {
			return blockStates.PendingNft[i].NftIndex < blockStates.PendingNft[j].NftIndex
		})
		fromIndex := 0
		limit := 100
		toIndex := limit
		for {
			if fromIndex >= pendingNftLen {
				break
			}
			if toIndex > pendingNftLen {
				toIndex = pendingNftLen
			}
			nfts := blockStates.PendingNft[fromIndex:toIndex]
			fromIndex = toIndex
			toIndex += limit

			totalTask++
			err := func(nfts []*nft.L2Nft) error {
				return c.pool.Submit(func() {
					start := time.Now()
					err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
						return c.bc.DB().L2NftModel.BatchInsertOrUpdateInTransact(dbTx, nfts)
					})
					saveAccountsGoroutineMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(nfts)
			if err != nil {
				panic("batchInsertOrUpdate nfts failed: " + err.Error())
			}
		}
	}
	pendingNftHistoryLen := len(blockStates.PendingNftHistory)
	if pendingNftHistoryLen > 0 {
		fromIndex := 0
		limit := 100
		toIndex := limit
		for {
			if fromIndex >= pendingNftHistoryLen {
				break
			}
			if toIndex > pendingNftHistoryLen {
				toIndex = pendingNftHistoryLen
			}
			nftHistories := blockStates.PendingNftHistory[fromIndex:toIndex]
			fromIndex = toIndex
			toIndex += limit

			totalTask++
			err := func(nftHistories []*nft.L2NftHistory) error {
				return c.pool.Submit(func() {
					start := time.Now()
					err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
						return c.bc.DB().L2NftHistoryModel.CreateNftHistoriesInTransact(dbTx, nftHistories)
					})
					addAccountHistoryMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(nftHistories)
			if err != nil {
				panic("createNftHistories failed: " + err.Error())
			}
		}
	}

	txsLen := len(blockStates.Block.Txs)
	if txsLen > 0 {
		fromIndex := 0
		limit := 100
		toIndex := limit
		for {
			if fromIndex >= txsLen {
				break
			}
			if toIndex > txsLen {
				toIndex = txsLen
			}
			txs := blockStates.Block.Txs[fromIndex:toIndex]
			fromIndex = toIndex
			toIndex += limit
			totalTask++
			err := func(txs []*tx.Tx) error {
				return c.pool.Submit(func() {
					start := time.Now()
					err = c.bc.DB().TxModel.CreateTxs(txs)
					addTxsMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(txs)
			if err != nil {
				panic("CreateTxs failed: " + err.Error())
			}
		}

		txDetails := make([]*tx.TxDetail, 0)
		for _, txInfo := range blockStates.Block.Txs {
			txDetails = append(txDetails, txInfo.TxDetails...)
		}
		txDetailsLen := len(txDetails)
		if txDetailsLen > 0 {
			fromIndex := 0
			limit := 100
			toIndex := limit
			for {
				if fromIndex >= txDetailsLen {
					break
				}
				if toIndex > txDetailsLen {
					toIndex = txDetailsLen
				}
				txDetailsSlice := txDetails[fromIndex:toIndex]
				fromIndex = toIndex
				toIndex += limit
				totalTask++
				err := func(txDetails []*tx.TxDetail) error {
					return c.pool.Submit(func() {
						start := time.Now()
						err = c.bc.DB().TxDetailModel.CreateTxDetails(txDetails)
						addTxDetailsMetrics.Set(float64(time.Since(start).Milliseconds()))
						if err != nil {
							errChan <- err
							return
						}
						errChan <- nil
					})
				}(txDetailsSlice)
				if err != nil {
					panic("CreateTxDetails failed: " + err.Error())
				}
			}
		}
	}
	for i := 0; i < totalTask; i++ {
		err := <-errChan
		if err != nil {
			panic("saveBlockDataFunc failed: " + err.Error())
		}
	}

	saveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
	c.finalSaveBlockDataWorker.Enqueue(blockStates)
}

func (c *Committer) finalSaveBlockDataFunc(blockStates *block.BlockStates) {
	start := time.Now()
	logx.Infof("finalSaveBlockDataFunc start, blockHeight:%d", blockStates.Block.BlockHeight)
	// update db
	err := c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
		if blockStates.CompressedBlock != nil {
			start := time.Now()
			err := c.bc.DB().CompressedBlockModel.CreateCompressedBlockInTransact(tx, blockStates.CompressedBlock)
			finalSaveBlockDataMetrics.WithLabelValues("add_compressed_block").Set(float64(time.Since(start).Milliseconds()))
			if err != nil {
				return err
			}
		}
		start := time.Now()
		err := c.bc.DB().BlockModel.UpdateBlockToPendingInTransact(tx, blockStates.Block)
		if err != nil {
			return err
		}
		finalSaveBlockDataMetrics.WithLabelValues("update_block_to_pending").Set(float64(time.Since(start).Milliseconds()))
		return err
	})
	if err != nil {
		logx.Errorf("finalSaveBlockDataFunc failed:%s,blockHeight:%d", err.Error(), blockStates.Block.BlockHeight)
		panic("finalSaveBlockDataFunc failed: " + err.Error())
	}
	l2BlockDbHeightMetric.Set(float64(blockStates.Block.BlockHeight))
	finalSaveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
}

func (c *Committer) createNewBlock(curBlock *block.Block) error {
	return c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
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

func (c *Committer) loadAllAccounts() {
	limit := int64(1000)
	offset := int64(0)
	for {
		accounts, err := c.bc.AccountModel.GetUsers(limit, offset)
		if err != nil {
			logx.Errorf("load all accounts failed:%s", err.Error())
			panic("load all accounts failed: " + err.Error())
		}
		if accounts == nil {
			return
		}
		for _, accountInfo := range accounts {
			offset++
			formatAccount, err := chain.ToFormatAccountInfo(accountInfo)
			if err != nil {
				logx.Errorf("load all accounts failed:%s", err.Error())
				panic("load all accounts failed: " + err.Error())
			}
			c.bc.Statedb.AccountCache.Add(accountInfo.AccountIndex, formatAccount)
		}
	}
}

func (c *Committer) loadAllNfts() {
	limit := int64(1000)
	offset := int64(0)
	for {
		nfts, err := c.bc.L2NftModel.GetNfts(limit, offset)
		if err != nil {
			logx.Errorf("load all nfts failed:%s", err.Error())
			panic("load all nfts failed: " + err.Error())
		}
		if nfts == nil {
			return
		}
		for _, nftInfo := range nfts {
			offset++
			c.bc.Statedb.NftCache.Add(nftInfo.NftIndex, nftInfo)
		}
	}
}

func (c *Committer) Shutdown() {
	c.running = false
	c.executeTxWorker.Stop()
	c.syncAccountToRedisWorker.Stop()
	c.updatePoolTxWorker.Stop()
	c.updateAccountAssetTreeWorker.Stop()
	c.updateAccountTreeAndNftTreeWorker.Stop()
	c.preSaveBlockDataWorker.Stop()
	c.saveBlockDataWorker.Stop()
	c.finalSaveBlockDataWorker.Stop()
	c.bc.Statedb.Close()
	c.bc.ChainDB.Close()
}
