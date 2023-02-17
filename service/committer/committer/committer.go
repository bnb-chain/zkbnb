package committer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/gopool"
	"github.com/bnb-chain/zkbnb/common/metrics"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/panjf2000/ants/v2"
	"sort"
	"strconv"
	"time"

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

type Config struct {
	core.ChainConfig

	BlockConfig struct {
		OptionalBlockSizes    []int
		SaveBlockDataPoolSize int  `json:",optional"`
		RollbackOnly          bool `json:",optional"`
	}
	LogConf logx.LogConf
	IpfsUrl string
}

type Committer struct {
	running            bool
	config             *Config
	maxTxsPerBlock     int
	optionalBlockSizes []int

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

//work flow: pullPoolTxs->executeTxFunc->updatePoolTxFunc->preSaveBlockDataFunc
//->updateAssetTreeFunc->updateAccountAndNftTreeFunc->saveBlockDataFunc->finalSaveBlockDataFunc

func NewCommitter(config *Config) (*Committer, error) {
	if len(config.BlockConfig.OptionalBlockSizes) == 0 {
		return nil, errors.New("nil optional block sizes")
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
	pool, err := ants.NewPool(saveBlockDataPoolSize)
	common.NewIPFS(config.IpfsUrl)
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

func (c *Committer) Run() {
	if c.config.BlockConfig.RollbackOnly {
		for {
			if !c.running {
				break
			}
			logx.Info("do rollback only")
			time.Sleep(1 * time.Minute)
		}
		return
	}
	c.executeTxWorker = core.ExecuteTxWorker(10000, func() {
		c.executeTxFunc()
	})
	c.updatePoolTxWorker = core.UpdatePoolTxWorker(10000, func(item interface{}) {
		c.updatePoolTxFunc(item.(*UpdatePoolTx))
	})
	c.syncAccountToRedisWorker = core.SyncAccountToRedisWorker(10000, func(item interface{}) {
		c.syncAccountToRedisFunc(item.(*PendingMap))
	})
	c.preSaveBlockDataWorker = core.PreSaveBlockDataWorker(10, func(item interface{}) {
		c.preSaveBlockDataFunc(item.(*statedb.StateDataCopy))
	})
	c.updateAssetTreeWorker = core.UpdateAssetTreeWorker(10, func(item interface{}) {
		c.updateAssetTreeFunc(item.(*statedb.StateDataCopy))
	})
	c.updateAccountAndNftTreeWorker = core.UpdateAccountAndNftTreeWorker(10, func(item interface{}) {
		c.updateAccountAndNftTreeFunc(item.(*statedb.StateDataCopy))
	})
	c.saveBlockDataWorker = core.SaveBlockDataWorker(10, func(item interface{}) {
		c.saveBlockDataFunc(item.(*block.BlockStates))
	})
	c.finalSaveBlockDataWorker = core.FinalSaveBlockDataWorker(10, func(item interface{}) {
		c.finalSaveBlockDataFunc(item.(*block.BlockStates))
	})

	c.loadAllAccounts()
	c.loadAllNfts()
	c.executeTxWorker.Start()
	c.syncAccountToRedisWorker.Start()
	c.updatePoolTxWorker.Start()
	c.preSaveBlockDataWorker.Start()
	c.updateAssetTreeWorker.Start()
	c.updateAccountAndNftTreeWorker.Start()
	c.saveBlockDataWorker.Start()
	c.finalSaveBlockDataWorker.Start()

	c.pullPoolTxs()
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
	pendingTxs := make([]*tx.Tx, 0)
	if c.bc.Statedb.NeedRestoreExecutedTxs() {
		pendingTxs, err = c.bc.TxPoolModel.GetTxsByStatusAndIdRange(tx.StatusPending, executedTxMaxId+1, c.bc.Statedb.MaxPollTxIdRollbackImmutable)
		if err != nil {
			logx.Errorf("get rollback transactions from tx pool failed:%s", err.Error())
			panic("get rollback transactions from tx pool failed: " + err.Error())
		}
		executedTxMaxId = c.bc.Statedb.MaxPollTxIdRollbackImmutable
		metrics.GetPendingTxsToQueueMetric.Set(float64(len(pendingTxs)))
		for _, poolTx := range pendingTxs {
			c.executeTxWorker.Enqueue(poolTx)
		}
	}

	limit := 1000
	for {
		if !c.running {
			break
		}
		start := time.Now()
		pendingTxs, err = c.bc.TxPoolModel.GetTxsByStatusAndMaxId(tx.StatusPending, executedTxMaxId, int64(limit))
		metrics.GetPendingPoolTxMetrics.Set(float64(time.Since(start).Milliseconds()))
		metrics.GetPendingTxsToQueueMetric.Set(float64(len(pendingTxs)))
		metrics.TxWorkerQueueMetric.Set(float64(c.executeTxWorker.GetQueueSize()))
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
	l1LatestRequestId, err := c.getLatestExecutedRequestId()
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
			logx.Infof("1 init new block, current height=%d,previous height=%d,blockId=%d", curBlock.BlockHeight, previousHeight, curBlock.ID)
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
				c.addUpdatePoolTxQueue(pendingUpdatePoolTxs, nil)
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
			metrics.ExecuteTxMetrics.Inc()
			startApplyTx := time.Now()
			logx.Infof("start apply pool tx ID: %d", poolTx.ID)
			err = c.bc.ApplyTransaction(poolTx)
			if c.bc.Statedb.NeedRestoreExecutedTxs() && poolTx.ID >= c.bc.Statedb.MaxPollTxIdRollbackImmutable {
				logx.Infof("update needRestoreExecutedTxs to false,blockHeight:%d", curBlock.BlockHeight)
				c.bc.Statedb.UpdateNeedRestoreExecutedTxs(false)
				err := c.bc.DB().BlockModel.DeleteBlockGreaterThanHeight(curBlock.BlockHeight, []int{block.StatusProposing, block.StatusPacked})
				if err != nil {
					logx.Errorf("DeleteBlockGreaterThanHeight failed:%s,blockHeight:%d", err.Error(), curBlock.BlockHeight)
					panic("DeleteBlockGreaterThanHeight failed: " + err.Error())
				}
				c.bc.ClearRollbackBlockMap()
			}
			metrics.ExecuteTxApply1TxMetrics.Set(float64(time.Since(startApplyTx).Milliseconds()))
			if err != nil {
				logx.Severef("apply pool tx ID: %d failed, err %v ", poolTx.ID, err)
				if types.IsPriorityOperationTx(poolTx.TxType) {
					metrics.PoolTxL1ErrorCountMetics.Inc()
					logx.Severef("apply priority pool tx failed,id=%s,error=%s", strconv.Itoa(int(poolTx.ID)), err.Error())
					panic("apply priority pool tx failed,id=" + strconv.Itoa(int(poolTx.ID)) + ",error=" + err.Error())
				} else {
					expectNonce, err := c.bc.Statedb.GetCommittedNonce(poolTx.AccountIndex)
					if err != nil {
						c.bc.Statedb.ClearPendingNonceFromRedisCache(poolTx.AccountIndex)
					} else {
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
					logx.Severef("invalid request id=%s,txHash=%s", strconv.Itoa(int(poolTx.L1RequestId)), err.Error())
					panic("invalid request id=" + strconv.Itoa(int(poolTx.L1RequestId)) + ",txHash=" + poolTx.TxHash)
				}
				l1LatestRequestId = poolTx.L1RequestId
			}

			// Write the proposed block into database when the first transaction executed.
			if len(c.bc.Statedb.Txs) == 1 {
				previousHeight := curBlock.BlockHeight
				if curBlock.ID == 0 {
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
		metrics.ExecuteTxOperationMetrics.Set(float64(time.Since(start).Milliseconds()))

		c.bc.Statedb.SyncPendingAccountToMemoryCache(c.bc.Statedb.PendingAccountMap)
		c.bc.Statedb.SyncPendingNftToMemoryCache(c.bc.Statedb.PendingNftMap)

		c.addSyncAccountToRedisQueue(c.bc.Statedb.PendingAccountMap, c.bc.Statedb.PendingNftMap)
		c.addUpdatePoolTxQueue(nil, pendingDeletePoolTxs)

		if c.shouldCommit(curBlock) {
			start := time.Now()
			logx.Infof("commit new block, height=%d,blockSize=%d", curBlock.BlockHeight, curBlock.BlockSize)
			pendingUpdatePoolTxs = make([]*tx.Tx, 0, c.maxTxsPerBlock)
			stateDataCopy := c.buildStateDataCopy(curBlock)
			c.preSaveBlockDataWorker.Enqueue(stateDataCopy)
			metrics.AccountAssetTreeQueueMetric.Set(float64(c.updateAssetTreeWorker.GetQueueSize()))

			metrics.L2BlockMemoryHeightMetric.Set(float64(stateDataCopy.CurrentBlock.BlockHeight))
			previousHeight := stateDataCopy.CurrentBlock.BlockHeight
			curBlock, err = c.bc.InitNewBlock()
			if err != nil {
				logx.Errorf("propose new block failed:%s ", err.Error())
				panic("propose new block failed: " + err.Error())
			}
			logx.Infof("2 init new block, current height=%d,previous height=%d,blockId=%d", curBlock.BlockHeight, previousHeight, curBlock.ID)

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

func (c *Committer) buildStateDataCopy(curBlock *block.Block) *statedb.StateDataCopy {
	deleteGasAccount := false
	for _, formatAccount := range c.bc.Statedb.StateCache.PendingAccountMap {
		if formatAccount.AccountIndex == types.GasAccount {
			if len(c.bc.Statedb.StateCache.PendingGasMap) != 0 {
				for assetId, delta := range c.bc.Statedb.StateCache.PendingGasMap {
					if asset, ok := formatAccount.AssetInfo[assetId]; ok {
						formatAccount.AssetInfo[assetId].Balance = ffmath.Add(asset.Balance, delta)
					} else {
						formatAccount.AssetInfo[assetId] = &types.AccountAsset{
							Balance:                  delta,
							OfferCanceledOrFinalized: types.ZeroBigInt,
						}
					}
					c.bc.Statedb.MarkAccountAssetsDirty(formatAccount.AccountIndex, []int64{assetId})
				}
			} else {
				assetsMap := c.bc.Statedb.GetDirtyAccountsAndAssetsMap()[formatAccount.AccountIndex]
				if assetsMap == nil {
					deleteGasAccount = true
				}
			}
		} else {
			assetsMap := c.bc.Statedb.GetDirtyAccountsAndAssetsMap()[formatAccount.AccountIndex]
			if assetsMap == nil {
				logx.Errorf("%s exists in PendingAccountMap but not in GetDirtyAccountsAndAssetsMap", formatAccount.AccountIndex)
				panic(strconv.FormatInt(formatAccount.AccountIndex, 10) + " exists in PendingAccountMap but not in GetDirtyAccountsAndAssetsMap")
			}
		}
	}
	if deleteGasAccount {
		delete(c.bc.Statedb.StateCache.PendingAccountMap, types.GasAccount)
	}
	for accountIndex, _ := range c.bc.Statedb.GetDirtyAccountsAndAssetsMap() {
		_, exist := c.bc.Statedb.StateCache.GetPendingAccount(accountIndex)
		if !exist {
			accountInfo, err := c.bc.Statedb.GetFormatAccount(accountIndex)
			if err != nil {
				logx.Errorf("get account info failed,accountIndex=%s,err=%s ", accountIndex, err.Error())
				panic("get account info failed: " + err.Error())
			}
			c.bc.Statedb.SetPendingAccount(accountIndex, accountInfo)
		}
	}

	for _, nftInfo := range c.bc.Statedb.StateCache.PendingNftMap {
		if c.bc.Statedb.GetDirtyNftMap()[nftInfo.NftIndex] == false {
			logx.Errorf("%s exists in PendingNftMap but not in DirtyNftMap", nftInfo.NftIndex)
			panic(strconv.FormatInt(nftInfo.NftIndex, 10) + " exists in PendingNftMap but not in DirtyNftMap")
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
			logx.Errorf("account info format failed:%s ", err.Error())
			panic("account info format failed: " + err.Error())
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
				err := c.bc.DB().AccountModel.BatchInsertOrUpdateInTransact(dbTx, addPendingAccounts)
				if err != nil {
					logx.Errorf("account batch insert or update failed:%s ", err.Error())
					panic("account batch insert or update failed: " + err.Error())
				}
			}
			if len(addPendingNfts) != 0 {
				err := c.bc.DB().L2NftModel.BatchInsertOrUpdateInTransact(dbTx, addPendingNfts)
				if err != nil {
					logx.Errorf("l2nft batch insert or update failed:%s ", err.Error())
					panic("l2nft batch insert or update failed: " + err.Error())
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
				logx.Errorf("marshal accountIndexes failed:%s,blockHeight:%s", err, curBlock.BlockHeight)
				panic("marshal accountIndexes failed: " + err.Error())
			}
			nftIndexesJson, err := json.Marshal(nftIndexes)
			if err != nil {
				logx.Errorf("marshal nftIndexesJson failed:%s,blockHeight:%s", err, curBlock.BlockHeight)
				panic("marshal nftIndexesJson failed: " + err.Error())
			}
			curBlock.AccountIndexes = string(accountIndexesJson)
			curBlock.NftIndexes = string(nftIndexesJson)
			curBlock.BlockStatus = block.StatusPacked
			err = c.bc.DB().BlockModel.PreSaveBlockDataInTransact(dbTx, curBlock)
			if err != nil {
				logx.Errorf("PreSaveBlockDataInTransact failed:%s,blockHeight:%s", err, curBlock.BlockHeight)
				panic("PreSaveBlockDataInTransact failed: " + err.Error())
			}
			return nil
		})
		if err != nil {
			logx.Errorf("account or nft insert failed:%s ", err.Error())
			panic("account or nft insert failed: " + err.Error())
		}

		for _, accountInfo := range addPendingAccounts {
			c.bc.Statedb.StateCache.PendingAccountMap[accountInfo.AccountIndex].AccountId = accountInfo.ID
			updateAccountMap[accountInfo.AccountIndex] = c.bc.Statedb.StateCache.PendingAccountMap[accountInfo.AccountIndex]
		}
		for _, nftInfo := range addPendingNfts {
			c.bc.Statedb.StateCache.PendingNftMap[nftInfo.NftIndex].ID = nftInfo.ID
			updateNftMap[nftInfo.NftIndex] = c.bc.Statedb.StateCache.PendingNftMap[nftInfo.NftIndex]
		}
	}
	gasAccount := c.bc.Statedb.StateCache.PendingAccountMap[types.GasAccount]
	if gasAccount != nil {
		updateAccountMap[types.GasAccount] = gasAccount
	}
	c.bc.Statedb.SyncPendingAccountToMemoryCache(updateAccountMap)
	c.bc.Statedb.SyncPendingNftToMemoryCache(updateNftMap)
	c.addSyncAccountToRedisQueue(updateAccountMap, updateNftMap)

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
	return stateDataCopy
}

func (c *Committer) addUpdatePoolTxQueue(pendingUpdatePoolTxs []*tx.Tx, pendingDeletePoolTxs []*tx.Tx) {
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
						BaseTx: tx.BaseTx{Model: gorm.Model{ID: pendingUpdatePoolTx.ID},
							NftIndex:     pendingUpdatePoolTx.NftIndex,
							CollectionId: pendingUpdatePoolTx.CollectionId},
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
	metrics.UpdatePoolTxsMetrics.Set(float64(time.Since(start).Milliseconds()))
}

func (c *Committer) addSyncAccountToRedisQueue(originPendingAccountMap map[int64]*types.AccountInfo, originPendingNftMap map[int64]*nft.L2Nft) {
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

func (c *Committer) syncAccountToRedisFunc(pendingMap *PendingMap) {
	start := time.Now()
	c.bc.Statedb.SyncPendingAccountToRedis(pendingMap.PendingAccountMap)
	c.bc.Statedb.SyncPendingNftToRedis(pendingMap.PendingNftMap)
	metrics.SyncAccountToRedisMetrics.Set(float64(time.Since(start).Milliseconds()))
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
	err = c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
		return c.bc.DB().BlockModel.PreSaveBlockDataInTransact(dbTx, stateDataCopy.CurrentBlock)
	})
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

	metrics.PreSaveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
	c.updateAssetTreeWorker.Enqueue(stateDataCopy)
}

func (c *Committer) updateAssetTreeFunc(stateDataCopy *statedb.StateDataCopy) {
	start := time.Now()
	metrics.UpdateAssetTreeTxMetrics.Add(float64(len(stateDataCopy.StateCache.Txs)))
	logx.Infof("updateAssetTreeFunc blockHeight:%s,blockId:%s", stateDataCopy.CurrentBlock.BlockHeight, stateDataCopy.CurrentBlock.ID)
	err := c.bc.UpdateAssetTree(stateDataCopy)
	if err != nil {
		logx.Errorf("updateAssetTreeFunc failed:%s,blockHeight:%s", err, stateDataCopy.CurrentBlock.BlockHeight)
		panic("updateAssetTreeFunc failed: " + err.Error())
	}
	c.updateAccountAndNftTreeWorker.Enqueue(stateDataCopy)
	metrics.AccountAndNftTreeQueueMetric.Set(float64(c.updateAccountAndNftTreeWorker.GetQueueSize()))
	metrics.UpdateAccountAssetTreeMetrics.Set(float64(time.Since(start).Milliseconds()))

}

func (c *Committer) updateAccountAndNftTreeFunc(stateDataCopy *statedb.StateDataCopy) {
	start := time.Now()
	metrics.UpdateAccountAndNftTreeTxMetrics.Add(float64(len(stateDataCopy.StateCache.Txs)))
	logx.Infof("updateAccountAndNftTreeFunc blockHeight:%s,blockId:%s", stateDataCopy.CurrentBlock.BlockHeight, stateDataCopy.CurrentBlock.ID)
	blockSize := c.computeCurrentBlockSize(stateDataCopy)
	if blockSize < len(stateDataCopy.StateCache.Txs) {
		panic("block size too small")
	}
	blockStates, err := c.bc.UpdateAccountAndNftTree(blockSize, stateDataCopy)
	if err != nil {
		logx.Errorf("updateAccountAndNftTreeFunc failed:%s,blockHeight:%s", err, stateDataCopy.CurrentBlock.BlockHeight)
		panic("updateAccountAndNftTreeFunc failed: " + err.Error())
	}
	c.saveBlockDataWorker.Enqueue(blockStates)
	metrics.L2BlockRedisHeightMetric.Set(float64(blockStates.Block.BlockHeight))
	metrics.AccountLatestVersionTreeMetric.Set(float64(c.bc.StateDB().AccountTree.LatestVersion()))
	metrics.AccountRecentVersionTreeMetric.Set(float64(c.bc.StateDB().AccountTree.RecentVersion()))
	metrics.NftTreeLatestVersionMetric.Set(float64(c.bc.StateDB().NftTree.LatestVersion()))
	metrics.NftTreeRecentVersionMetric.Set(float64(c.bc.StateDB().NftTree.RecentVersion()))
	metrics.UpdateAccountTreeAndNftTreeMetrics.Set(float64(time.Since(start).Milliseconds()))
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
			updateNftIndexOrCollectionIdList = append(updateNftIndexOrCollectionIdList, &tx.PoolTx{BaseTx: tx.BaseTx{
				Model:        gorm.Model{ID: poolTx.ID},
				NftIndex:     poolTx.NftIndex,
				CollectionId: poolTx.CollectionId,
			}})
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
			metrics.DeletePoolTxMetrics.Set(float64(time.Since(start).Milliseconds()))
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
					metrics.SaveAccountsGoroutineMetrics.Set(float64(time.Since(start).Milliseconds()))
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
					metrics.AddAccountHistoryMetrics.Set(float64(time.Since(start).Milliseconds()))
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
					metrics.SaveAccountsGoroutineMetrics.Set(float64(time.Since(start).Milliseconds()))
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
					metrics.AddAccountHistoryMetrics.Set(float64(time.Since(start).Milliseconds()))
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
					metrics.AddTxsMetrics.Set(float64(time.Since(start).Milliseconds()))
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
						metrics.AddTxDetailsMetrics.Set(float64(time.Since(start).Milliseconds()))
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

	metrics.SaveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
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
		logx.Errorf("finalSaveBlockDataFunc failed:%s,blockHeight:%d", err.Error(), blockStates.Block.BlockHeight)
		panic("finalSaveBlockDataFunc failed: " + err.Error())
	}
	c.bc.Statedb.UpdateMaxPoolTxIdFinished(blockStates.Block.Txs[len(blockStates.Block.Txs)-1].PoolTxId)
	metrics.L2BlockDbHeightMetric.Set(float64(blockStates.Block.BlockHeight))
	metrics.FinalSaveBlockDataMetrics.WithLabelValues("all").Set(float64(time.Since(start).Milliseconds()))
}

func (c *Committer) createNewBlock(curBlock *block.Block) error {
	return c.bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
		return c.bc.BlockModel.CreateBlockInTransact(dbTx, curBlock)
	})
}

func (c *Committer) shouldCommit(curBlock *block.Block) bool {
	if c.bc.Statedb.NeedRestoreExecutedTxs() {
		if len(c.bc.Statedb.Txs) >= c.maxTxsPerBlock {
			return true
		}
		return false
	}
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
	return latestTx.L1RequestId, nil
}

func (c *Committer) loadAllAccounts() {
	start := time.Now()
	logx.Infof("load all accounts start")
	totalTask := 0
	errChan := make(chan error, 1)
	defer close(errChan)
	batchReloadSize := 1000
	maxAccountIndex, err := c.bc.AccountModel.GetMaxAccountIndex()
	if err != nil && err != types.DbErrNotFound {
		logx.Severef("load all accounts failed:%s", err.Error())
		panic("load all accounts failed: " + err.Error())
	}
	if maxAccountIndex == -1 {
		return
	}
	for i := 0; int64(i) <= maxAccountIndex; i += batchReloadSize {
		toAccountIndex := int64(i+batchReloadSize) - 1
		if toAccountIndex > maxAccountIndex {
			toAccountIndex = maxAccountIndex
		}
		totalTask++
		err := func(fromAccountIndex int64, toAccountIndex int64) error {
			return c.pool.Submit(func() {
				start := time.Now()
				accounts, err := c.bc.AccountModel.GetByAccountIndexRange(fromAccountIndex, toAccountIndex)
				if err != nil && err != types.DbErrNotFound {
					logx.Severef("load all accounts failed:%s", err.Error())
					errChan <- err
					return
				}
				if accounts != nil {
					for _, accountInfo := range accounts {
						formatAccount, err := chain.ToFormatAccountInfo(accountInfo)
						if err != nil {
							logx.Severef("load all accounts failed:%s", err.Error())
							errChan <- err
							return
						}
						c.bc.Statedb.AccountCache.Add(accountInfo.AccountIndex, formatAccount)
					}
				}
				logx.Infof("GetByNftIndexRange cost time %s", float64(time.Since(start).Milliseconds()))
				errChan <- nil
			})
		}(int64(i), toAccountIndex)
		if err != nil {
			logx.Severef("load all accounts failed:%s", err.Error())
			panic("load all accounts failed: " + err.Error())
		}
	}

	for i := 0; i < totalTask; i++ {
		err := <-errChan
		if err != nil {
			logx.Severef("load all accounts failed:%s", err.Error())
			panic("load all accounts failed: " + err.Error())
		}
	}
	logx.Infof("load all accounts end. cost time %s", float64(time.Since(start).Milliseconds()))
}

func (c *Committer) loadAllNfts() {
	start := time.Now()
	logx.Infof("load all nfts start")
	totalTask := 0
	errChan := make(chan error, 1)
	defer close(errChan)
	batchReloadSize := 1000
	maxNftIndex, err := c.bc.L2NftModel.GetMaxNftIndex()
	if err != nil && err != types.DbErrNotFound {
		logx.Severef("load all nfts failed:%s", err.Error())
		panic("load all nfts failed: " + err.Error())
	}
	if maxNftIndex == -1 {
		return
	}
	for i := 0; int64(i) <= maxNftIndex; i += batchReloadSize {
		toNftIndex := int64(i+batchReloadSize) - 1
		if toNftIndex > maxNftIndex {
			toNftIndex = maxNftIndex
		}
		totalTask++
		err := func(fromNftIndex int64, toNftIndex int64) error {
			return c.pool.Submit(func() {
				start := time.Now()
				nfts, err := c.bc.L2NftModel.GetByNftIndexRange(fromNftIndex, toNftIndex)
				if err != nil && err != types.DbErrNotFound {
					logx.Severef("load all nfts failed:%s", err.Error())
					errChan <- err
					return
				}
				if nfts != nil {
					for _, nftInfo := range nfts {
						c.bc.Statedb.NftCache.Add(nftInfo.NftIndex, nftInfo)
					}
				}
				logx.Infof("GetByNftIndexRange cost time %s", float64(time.Since(start).Milliseconds()))
				errChan <- nil
			})
		}(int64(i), toNftIndex)
		if err != nil {
			logx.Severef("load all nfts failed:%s", err.Error())
			panic("load all nfts failed: " + err.Error())
		}
	}

	for i := 0; i < totalTask; i++ {
		err := <-errChan
		if err != nil {
			logx.Severef("load all nfts failed:%s", err.Error())
			panic("load all nfts failed: " + err.Error())
		}
	}
	logx.Infof("load all nfts end. cost time %s", float64(time.Since(start).Milliseconds()))
}

func (c *Committer) PendingTxNum() {
	txStatuses := []int64{tx.StatusPending}
	pendingTxCount, _ := c.bc.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
	metrics.PendingTxNumMetrics.Set(float64(pendingTxCount))
}

func (c *Committer) CompensatePendingPoolTx() {
	fromCreatAt := time.Now().Add(time.Duration(-10) * time.Minute).UnixMilli()
	pendingTxs, err := c.bc.TxPoolModel.GetTxsByStatusAndCreateTime(tx.StatusPending, time.UnixMilli(fromCreatAt), c.bc.Statedb.GetMaxPoolTxIdFinished())
	if err != nil {
		logx.Errorf("get pending transactions from tx pool for compensation failed:%s", err.Error())
		return
	}
	if pendingTxs != nil {
		for _, poolTx := range pendingTxs {
			logx.Severef("get pending transactions from tx pool for compensation id:%s", poolTx.ID)
			_, found := c.bc.Statedb.MemCache.Get(dbcache.PendingPoolTxKeyByPoolTxId(poolTx.ID))
			if found {
				logx.Infof("add pool tx to the queue repeatedly in the compensation task id:%s", poolTx.ID)
				continue
			}
			c.bc.Statedb.MemCache.SetWithTTL(dbcache.PendingPoolTxKeyByPoolTxId(poolTx.ID), poolTx.ID, 0, time.Duration(1)*time.Hour)
			c.executeTxWorker.Enqueue(poolTx)
		}
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

func (c *Committer) SendIpfsServer() error {
	historiesIpfs, err := c.bc.L2NftMetadataHistoryModel.GetL2NftMetadataHistory(nft.NotConfirmed)
	if err != nil {
		if err == types.DbErrSqlOperation {
			return err
		}
	}
	if historiesIpfs != nil {
		for _, history := range historiesIpfs {
			if history.NftIndex == types.NilNftIndex {
				poolTx, err := c.bc.TxPoolModel.GetTxUnscopedByTxHash(history.TxHash)
				if err != nil {
					return err
				}
				if poolTx.TxStatus == tx.StatusFailed {
					err = c.bc.L2NftMetadataHistoryModel.DeleteInTransact(history.ID)
					if err != nil {
						return err
					}
					continue
				} else if poolTx.TxStatus == tx.StatusExecuted {
					tx, err := c.bc.TxModel.GetTxByHash(history.TxHash)
					if err != nil {
						return err
					}
					history.NftIndex = tx.NftIndex
					err = saveIpfs(history)
					if err != nil {
						return err
					}
					err = c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
						err = c.bc.L2NftMetadataHistoryModel.UpdateL2NftMetadataHistoryInTransact(tx, history)
						if err != nil {
							return err
						}
						err = c.bc.L2NftModel.UpdateIpfsStatusByNftIndexInTransact(tx, history.NftIndex)
						return err
					})
					if err != nil {
						return err
					}
				}
			} else {
				err = saveIpfs(history)
				if err != nil {
					return err
				}
				err = c.bc.L2NftMetadataHistoryModel.UpdateL2NftMetadataHistoryInTransact(c.bc.DB().DB, history)
				if err != nil {
					return err
				}
			}
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
	history.Cid = cid
	return nil
}

func (c *Committer) RefreshServer() error {
	historiesIpns, err := c.bc.L2NftMetadataHistoryModel.GetL2NftMetadataHistory(nft.Confirmed)
	if err != nil {
		if err == types.DbErrSqlOperation {
			return err
		}
	}
	if historiesIpns != nil {
		for _, hostory := range historiesIpns {
			_, err = common.Ipfs.PublishWithDetails(hostory.Cid, hostory.IpnsName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
