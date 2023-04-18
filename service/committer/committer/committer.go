package committer

import (
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/gopool"
	"github.com/bnb-chain/zkbnb/common/log"
	"github.com/bnb-chain/zkbnb/common/metrics"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/panjf2000/ants/v2"
	"sort"
	"strconv"
	"time"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/committer/config"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	MaxPackedInterval = 60 * 1
)

type Committer struct {
	running              bool
	config               *config.Config
	maxTxsPerBlock       int
	maxCommitterInterval int
	optionalBlockSizes   []int

	bc                            *core.BlockChain
	executeTxWorker               *core.TxWorker
	updatePoolTxWorker            *core.Worker
	syncAccountToRedisWorker      *core.Worker
	updateAssetTreeWorker         *core.Worker
	updateAccountAndNftTreeWorker *core.Worker
	preSaveBlockDataWorker        *core.Worker
	saveBlockDataWorker           *core.Worker
	finalSaveBlockDataWorker      *core.Worker
	pool                          *ants.Pool
}

type PendingMap struct {
	PendingAccountMap map[int64]*types.AccountInfo
	PendingNftMap     map[int64]*nft.L2Nft
}
type UpdatePoolTx struct {
	PendingUpdatePoolTxs []*tx.Tx
	PendingDeletePoolTxs []*tx.Tx
}

//work flow: pullPoolTxsToQueue->executeTxFunc->updatePoolTxFunc->preSaveBlockDataFunc
//->updateAssetTreeFunc->updateAccountAndNftTreeFunc->saveBlockDataFunc->finalSaveBlockDataFunc

func NewCommitter(config *config.Config) (*Committer, error) {
	if len(config.BlockConfig.OptionalBlockSizes) == 0 {
		return nil, types.AppErrNilOptionalBlockSize
	}

	err := metrics.InitCommitterMetrics()
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
	if config.BlockConfig.MaxPackedInterval == 0 {
		config.BlockConfig.MaxPackedInterval = MaxPackedInterval
	}
	pool, err := ants.NewPool(saveBlockDataPoolSize, ants.WithPanicHandler(func(p interface{}) {
		//sets up panic handler.
		panic("worker exits from a panic")
	}))
	common.NewIPFS(config.IpfsUrl)
	committer := &Committer{
		running:              true,
		config:               config,
		maxTxsPerBlock:       config.BlockConfig.OptionalBlockSizes[len(config.BlockConfig.OptionalBlockSizes)-1],
		maxCommitterInterval: config.BlockConfig.MaxPackedInterval,
		optionalBlockSizes:   config.BlockConfig.OptionalBlockSizes,
		bc:                   bc,
		pool:                 pool,
	}

	return committer, nil
}

func (c *Committer) Run() error {
	//rollback only
	if c.config.BlockConfig.RollbackOnly {
		for {
			if !c.running {
				break
			}
			logx.Info("do rollback only")
			time.Sleep(1 * time.Minute)
		}
		return nil
	}

	//execute tx,generate a block
	c.executeTxWorker = core.ExecuteTxWorker(10000, func() error {
		return c.executeTxFunc()
	})

	//update pool tx status
	c.updatePoolTxWorker = core.UpdatePoolTxWorker(10000, func(item interface{}) error {
		return c.updatePoolTxFunc(item.(*UpdatePoolTx))
	})
	c.syncAccountToRedisWorker = core.SyncAccountToRedisWorker(10000, func(item interface{}) error {
		return c.syncAccountToRedisFunc(item.(*PendingMap))
	})
	c.preSaveBlockDataWorker = core.PreSaveBlockDataWorker(10, func(item interface{}) error {
		return c.preSaveBlockDataFunc(item.(*statedb.StateDataCopy))
	})
	c.updateAssetTreeWorker = core.UpdateAssetTreeWorker(10, func(item interface{}) error {
		return c.updateAssetTreeFunc(item.(*statedb.StateDataCopy))
	})
	c.updateAccountAndNftTreeWorker = core.UpdateAccountAndNftTreeWorker(10, func(item interface{}) error {
		return c.updateAccountAndNftTreeFunc(item.(*statedb.StateDataCopy))
	})
	c.saveBlockDataWorker = core.SaveBlockDataWorker(10, func(item interface{}) error {
		return c.saveBlockDataFunc(item.(*block.BlockStates))
	})
	c.finalSaveBlockDataWorker = core.FinalSaveBlockDataWorker(10, func(item interface{}) error {
		return c.finalSaveBlockDataFunc(item.(*block.BlockStates))
	})

	//load accounts from db to memcache
	err := c.bc.LoadAllAccounts(c.pool)
	if err != nil {
		return err
	}

	//load nfts from db to memcache
	err = c.bc.LoadAllNfts(c.pool)
	if err != nil {
		return err
	}

	c.executeTxWorker.Start()
	c.syncAccountToRedisWorker.Start()
	c.updatePoolTxWorker.Start()
	c.preSaveBlockDataWorker.Start()
	c.updateAssetTreeWorker.Start()
	c.updateAccountAndNftTreeWorker.Start()
	c.saveBlockDataWorker.Start()
	c.finalSaveBlockDataWorker.Start()

	//pull pool txs from db to queue
	err = c.pullPoolTxsToQueue()
	if err != nil {
		return err
	}
	return nil
}

// pull pool txs from db to queue
func (c *Committer) pullPoolTxsToQueue() error {
	executedTx, err := c.bc.TxPoolModel.GetLatestExecutedTx()
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get executed tx from tx pool failed:%s", err.Error())
	}

	var executedTxMaxId uint = 0
	if executedTx != nil {
		executedTxMaxId = executedTx.ID
	}
	limit := 1000
	pendingTxs := make([]*tx.Tx, 0)

	for {
		if !c.running {
			break
		}
		start := time.Now()
		pendingTxs, err = c.bc.TxPoolModel.GetTxsByStatusAndMaxId(tx.StatusPending, executedTxMaxId, int64(limit))
		if err != nil {
			logx.Severef("get pending transactions from tx pool failed:%s", err.Error())
			time.Sleep(500 * time.Millisecond)
			continue
		}
		metrics.GetPendingPoolTxMetrics.Set(float64(time.Since(start).Milliseconds()))
		metrics.GetPendingTxsToQueueMetric.Set(float64(len(pendingTxs)))
		metrics.TxWorkerQueueMetric.Set(float64(c.executeTxWorker.GetQueueSize()))

		if len(pendingTxs) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		limit = 1000
		for _, poolTx := range pendingTxs {
			if int(poolTx.ID)-int(executedTxMaxId) != 1 {
				// Ensure orderly execution of pool tx. In high concurrency scenarios, the ids obtained from the database are not continuous
				// Wait for while, the data is successfully written to the database, and then re-get the data
				if time.Now().Sub(poolTx.CreatedAt).Seconds() < 5 {
					limit = 10
					time.Sleep(50 * time.Millisecond)
					logx.Infof("not equal id=%d,but delay seconds<5,break it", poolTx.ID)
					break
				} else {
					//If the time is greater than 5 seconds, skip this id and compensate through CompensatePendingPoolTx
					logx.Infof("not equal id=%d,but delay seconds>5,do it", poolTx.ID)
				}
			}
			executedTxMaxId = poolTx.ID
			c.executeTxWorker.Enqueue(poolTx)
		}
	}
	return nil
}

// get pool txs from queue
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

// execute tx,generate a block
func (c *Committer) executeTxFunc() error {
	l1LatestRequestId, err := c.getLatestExecutedRequestId()
	if err != nil {
		return fmt.Errorf("get latest executed request id failed:%s", err.Error())
	}

	var subPendingTxs []*tx.Tx
	var pendingTxs []*tx.Tx
	pendingUpdatePoolTxs := make([]*tx.Tx, 0, c.maxTxsPerBlock)
	for {
		curBlock := c.bc.CurrentBlock()
		ctx := log.NewCtxWithKV(log.BlockHeightContext, curBlock.BlockHeight)
		if curBlock.BlockStatus > block.StatusProposing {
			previousHeight := curBlock.BlockHeight
			curBlock, err = c.bc.InitNewBlock()
			if err != nil {
				return fmt.Errorf("propose new block failed: %s", err.Error())
			}
			ctx := log.UpdateCtxWithKV(ctx, log.BlockHeightContext, curBlock.BlockHeight)
			logx.WithContext(ctx).Infof("1 init new block, current height=%d,previous height=%d,blockId=%d", curBlock.BlockHeight, previousHeight, curBlock.ID)
		}

		if subPendingTxs != nil && len(subPendingTxs) > 0 {
			pendingTxs = subPendingTxs
			subPendingTxs = nil
		} else {
			pendingTxs = c.getPoolTxsFromQueue()
			//c.preLoadAccountAndNft(pendingTxs)
		}

		for len(pendingTxs) == 0 {
			if c.shouldCommit(curBlock) {
				break
			}
			if len(pendingUpdatePoolTxs) > 0 {
				c.addUpdatePoolTxToQueue(pendingUpdatePoolTxs, nil)
				pendingUpdatePoolTxs = make([]*tx.Tx, 0, c.maxTxsPerBlock)
			}

			time.Sleep(100 * time.Millisecond)
			pendingTxs = c.getPoolTxsFromQueue()
			//c.preLoadAccountAndNft(pendingTxs)
		}

		pendingDeletePoolTxs := make([]*tx.Tx, 0, len(pendingTxs))
		start := time.Now()

		for _, poolTx := range pendingTxs {
			if c.shouldCommit(curBlock) {
				subPendingTxs = append(subPendingTxs, poolTx)
				continue
			}
			ctx := log.UpdateCtxWithKV(ctx, log.PoolTxIdContext, poolTx.ID)

			metrics.ExecuteTxMetrics.Inc()
			startApplyTx := time.Now()
			logx.WithContext(ctx).Infof("start apply pool tx ID: %d", poolTx.ID)
			err = c.bc.ApplyTransaction(poolTx)
			if c.bc.Statedb.NeedRestoreExecutedTxs() && poolTx.ID >= c.bc.Statedb.MaxPollTxIdRollbackImmutable {
				logx.WithContext(ctx).Infof("update needRestoreExecutedTxs to false,blockHeight:%d", curBlock.BlockHeight)
				c.bc.Statedb.UpdateNeedRestoreExecutedTxs(false)
				err := c.bc.DB().BlockModel.DeleteBlockGreaterThanHeight(curBlock.BlockHeight, []int{block.StatusProposing, block.StatusPacked})
				if err != nil {
					return fmt.Errorf("DeleteBlockGreaterThanHeight failed:%s,blockHeight:%d", err.Error(), curBlock.BlockHeight)
				}
				c.bc.ClearRollbackBlockMap()
			}
			metrics.ExecuteTxApply1TxMetrics.Set(float64(time.Since(startApplyTx).Milliseconds()))
			if err != nil {
				logx.Severef("apply pool tx failed,id=%d, err %v ", poolTx.ID, err)
				if types.IsPriorityOperationTx(poolTx.TxType) {
					metrics.PoolTxL1ErrorCountMetics.Inc()
					return fmt.Errorf("apply priority pool tx failed,id=%d,error=%s", poolTx.ID, err.Error())
				} else {
					expectNonce, err := c.bc.Statedb.GetCommittedNonce(poolTx.AccountIndex)
					if err != nil {
						c.bc.Statedb.ClearPendingNonceFromRedisCache(poolTx.AccountIndex)
					} else {
						if poolTx.IsNonceChanged {
							expectNonce = expectNonce - 1
						}
						c.bc.Statedb.SetPendingNonceToRedisCache(poolTx.AccountIndex, expectNonce-1)
					}
					metrics.PoolTxL2ErrorCountMetics.Inc()
				}
				poolTx.TxStatus = tx.StatusFailed
				pendingDeletePoolTxs = append(pendingDeletePoolTxs, poolTx)
				continue
			}

			if types.IsPriorityOperationTx(poolTx.TxType) {
				metrics.PriorityOperationMetric.Set(float64(poolTx.L1RequestId))
				if l1LatestRequestId != -1 && poolTx.L1RequestId != l1LatestRequestId+1 {
					return fmt.Errorf("invalid request id=%d", poolTx.L1RequestId)
				}
				l1LatestRequestId = poolTx.L1RequestId
			}

			// Write the proposed block into database when the first transaction executed.
			if len(c.bc.Statedb.Txs) == 1 {
				previousHeight := curBlock.BlockHeight
				if curBlock.ID == 0 {
					err = c.createNewBlock(curBlock)
					logx.WithContext(ctx).Infof("create new block, current height=%d,previous height=%d,blockId=%d", curBlock.BlockHeight, previousHeight, curBlock.ID)
					if err != nil {
						return fmt.Errorf("create new block failed:%s", err.Error())
					}
				} else {
					logx.WithContext(ctx).Infof("not create new block,use old block data, current height=%d,previous height=%d,blockId=%d", curBlock.BlockHeight, previousHeight, curBlock.ID)
				}
			}
			pendingUpdatePoolTxs = append(pendingUpdatePoolTxs, poolTx)
		}

		metrics.ExecuteTxOperationMetrics.Set(float64(time.Since(start).Milliseconds()))

		c.bc.Statedb.SyncPendingAccountToMemoryCache(c.bc.Statedb.PendingAccountMap)
		c.bc.Statedb.SyncPendingNftToMemoryCache(c.bc.Statedb.PendingNftMap)

		c.addSyncAccountToRedisToQueue(c.bc.Statedb.PendingAccountMap, c.bc.Statedb.PendingNftMap)
		c.addUpdatePoolTxToQueue(nil, pendingDeletePoolTxs)

		if c.shouldCommit(curBlock) {
			start := time.Now()
			logx.WithContext(ctx).Infof("commit new block, height=%d,blockSize=%d", curBlock.BlockHeight, curBlock.BlockSize)
			pendingUpdatePoolTxs = make([]*tx.Tx, 0, c.maxTxsPerBlock)
			stateDataCopy, err := c.buildStateDataCopy(curBlock)
			if err != nil {
				return err
			}

			c.preSaveBlockDataWorker.Enqueue(stateDataCopy)
			metrics.AccountAssetTreeQueueMetric.Set(float64(c.updateAssetTreeWorker.GetQueueSize()))

			metrics.L2BlockMemoryHeightMetric.Set(float64(stateDataCopy.CurrentBlock.BlockHeight))
			previousHeight := stateDataCopy.CurrentBlock.BlockHeight
			curBlock, err = c.bc.InitNewBlock()
			if err != nil {
				return fmt.Errorf("propose new block failed: %s", err.Error())
			}
			logx.WithContext(ctx).Infof("2 init new block, current height=%d,previous height=%d,blockId=%d", curBlock.BlockHeight, previousHeight, curBlock.ID)

			metrics.AntsPoolGaugeMetric.WithLabelValues("smt-pool-cap").Set(float64(c.bc.Statedb.TreeCtx.RoutinePool().Cap()))
			metrics.AntsPoolGaugeMetric.WithLabelValues("smt-pool-free").Set(float64(c.bc.Statedb.TreeCtx.RoutinePool().Free()))
			metrics.AntsPoolGaugeMetric.WithLabelValues("smt-pool-running").Set(float64(c.bc.Statedb.TreeCtx.RoutinePool().Running()))

			metrics.AntsPoolGaugeMetric.WithLabelValues("committer-pool-cap").Set(float64(gopool.Cap()))
			metrics.AntsPoolGaugeMetric.WithLabelValues("committer-pool-free").Set(float64(gopool.Free()))
			metrics.AntsPoolGaugeMetric.WithLabelValues("committer-pool-running").Set(float64(gopool.Running()))

			metrics.CommitOperationMetics.Set(float64(time.Since(start).Milliseconds()))
		}
	}
}

// copy state cache
func (c *Committer) buildStateDataCopy(curBlock *block.Block) (*statedb.StateDataCopy, error) {
	gasAccount := c.bc.Statedb.StateCache.PendingAccountMap[types.GasAccount]
	if gasAccount != nil {
		if len(c.bc.Statedb.StateCache.PendingGasMap) != 0 {
			for assetId, delta := range c.bc.Statedb.StateCache.PendingGasMap {
				if asset, ok := gasAccount.AssetInfo[assetId]; ok {
					gasAccount.AssetInfo[assetId].Balance = ffmath.Add(asset.Balance, delta)
				} else {
					gasAccount.AssetInfo[assetId] = &types.AccountAsset{
						Balance:                  delta,
						OfferCanceledOrFinalized: types.ZeroBigInt,
					}
				}
				c.bc.Statedb.MarkAccountAssetsDirty(gasAccount.AccountIndex, []int64{assetId})
			}
		} else {
			assetsMap := c.bc.Statedb.GetDirtyAccountsAndAssetsMap()[gasAccount.AccountIndex]
			if assetsMap == nil {
				delete(c.bc.Statedb.StateCache.PendingAccountMap, types.GasAccount)
			}
		}
	}

	for _, formatAccount := range c.bc.Statedb.StateCache.PendingAccountMap {
		assetsMap := c.bc.Statedb.GetDirtyAccountsAndAssetsMap()[formatAccount.AccountIndex]
		if assetsMap == nil {
			return nil, fmt.Errorf("%d exists in PendingAccountMap but not in GetDirtyAccountsAndAssetsMap", formatAccount.AccountIndex)
		}
	}

	for accountIndex := range c.bc.Statedb.GetDirtyAccountsAndAssetsMap() {
		_, exist := c.bc.Statedb.StateCache.GetPendingAccount(accountIndex)
		if !exist {
			accountInfo, err := c.bc.Statedb.GetFormatAccount(accountIndex)
			if err != nil {
				return nil, fmt.Errorf("get account info failed,accountIndex=%d,err=%s ", accountIndex, err.Error())
			}
			c.bc.Statedb.SetPendingAccount(accountIndex, accountInfo)
		}
	}

	for _, nftInfo := range c.bc.Statedb.StateCache.PendingNftMap {
		if c.bc.Statedb.GetDirtyNftMap()[nftInfo.NftIndex] == false {
			return nil, fmt.Errorf(strconv.FormatInt(nftInfo.NftIndex, 10) + " exists in PendingNftMap but not in DirtyNftMap")
		}
	}
	for nftIndex := range c.bc.Statedb.StateCache.GetDirtyNftMap() {
		_, exist := c.bc.Statedb.StateCache.GetPendingNft(nftIndex)
		if !exist {
			nftInfo, err := c.bc.Statedb.GetNft(nftIndex)
			if err != nil {
				return nil, fmt.Errorf("get nft info failed,nftIndex=%d,err=%s ", nftIndex, err.Error())
			}
			c.bc.Statedb.SetPendingNft(nftIndex, nftInfo)
		}
	}

	addPendingAccounts := make([]*account.Account, 0)
	for _, accountInfo := range c.bc.Statedb.StateCache.PendingAccountMap {
		if accountInfo.AccountId != 0 {
			continue
		}
		newAccount, err := chain.FromFormatAccountInfo(accountInfo)
		if err != nil {
			return nil, fmt.Errorf("account info format failed: %s", err.Error())
		}
		newAccount.L2BlockHeight = curBlock.BlockHeight
		addPendingAccounts = append(addPendingAccounts, newAccount)
	}

	addPendingNfts := make([]*nft.L2Nft, 0)
	for _, nftInfo := range c.bc.Statedb.StateCache.PendingNftMap {
		nftInfo.L2BlockHeight = curBlock.BlockHeight
		if nftInfo.ID != 0 {
			continue
		}
		addPendingNfts = append(addPendingNfts, nftInfo)
	}

	updateAccountMap := make(map[int64]*types.AccountInfo, 0)
	updateNftMap := make(map[int64]*nft.L2Nft, 0)
	if len(addPendingAccounts) > 0 || len(addPendingNfts) > 0 {
		err := c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
			if len(addPendingAccounts) != 0 {
				err := c.bc.DB().AccountModel.BatchInsertInTransact(dbTx, addPendingAccounts)
				if err != nil {
					return fmt.Errorf("account batch insert or update failed: %s", err.Error())
				}
			}

			if len(addPendingNfts) != 0 {
				err := c.bc.DB().L2NftModel.BatchInsertInTransact(dbTx, addPendingNfts)
				if err != nil {
					return fmt.Errorf("l2nft batch insert or update failed: %s", err.Error())
				}
			}

			accountIndexes := make([]int64, 0, len(addPendingAccounts))
			for _, accountInfo := range addPendingAccounts {
				accountIndexes = append(accountIndexes, accountInfo.AccountIndex)
			}

			nftIndexes := make([]int64, 0, len(addPendingNfts))
			for _, nftInfo := range addPendingNfts {
				nftIndexes = append(nftIndexes, nftInfo.NftIndex)
			}

			accountIndexesJson, err := json.Marshal(accountIndexes)
			if err != nil {
				return fmt.Errorf("marshal accountIndexes failed:%s,blockHeight:%d", err, curBlock.BlockHeight)
			}
			nftIndexesJson, err := json.Marshal(nftIndexes)
			if err != nil {
				return fmt.Errorf("marshal nftIndexesJson failed:%s,blockHeight:%d", err, curBlock.BlockHeight)
			}

			curBlock.AccountIndexes = string(accountIndexesJson)
			curBlock.NftIndexes = string(nftIndexesJson)
			curBlock.BlockStatus = block.StatusPacked
			err = c.bc.DB().BlockModel.PreSaveBlockDataInTransact(dbTx, curBlock)
			if err != nil {
				return fmt.Errorf("PreSaveBlockDataInTransact failed:%s,blockHeight:%d", err, curBlock.BlockHeight)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("account or nft insert failed: %s", err.Error())
		}

		for _, accountInfo := range addPendingAccounts {
			c.bc.Statedb.StateCache.PendingAccountMap[accountInfo.AccountIndex].AccountId = int64(accountInfo.ID)
			updateAccountMap[accountInfo.AccountIndex] = c.bc.Statedb.StateCache.PendingAccountMap[accountInfo.AccountIndex]
		}
		for _, nftInfo := range addPendingNfts {
			c.bc.Statedb.StateCache.PendingNftMap[nftInfo.NftIndex].ID = nftInfo.ID
			updateNftMap[nftInfo.NftIndex] = c.bc.Statedb.StateCache.PendingNftMap[nftInfo.NftIndex]
		}
	}
	gasAccount = c.bc.Statedb.StateCache.PendingAccountMap[types.GasAccount]
	if gasAccount != nil {
		updateAccountMap[types.GasAccount] = gasAccount
	}
	c.bc.Statedb.SyncPendingAccountToMemoryCache(updateAccountMap)
	c.bc.Statedb.SyncPendingNftToMemoryCache(updateNftMap)
	c.addSyncAccountToRedisToQueue(updateAccountMap, updateNftMap)

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
	return stateDataCopy, nil
}

// put the pool txs that need to be updated into the queue
func (c *Committer) addUpdatePoolTxToQueue(pendingUpdatePoolTxs []*tx.Tx, pendingDeletePoolTxs []*tx.Tx) {
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

// update pool tx to StatusExecuted
func (c *Committer) updatePoolTxFunc(updatePoolTxMap *UpdatePoolTx) error {
	start := time.Now()
	if len(updatePoolTxMap.PendingUpdatePoolTxs) > 0 {
		ids := make([]uint, 0, len(updatePoolTxMap.PendingUpdatePoolTxs))
		updateNftIndexOrCollectionIdList := make([]*tx.PoolTx, 0)
		var poolIdStr string
		for _, pendingUpdatePoolTx := range updatePoolTxMap.PendingUpdatePoolTxs {
			ids = append(ids, pendingUpdatePoolTx.ID)
			if !pendingUpdatePoolTx.IsPartialUpdate {
				continue
			}
			updateNftIndexOrCollectionIdList = append(updateNftIndexOrCollectionIdList, &tx.PoolTx{
				BaseTx: tx.BaseTx{Model: gorm.Model{ID: pendingUpdatePoolTx.ID},
					NftIndex:         pendingUpdatePoolTx.NftIndex,
					CollectionId:     pendingUpdatePoolTx.CollectionId,
					AccountIndex:     pendingUpdatePoolTx.AccountIndex,
					IsCreateAccount:  pendingUpdatePoolTx.IsCreateAccount,
					FromAccountIndex: pendingUpdatePoolTx.FromAccountIndex,
					ToAccountIndex:   pendingUpdatePoolTx.ToAccountIndex,
				},
			})
			poolIdStr += fmt.Sprintf("%d,", pendingUpdatePoolTx.ID)
		}
		ctx := log.NewCtxWithKV(log.PoolTxIdListContext, poolIdStr)
		if len(updateNftIndexOrCollectionIdList) > 0 {
			err := c.bc.TxPoolModel.BatchUpdateNftIndexOrCollectionId(updateNftIndexOrCollectionIdList)
			if err != nil {
				logx.WithContext(ctx).Error("update tx pool failed:", err)
				return nil
			}
			jsonInfo, err := json.Marshal(updateNftIndexOrCollectionIdList)
			if err == nil {
				logx.WithContext(ctx).Infof("update tx pool success,%s", jsonInfo)
			}
		}
		err := c.bc.TxPoolModel.UpdateTxsStatusAndHeightByIds(ids, tx.StatusExecuted, updatePoolTxMap.PendingUpdatePoolTxs[0].BlockHeight)
		if err != nil {
			logx.WithContext(ctx).Error("update tx pool failed:", err)
		}
	}

	if len(updatePoolTxMap.PendingDeletePoolTxs) > 0 {
		poolTxIds := make([]uint, 0, len(updatePoolTxMap.PendingDeletePoolTxs))
		var poolTxIdsStr string
		for _, poolTx := range updatePoolTxMap.PendingDeletePoolTxs {
			poolTxIds = append(poolTxIds, poolTx.ID)
			poolTxIdsStr += fmt.Sprintf("%d,", poolTx.ID)
		}
		err := c.bc.TxPoolModel.DeleteTxsBatch(poolTxIds, tx.StatusFailed, -1)
		if err != nil {
			logx.WithContext(log.NewCtxWithKV(log.PoolTxIdContext, poolTxIdsStr)).Error("update tx pool failed:", err)
		}
	}
	metrics.UpdatePoolTxsMetrics.Set(float64(time.Since(start).Milliseconds()))
	return nil
}

// Put the accounts and nfts data that need to be synchronized to redis into the queue
func (c *Committer) addSyncAccountToRedisToQueue(originPendingAccountMap map[int64]*types.AccountInfo, originPendingNftMap map[int64]*nft.L2Nft) {
	if len(originPendingAccountMap) == 0 && len(originPendingNftMap) == 0 {
		return
	}
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

// sync accounts and nfts to redis
func (c *Committer) syncAccountToRedisFunc(pendingMap *PendingMap) error {
	start := time.Now()
	c.bc.Statedb.SyncPendingAccountToRedis(pendingMap.PendingAccountMap)
	c.bc.Statedb.SyncPendingNftToRedis(pendingMap.PendingNftMap)
	metrics.SyncAccountToRedisMetrics.Set(float64(time.Since(start).Milliseconds()))
	return nil
}

// preSaveBlockData,eg:AccountIndexes,NftIndexes
func (c *Committer) preSaveBlockDataFunc(stateDataCopy *statedb.StateDataCopy) error {
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
		return fmt.Errorf("marshal accountIndexes failed:%s,blockHeight:%d", err, stateDataCopy.CurrentBlock.BlockHeight)
	}
	nftIndexesJson, err := json.Marshal(nftIndexes)
	if err != nil {
		return fmt.Errorf("marshal nftIndexesJson failed:%s,blockHeight:%d", err, stateDataCopy.CurrentBlock.BlockHeight)
	}

	stateDataCopy.CurrentBlock.AccountIndexes = string(accountIndexesJson)
	stateDataCopy.CurrentBlock.NftIndexes = string(nftIndexesJson)
	stateDataCopy.CurrentBlock.BlockStatus = block.StatusPacked

	err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
		return c.bc.DB().BlockModel.PreSaveBlockDataInTransact(dbTx, stateDataCopy.CurrentBlock)
	})
	if err != nil {
		return fmt.Errorf("preSaveBlockDataFunc failed:%s,blockHeight:%d", err, stateDataCopy.CurrentBlock.BlockHeight)
	}

	latestVerifiedBlockNr, err := c.bc.BlockModel.GetLatestVerifiedHeight()
	if err != nil {
		return fmt.Errorf("get latest verified height failed:%s", err.Error())
	}
	c.bc.Statedb.UpdatePrunedBlockHeight(latestVerifiedBlockNr)

	metrics.PreSaveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
	c.updateAssetTreeWorker.Enqueue(stateDataCopy)
	return nil
}

// compute account asset hash, commit asset smt,compute account leaf hash, compute nft leaf hash
func (c *Committer) updateAssetTreeFunc(stateDataCopy *statedb.StateDataCopy) error {
	start := time.Now()
	metrics.UpdateAssetTreeTxMetrics.Add(float64(len(stateDataCopy.StateCache.Txs)))
	logx.Infof("updateAssetTreeFunc blockHeight:%d,blockId:%d", stateDataCopy.CurrentBlock.BlockHeight, stateDataCopy.CurrentBlock.ID)
	err := c.bc.UpdateAssetTree(stateDataCopy)
	if err != nil {
		return fmt.Errorf("updateAssetTreeFunc failed:%s,blockHeight:%d", err, stateDataCopy.CurrentBlock.BlockHeight)
	}
	c.updateAccountAndNftTreeWorker.Enqueue(stateDataCopy)
	metrics.AccountAndNftTreeQueueMetric.Set(float64(c.updateAccountAndNftTreeWorker.GetQueueSize()))
	metrics.UpdateAccountAssetTreeMetrics.Set(float64(time.Since(start).Milliseconds()))
	return nil
}

// UpdateAccountAndNftTree multi set account tree with version,multi set nft tree with version
// commit account and nft tree
// build Block CompressedBlock PendingAccount PendingAccountHistory PendingNft PendingNftHistory
func (c *Committer) updateAccountAndNftTreeFunc(stateDataCopy *statedb.StateDataCopy) error {
	start := time.Now()
	metrics.UpdateAccountAndNftTreeTxMetrics.Add(float64(len(stateDataCopy.StateCache.Txs)))

	logx.Infof("updateAccountAndNftTreeFunc blockHeight:%d,blockId:%d", stateDataCopy.CurrentBlock.BlockHeight, stateDataCopy.CurrentBlock.ID)
	blockSize := c.computeCurrentBlockSize(stateDataCopy)
	if blockSize < len(stateDataCopy.StateCache.Txs) {
		return fmt.Errorf("block size too small")
	}

	blockStates, err := c.bc.UpdateAccountAndNftTree(blockSize, stateDataCopy)
	if err != nil {
		return fmt.Errorf("updateAccountAndNftTreeFunc failed:%s,blockHeight:%d", err, stateDataCopy.CurrentBlock.BlockHeight)
	}

	c.saveBlockDataWorker.Enqueue(blockStates)

	metrics.L2BlockRedisHeightMetric.Set(float64(blockStates.Block.BlockHeight))
	metrics.AccountLatestVersionTreeMetric.Set(float64(c.bc.StateDB().AccountTree.LatestVersion()))
	metrics.AccountRecentVersionTreeMetric.Set(float64(c.bc.StateDB().AccountTree.RecentVersion()))
	metrics.NftTreeLatestVersionMetric.Set(float64(c.bc.StateDB().NftTree.LatestVersion()))
	metrics.NftTreeRecentVersionMetric.Set(float64(c.bc.StateDB().NftTree.RecentVersion()))
	metrics.UpdateAccountTreeAndNftTreeMetrics.Set(float64(time.Since(start).Milliseconds()))

	return nil
}

// save block data
func (c *Committer) saveBlockDataFunc(blockStates *block.BlockStates) error {
	start := time.Now()
	ctx := log.NewCtxWithKV(log.BlockHeightContext, blockStates.Block.BlockHeight)
	logx.WithContext(ctx).Infof("saveBlockDataFunc start, blockHeight:%d", blockStates.Block.BlockHeight)
	totalTask := 0
	errChan := make(chan error, 1)
	defer close(errChan)
	var err error

	poolTxIds := make([]uint, 0, len(blockStates.Block.Txs))
	updateNftIndexOrCollectionIdList := make([]*tx.PoolTx, 0)

	for _, poolTx := range blockStates.Block.Txs {
		poolTxIds = append(poolTxIds, poolTx.ID)
		if !poolTx.IsPartialUpdate {
			continue
		}
		updateNftIndexOrCollectionIdList = append(updateNftIndexOrCollectionIdList, &tx.PoolTx{BaseTx: tx.BaseTx{
			Model:            gorm.Model{ID: poolTx.ID},
			NftIndex:         poolTx.NftIndex,
			CollectionId:     poolTx.CollectionId,
			AccountIndex:     poolTx.AccountIndex,
			IsCreateAccount:  poolTx.IsCreateAccount,
			FromAccountIndex: poolTx.FromAccountIndex,
			ToAccountIndex:   poolTx.ToAccountIndex,
		}})
	}

	blockStates.Block.ClearTxsModel()
	totalTask++
	err = func(poolTxIds []uint, blockHeight int64, updateNftIndexOrCollectionIdList []*tx.PoolTx) error {
		return c.pool.Submit(func() {
			start := time.Now()
			if len(updateNftIndexOrCollectionIdList) > 0 {
				err := c.bc.TxPoolModel.BatchUpdateNftIndexOrCollectionId(updateNftIndexOrCollectionIdList)
				if err != nil {
					logx.WithContext(ctx).Error("update tx pool failed:", err)
					errChan <- err
					return
				}
			}

			err = c.bc.DB().TxPoolModel.DeleteTxsBatch(poolTxIds, tx.StatusExecuted, blockHeight)
			metrics.DeletePoolTxMetrics.Set(float64(time.Since(start).Milliseconds()))
			if err != nil {
				errChan <- err
				return
			}
			errChan <- nil
		})
	}(poolTxIds, blockStates.Block.BlockHeight, updateNftIndexOrCollectionIdList)
	if err != nil {
		return fmt.Errorf("DeleteTxsBatch failed: %s", err.Error())
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
					metrics.SaveAccountsGoroutineMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(accounts)
			if err != nil {
				return fmt.Errorf("batchInsertOrUpdate accounts failed: %s", err.Error())
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
					metrics.AddAccountHistoryMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(accountHistories)
			if err != nil {
				return fmt.Errorf("createAccountHistories failed: %s", err.Error())
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
					metrics.SaveAccountsGoroutineMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(nfts)
			if err != nil {
				return fmt.Errorf("batchInsertOrUpdate nfts failed: %s", err.Error())
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
					metrics.AddAccountHistoryMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(nftHistories)
			if err != nil {
				return fmt.Errorf("createNftHistories failed: %s", err.Error())
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
					metrics.AddTxsMetrics.Set(float64(time.Since(start).Milliseconds()))
					if err != nil {
						errChan <- err
						return
					}
					errChan <- nil
				})
			}(txs)
			if err != nil {
				return fmt.Errorf("CreateTxs failed: %s", err.Error())
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
						metrics.AddTxDetailsMetrics.Set(float64(time.Since(start).Milliseconds()))
						if err != nil {
							errChan <- err
							return
						}
						errChan <- nil
					})
				}(txDetailsSlice)
				if err != nil {
					return fmt.Errorf("CreateTxDetails failed: %s", err.Error())
				}
			}
		}
	}
	for i := 0; i < totalTask; i++ {
		err := <-errChan
		if err != nil {
			return fmt.Errorf("saveBlockDataFunc failed: %s", err.Error())
		}
	}

	metrics.SaveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
	c.finalSaveBlockDataWorker.Enqueue(blockStates)
	return nil
}

// final save block data
func (c *Committer) finalSaveBlockDataFunc(blockStates *block.BlockStates) error {
	start := time.Now()
	logx.Infof("finalSaveBlockDataFunc start, blockHeight:%d", blockStates.Block.BlockHeight)
	// update db
	err := c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
		if blockStates.CompressedBlock != nil {
			start := time.Now()
			err := c.bc.DB().CompressedBlockModel.CreateCompressedBlockInTransact(tx, blockStates.CompressedBlock)
			metrics.FinalSaveBlockDataMetrics.WithLabelValues("add_compressed_block").Set(float64(time.Since(start).Milliseconds()))
			if err != nil {
				return err
			}
		}
		start := time.Now()
		err := c.bc.DB().BlockModel.UpdateBlockToPendingInTransact(tx, blockStates.Block)
		if err != nil {
			return err
		}
		metrics.FinalSaveBlockDataMetrics.WithLabelValues("update_block_to_pending").Set(float64(time.Since(start).Milliseconds()))
		return nil
	})
	if err != nil {
		return fmt.Errorf("finalSaveBlockDataFunc failed:%s,blockHeight:%d", err.Error(), blockStates.Block.BlockHeight)
	}
	c.bc.Statedb.UpdateMaxPoolTxIdFinished(blockStates.Block.Txs[len(blockStates.Block.Txs)-1].PoolTxId)
	metrics.L2BlockDbHeightMetric.Set(float64(blockStates.Block.BlockHeight))
	metrics.FinalSaveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
	return nil
}

// create new block
func (c *Committer) createNewBlock(curBlock *block.Block) error {
	return c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
		return c.bc.BlockModel.CreateBlockInTransact(dbTx, curBlock)
	})
}

func (c *Committer) shouldCommit(curBlock *block.Block) bool {
	//After the rollback, re-execute tx and form a block  based on the block size,because curBlock.CreatedAt does not change
	if c.bc.Statedb.NeedRestoreExecutedTxs() {
		if len(c.bc.Statedb.Txs) >= c.maxTxsPerBlock {
			return true
		}
		return false
	}
	var now = time.Now()
	if (len(c.bc.Statedb.Txs) > 0 && now.Unix()-curBlock.CreatedAt.Unix() >= int64(c.maxCommitterInterval)) ||
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

	latestTx, err := c.bc.TxPoolModel.GetLatestTx(types.GetL1TxTypes(), statuses)
	if err != nil && err != types.DbErrNotFound {
		logx.Severef("get latest executed tx failed: %v", err)
		return -1, err
	} else if err == types.DbErrNotFound {
		return -1, nil
	}
	return latestTx.L1RequestId, nil
}

func (c *Committer) preLoadAccountAndNft(txs []*tx.Tx) {
	var accountIndexMap map[int64]bool
	var nftIndexMap map[int64]bool
	var addressMap map[string]bool
	for _, poolTx := range txs {
		c.bc.PreApplyTransaction(poolTx, accountIndexMap, nftIndexMap, addressMap)
	}
	c.bc.Statedb.PreLoadAccountAndNft(accountIndexMap, nftIndexMap, addressMap)
}

func (c *Committer) PendingTxNum() {
	txStatuses := []int64{tx.StatusPending}
	pendingTxCount, _ := c.bc.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
	metrics.PendingTxNumMetrics.Set(float64(pendingTxCount))
}

func (c *Committer) CompensatePendingPoolTx() {
	fromCreatAt := time.Now().Add(time.Duration(-10*c.maxCommitterInterval) * time.Second).UnixMilli()
	pendingTxs, err := c.bc.TxPoolModel.GetTxsByStatusAndCreateTime(tx.StatusPending, time.UnixMilli(fromCreatAt), c.bc.Statedb.GetMaxPoolTxIdFinished())
	if err != nil {
		logx.Errorf("get pending transactions from tx pool for compensation failed:%s", err.Error())
		return
	}

	for _, poolTx := range pendingTxs {
		logx.Severef("get pending transactions from tx pool for compensation id:%d", poolTx.ID)
		_, found := c.bc.Statedb.MemCache.Get(dbcache.PendingPoolTxKeyByPoolTxId(poolTx.ID))
		if found {
			logx.Infof("add pool tx to the queue repeatedly in the compensation task id:%d", poolTx.ID)
			continue
		}
		c.bc.Statedb.MemCache.SetWithTTL(dbcache.PendingPoolTxKeyByPoolTxId(poolTx.ID), poolTx.ID, 0, time.Duration(c.maxCommitterInterval*50)*time.Second)
		c.executeTxWorker.Enqueue(poolTx)
	}
}

func (c *Committer) Shutdown() {
	c.running = false
	c.executeTxWorker.Stop()
	c.syncAccountToRedisWorker.Stop()
	c.updatePoolTxWorker.Stop()
	c.updateAssetTreeWorker.Stop()
	c.updateAccountAndNftTreeWorker.Stop()
	c.preSaveBlockDataWorker.Stop()
	c.saveBlockDataWorker.Stop()
	c.finalSaveBlockDataWorker.Stop()
	c.bc.Statedb.Close()
	c.bc.ChainDB.Close()
}

func (c *Committer) SyncNftIndexServer() error {
	histories, err := c.bc.L2NftMetadataHistoryModel.GetL2NftMetadataHistoryList(nft.StatusNftIndex)
	if err != nil {
		return nil
	}
	for _, history := range histories {
		poolTx, err := c.bc.TxPoolModel.GetTxUnscopedByTxHash(history.TxHash)
		if err != nil {
			return err
		}
		if poolTx.TxStatus == tx.StatusFailed {
			err = c.bc.L2NftMetadataHistoryModel.DeleteInTransact(history.ID)
			if err != nil {
				return err
			}
		} else if poolTx.TxStatus == tx.StatusExecuted {
			tx, err := c.bc.TxModel.GetTxByHash(history.TxHash)
			if err != nil {
				return err
			}
			history.NftIndex = tx.NftIndex
			history.Status = nft.NotConfirmed
			err = c.bc.L2NftMetadataHistoryModel.UpdateL2NftMetadataHistoryInTransact(history)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Committer) SendIpfsServer() error {
	histories, err := c.bc.L2NftMetadataHistoryModel.GetL2NftMetadataHistoryList(nft.NotConfirmed)
	if err != nil {
		return nil
	}
	for _, history := range histories {
		err = saveIpfs(history)
		if err != nil {
			return err
		}
		err = c.bc.L2NftMetadataHistoryModel.UpdateL2NftMetadataHistoryInTransact(history)
		if err != nil {
			return err
		}
	}
	return nil
}

func saveIpfs(history *nft.L2NftMetadataHistory) error {
	cid, err := common.Ipfs.Upload(history.Mutable)
	if err != nil {
		return err
	}
	_, err = common.Ipfs.PublishWithDetails(cid, history.IpnsName)
	if err != nil {
		return err
	}
	history.Status = nft.Confirmed
	history.IpnsCid = cid
	return nil
}

func (c *Committer) RefreshServer() error {
	limit := 500
	offset := 0
	for {
		histories, err := c.bc.L2NftMetadataHistoryModel.GetL2NftMetadataHistoryPage(nft.Confirmed, limit, offset)
		if err != nil {
			return nil
		}
		for _, hostory := range histories {
			_, err = common.Ipfs.PublishWithDetails(hostory.IpnsCid, hostory.IpnsName)
			if err != nil {
				return err
			}
		}
		if len(histories) < limit {
			break
		} else {
			offset = offset + limit
		}
	}
	return nil
}
