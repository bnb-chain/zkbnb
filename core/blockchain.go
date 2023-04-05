package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb/common/metrics"
	"github.com/bnb-chain/zkbnb/tools/desertexit/config"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon"
	"github.com/dgraph-io/ristretto"
	"gorm.io/plugin/dbresolver"
	"math/big"
	"time"

	"gorm.io/gorm/logger"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core/statedb"
	sdb "github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/rollback"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

type ChainConfig struct {
	Postgres struct {
		MasterDataSource string
		SlaveDataSource  string
		LogLevel         logger.LogLevel `json:",optional"`
	}
	CacheRedis cache.CacheConf
	//second
	RedisExpiration int `json:",optional"`
	//nolint:staticcheck
	CacheConfig statedb.CacheConfig `json:",optional"`
	TreeDB      struct {
		Driver tree.Driver
		//nolint:staticcheck
		LevelDBOption tree.LevelDBOption `json:",optional"`
		//nolint:staticcheck
		RedisDBOption tree.RedisDBOption `json:",optional"`
		//nolint:staticcheck
		RoutinePoolSize    int `json:",optional"`
		AssetTreeCacheSize int
	}
}

type BlockChain struct {
	*sdb.ChainDB
	Statedb *sdb.StateDB // Cache for current block changes.

	chainConfig *ChainConfig
	dryRun      bool //dryRun mode is used for verifying user inputs, is not for execution

	currentBlock     *block.Block
	rollbackBlockMap map[int64]*block.Block
	processor        Processor
}

func NewBlockChain(config *ChainConfig, moduleName string) (*BlockChain, error) {
	masterDataSource := config.Postgres.MasterDataSource
	slaveDataSource := config.Postgres.SlaveDataSource
	db, err := gorm.Open(postgres.Open(config.Postgres.MasterDataSource), &gorm.Config{
		Logger: logger.Default.LogMode(config.Postgres.LogLevel),
	})
	if err != nil {
		logx.Severe("gorm connect db failed: ", err)
		return nil, err
	}
	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDataSource)},
		Replicas: []gorm.Dialector{postgres.Open(slaveDataSource)},
	}))
	if err != nil {
		logx.Severe("gorm connect db failed: ", err)
		return nil, err
	}

	bc := &BlockChain{
		ChainDB:     sdb.NewChainDB(db),
		chainConfig: config,
	}
	expiration := dbcache.RedisExpiration
	if config.RedisExpiration != 0 {
		expiration = time.Duration(config.RedisExpiration) * time.Second
	}
	redisCache := dbcache.NewRedisCache(config.CacheRedis[0].Host, config.CacheRedis[0].Pass, expiration)
	treeCtx, err := tree.NewContext(moduleName, config.TreeDB.Driver, false, false, config.TreeDB.RoutinePoolSize, &config.TreeDB.LevelDBOption, &config.TreeDB.RedisDBOption)
	if err != nil {
		return nil, err
	}
	treeCtx.SetOptions(bsmt.BatchSizeLimit(3 * 1024 * 1024))

	var statuses = []int{block.StatusPending, block.StatusCommitted, block.StatusVerifiedAndExecuted}
	curHeight, err := bc.BlockModel.GetLatestHeight(statuses)
	if err != nil {
		return nil, fmt.Errorf("get latest pending or committed or verified height failed: %v", err)
	}
	logx.Infof("get latest pending or committed or verified height: %d", curHeight)

	bc.currentBlock, err = bc.BlockModel.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, fmt.Errorf("get block by height failed: %v,curHeight=%d", err.Error(), curHeight)
	}

	accountIndexList, nftIndexList, heights, err := preRollBackFunc(bc, redisCache)
	if err != nil {
		return nil, fmt.Errorf("pre rollBack func failed: %v,curHeight=%d", err.Error(), curHeight)
	}

	bc.Statedb, err = sdb.NewStateDB(treeCtx, bc.ChainDB, redisCache, &config.CacheConfig, config.TreeDB.AssetTreeCacheSize, bc.currentBlock.StateRoot, accountIndexList, curHeight)
	if err != nil {
		return nil, err
	}
	bc.rollbackBlockMap = make(map[int64]*block.Block, 0)

	if len(heights) != 0 {
		err = rollbackFunc(bc, accountIndexList, nftIndexList, heights, curHeight)
		if err != nil {
			return nil, err
		}
		rollBackBlocks, err := bc.BlockModel.GetBlockByStatus([]int{block.StatusProposing})
		if err != nil && err != types.DbErrNotFound {
			return nil, fmt.Errorf("get blocks by status (StatusProposing,StatusPacked) failed: %s", err.Error())
		}
		if rollBackBlocks != nil {
			for _, rollBackBlock := range rollBackBlocks {
				bc.rollbackBlockMap[rollBackBlock.BlockHeight] = rollBackBlock
			}
		}
	}

	err = verifyRollbackTableDataFunc(bc, curHeight)
	if err != nil {
		return nil, err
	}

	err = verifyRollbackTreesFunc(bc, bc.currentBlock)
	if err != nil {
		return nil, err
	}

	bc.Statedb.PreviousStateRootImmutable = bc.currentBlock.StateRoot

	mintNft, err := bc.TxPoolModel.GetLatestMintNft()
	if err != nil && err != types.DbErrNotFound {
		return nil, fmt.Errorf("get latest mint nft failed:%s", err.Error())
	}
	bc.Statedb.UpdateNftIndex(types.NilNftIndex)
	if mintNft != nil {
		bc.Statedb.UpdateNftIndex(mintNft.NftIndex)
	}

	poolTx, err := bc.TxPoolModel.GetLatestAccountIndex()
	if err != nil && err != types.DbErrNotFound {
		return nil, fmt.Errorf("get latest account index failed:%s", err.Error())
	}
	bc.Statedb.UpdateAccountIndex(types.NilAccountIndex)
	if poolTx != nil {
		bc.Statedb.UpdateAccountIndex(poolTx.AccountIndex)
	}

	latestRollback, err := bc.TxPoolModel.GetLatestRollback(tx.StatusPending, true)
	if err != nil && err != types.DbErrNotFound {
		return nil, fmt.Errorf("get latest pool tx rollback failed:%s", err.Error())
	}
	maxPollTxIdRollback := uint(0)
	if latestRollback != nil {
		maxPollTxIdRollback = latestRollback.ID
	}
	bc.Statedb.MaxPollTxIdRollbackImmutable = maxPollTxIdRollback
	bc.Statedb.UpdateNeedRestoreExecutedTxs(maxPollTxIdRollback > 0)

	err = metrics.InitBlockChainMetrics()
	if err != nil {
		return nil, err
	}
	bc.processor = NewCommitProcessor(bc)
	return bc, nil
}

// NewBlockChainForDryRun - for dry run mode, we can reuse existing models for quick creation
// , e.g., for sending tx, we can create blockchain for each request quickly
func NewBlockChainForDryRun(accountModel account.AccountModel,
	nftModel nft.L2NftModel, txPoolModel tx.TxPoolModel, assetModel asset.AssetModel,
	sysConfigModel sysconfig.SysConfigModel, redisCache dbcache.Cache, memCache *ristretto.Cache) (*BlockChain, error) {
	chainDb := &sdb.ChainDB{
		AccountModel:     accountModel,
		L2NftModel:       nftModel,
		TxPoolModel:      txPoolModel,
		L2AssetInfoModel: assetModel,
		SysConfigModel:   sysConfigModel,
	}
	statedb, err := sdb.NewStateDBForDryRun(redisCache, &statedb.DefaultCacheConfig, chainDb, memCache)
	if err != nil {
		return nil, err
	}
	bc := &BlockChain{
		ChainDB: chainDb,
		dryRun:  true,
		Statedb: statedb,
	}
	bc.processor = NewAPIProcessor(bc)
	return bc, nil
}

func NewBlockChainForDesertExit(config *config.Config) (*BlockChain, error) {
	masterDataSource := config.Postgres.MasterDataSource
	db, err := gorm.Open(postgres.Open(config.Postgres.MasterDataSource), &gorm.Config{
		Logger: logger.Default.LogMode(config.Postgres.LogLevel),
	})
	if err != nil {
		logx.Severe("gorm connect db failed: ", err)
		return nil, err
	}
	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{postgres.Open(masterDataSource)},
	}))
	if err != nil {
		logx.Severe("gorm connect db failed: ", err)
		return nil, err
	}

	bc := &BlockChain{
		dryRun:       false,
		ChainDB:      sdb.NewChainDB(db),
		currentBlock: &block.Block{},
	}
	bc.Statedb, err = sdb.NewStateDBForDesertExit(nil, &config.CacheConfig, bc.ChainDB)
	if err != nil {
		return nil, err
	}
	return bc, nil
}

func preRollBackFunc(bc *BlockChain, redisCache dbcache.Cache) ([]int64, []int64, []int64, error) {
	accountIndexList := make([]int64, 0)
	nftIndexList := make([]int64, 0)
	heights := make([]int64, 0)

	curHeight, err := bc.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get current block height failed: %s", err.Error())
	}
	logx.Infof("get current block height: %d", curHeight)

	blocks, err := bc.BlockModel.GetBlockByStatus([]int{block.StatusProposing, block.StatusPacked})
	if err != nil && err != types.DbErrNotFound {
		return nil, nil, nil, fmt.Errorf("get blocks by status (StatusProposing,StatusPacked) failed: %s", err.Error())
	}
	if blocks == nil {
		return accountIndexList, nftIndexList, heights, nil
	}

	accountIndexMap := make(map[int64]bool, 0)
	nftIndexMap := make(map[int64]bool, 0)

	for _, blockInfo := range blocks {
		if blockInfo.AccountIndexes != "[]" && blockInfo.AccountIndexes != "" {
			var accountIndexes []int64
			err = json.Unmarshal([]byte(blockInfo.AccountIndexes), &accountIndexes)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("json err unmarshal failed: %s", err.Error())
			}
			for _, accountIndex := range accountIndexes {
				accountIndexMap[accountIndex] = true
			}
		}

		if blockInfo.NftIndexes != "[]" && blockInfo.NftIndexes != "" {
			var nftIndexes []int64
			err = json.Unmarshal([]byte(blockInfo.NftIndexes), &nftIndexes)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("json err unmarshal failed: %s", err.Error())
			}
			for _, nftIndex := range nftIndexes {
				nftIndexMap[nftIndex] = true
			}
		}

		heights = append(heights, blockInfo.BlockHeight)
	}
	if len(heights) == 0 {
		return accountIndexList, nftIndexList, heights, nil
	}

	for k := range accountIndexMap {
		accountIndexList = append(accountIndexList, k)
	}

	for k := range nftIndexMap {
		nftIndexList = append(nftIndexList, k)
	}

	for _, accountIndex := range accountIndexList {
		var redisAccount interface{}
		accountInfo := &account.Account{}
		redisAccount, err := redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex), accountInfo)
		if err == nil && redisAccount != nil {
			err := redisCache.Delete(context.Background(), dbcache.AccountKeyByL1Address(accountInfo.L1Address))
			if err != nil {
				logx.Severef("delete redis failed: %v,accountIndex=%v", err, accountIndex)
			}
		}

		err = redisCache.Delete(context.Background(), dbcache.AccountKeyByIndex(accountIndex))
		if err != nil {
			logx.Severef("delete redis failed: %v,accountIndex=%v", err, accountIndex)
		}
	}

	for _, nftIndex := range nftIndexList {
		err := redisCache.Delete(context.Background(), dbcache.NftKeyByIndex(nftIndex))
		if err != nil {
			logx.Severef("cache to redis failed: %v,nftIndex=%v", err, nftIndex)
		}
	}

	return accountIndexList, nftIndexList, heights, nil
}

func rollbackFunc(bc *BlockChain, accountIndexList []int64, nftIndexList []int64, heights []int64, curHeight int64) (err error) {
	accountIndexSlice := make([]int64, 0)
	accountHistories := make([]*account.AccountHistory, 0)
	accountIndexLen := len(accountIndexList)

	for _, accountIndex := range accountIndexList {
		accountIndexLen--
		accountIndexSlice = append(accountIndexSlice, accountIndex)
		if len(accountIndexSlice) == 100 || accountIndexLen == 0 {
			_, accountHistoryList, err := bc.AccountHistoryModel.GetLatestAccountHistories(accountIndexSlice, curHeight)
			if err != nil && err != types.DbErrNotFound {
				return fmt.Errorf("get latest account histories failed: %s", err.Error())
			}
			if accountHistoryList != nil {
				accountHistories = append(accountHistories, accountHistoryList...)
			}
			accountIndexSlice = make([]int64, 0)
		}
	}

	deleteAccountIndexMap := make(map[int64]bool, 0)
	for _, accountIndex := range accountIndexList {
		findAccountIndex := false
		for _, accountHistory := range accountHistories {
			if accountIndex == accountHistory.AccountIndex {
				findAccountIndex = true
				break
			}
		}
		if findAccountIndex == false {
			deleteAccountIndexMap[accountIndex] = true
		}
	}

	nftIndexSlice := make([]int64, 0)
	nftHistories := make([]*nft.L2NftHistory, 0)
	nftIndexLen := len(nftIndexList)
	for _, nftIndex := range nftIndexList {
		nftIndexLen--
		nftIndexSlice = append(nftIndexSlice, nftIndex)
		if len(nftIndexSlice) == 100 || nftIndexLen == 0 {
			_, nftHistoryList, err := bc.L2NftHistoryModel.GetLatestNftHistories(nftIndexSlice, curHeight)
			if err != nil && err != types.DbErrNotFound {
				return fmt.Errorf("get latest nft histories failed: %s", err.Error())
			}
			if nftHistoryList != nil {
				nftHistories = append(nftHistories, nftHistoryList...)
			}
			nftIndexSlice = make([]int64, 0)
		}
	}

	deleteNftIndexMap := make(map[int64]bool, 0)
	for _, nftIndex := range nftIndexList {
		findNftIndex := false
		for _, nftHistory := range nftHistories {
			if nftIndex == nftHistory.NftIndex {
				findNftIndex = true
				break
			}
		}
		if findNftIndex == false {
			deleteNftIndexMap[nftIndex] = true
		}
	}

	txs, err := bc.TxPoolModel.GetTxsUnscopedByHeights(heights)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get pool txs by heights failed: %s", err.Error())
	}

	err = bc.DB().DB.Transaction(func(dbTx *gorm.DB) error {
		logx.Info("roll back account start")
		for _, accountHistory := range accountHistories {
			if deleteAccountIndexMap[accountHistory.AccountIndex] {
				continue
			}
			accountInfo := &account.Account{
				AccountIndex:    accountHistory.AccountIndex,
				Nonce:           accountHistory.Nonce,
				CollectionNonce: accountHistory.CollectionNonce,
				AssetInfo:       accountHistory.AssetInfo,
				AssetRoot:       accountHistory.AssetRoot,
				L2BlockHeight:   accountHistory.L2BlockHeight,
				Status:          accountHistory.Status,
				PublicKey:       accountHistory.PublicKey,
			}
			err := bc.AccountModel.UpdateByIndexInTransact(dbTx, accountInfo)
			if err != nil {
				return fmt.Errorf("roll back account failed: %s", err.Error())
			}
		}

		logx.Info("roll back account,delete account start")
		if len(deleteAccountIndexMap) > 0 {
			deleteAccountIndexList := make([]int64, 0)
			for k := range deleteAccountIndexMap {
				deleteAccountIndexList = append(deleteAccountIndexList, k)
			}
			err := bc.AccountModel.DeleteByIndexesInTransact(dbTx, deleteAccountIndexList)
			if err != nil {
				return fmt.Errorf("roll back account,delete account failed: %s", err.Error())
			}
		}

		logx.Info("roll back nft start")
		for _, nftHistory := range nftHistories {
			if deleteNftIndexMap[nftHistory.NftIndex] {
				continue
			}
			nftInfo := &nft.L2Nft{
				NftIndex:            nftHistory.NftIndex,
				OwnerAccountIndex:   nftHistory.OwnerAccountIndex,
				NftContentHash:      nftHistory.NftContentHash,
				CollectionId:        nftHistory.CollectionId,
				RoyaltyRate:         nftHistory.RoyaltyRate,
				CreatorAccountIndex: nftHistory.CreatorAccountIndex,
				L2BlockHeight:       nftHistory.L2BlockHeight,
				NftContentType:      nftHistory.NftContentType,
			}
			err := bc.L2NftModel.UpdateByIndexInTransact(dbTx, nftInfo)
			if err != nil {
				return fmt.Errorf("roll back nft failed: %s", err.Error())
			}
		}

		logx.Info("roll back nft,delete nft start")
		if len(deleteNftIndexMap) > 0 {
			deleteNftIndexList := make([]int64, 0)
			for k := range deleteNftIndexMap {
				deleteNftIndexList = append(deleteNftIndexList, k)
			}
			err := bc.L2NftModel.DeleteByIndexesInTransact(dbTx, deleteNftIndexList)
			if err != nil {
				return fmt.Errorf("roll back nft,delete nft failed: %s", err.Error())
			}
		}

		logx.Info("roll back account history,delete account start")
		err := bc.AccountHistoryModel.DeleteByHeightsInTransact(dbTx, heights)
		if err != nil {
			return fmt.Errorf("roll back account history,delete account history failed: %s", err.Error())
		}

		logx.Info("roll back l2nft history,delete l2nft history start")
		err = bc.L2NftHistoryModel.DeleteByHeightsInTransact(dbTx, heights)
		if err != nil {
			return fmt.Errorf("roll back account l2nft,delete l2nft history failed: %s", err.Error())
		}

		logx.Info("roll back tx detail start")
		err = bc.TxDetailModel.DeleteByHeightsInTransact(dbTx, heights)
		if err != nil {
			return fmt.Errorf("roll back tx detail failed: %s", err.Error())
		}

		logx.Info("roll back tx start")
		err = bc.TxModel.DeleteByHeightsInTransact(dbTx, heights)
		if err != nil {
			return fmt.Errorf("roll back tx failed: %s", err.Error())
		}

		logx.Info("roll back block start")
		err = bc.BlockModel.UpdateBlockToProposingInTransact(dbTx, heights)
		if err != nil {
			return fmt.Errorf("roll back block failed: %s", err.Error())
		}

		logx.Info("roll back compressed block start")
		err = bc.CompressedBlockModel.DeleteByHeightsInTransact(dbTx, heights)
		if err != nil {
			return fmt.Errorf("roll back compressed block failed: %s", err.Error())
		}

		logx.Info("roll back pool tx start")
		err = bc.TxPoolModel.UpdateTxsToPendingByHeights(dbTx, heights)
		if err != nil {
			return fmt.Errorf("roll back pool tx step 2 failed: %s", err.Error())
		}

		heightsJson, err := json.Marshal(heights)
		if err != nil {
			return fmt.Errorf("unmarshal heights failed: %s", err.Error())
		}
		nftIndexListJson, err := json.Marshal(nftIndexList)
		if err != nil {
			return fmt.Errorf("unmarshal nftIndexList failed: %s", err.Error())
		}
		accountIndexListJson, err := json.Marshal(accountIndexList)
		if err != nil {
			return fmt.Errorf("unmarshal accountIndexList failed: %s", err.Error())
		}
		pollTxIds := make([]uint, 0)
		for _, txInfo := range txs {
			pollTxIds = append(pollTxIds, txInfo.ID)
		}
		pollTxIdsJson, err := json.Marshal(pollTxIds)
		if err != nil {
			return fmt.Errorf("unmarshal pollTxIds failed: %s", err.Error())

		}

		fromTxHash := ""
		fromPoolTxId := uint(0)
		if len(txs) > 0 {
			fromTxHash = txs[0].TxHash
			fromPoolTxId = txs[0].ID
		}
		rollbackInfo := &rollback.Rollback{FromTxHash: fromTxHash, FromPoolTxId: fromPoolTxId, FromBlockHeight: heights[0], PoolTxIds: string(pollTxIdsJson), BlockHeights: string(heightsJson), AccountIndexes: string(accountIndexListJson), NftIndexes: string(nftIndexListJson)}
		err = bc.RollbackModel.CreateInTransact(dbTx, rollbackInfo)
		if err != nil {
			return fmt.Errorf("create rollback failed: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func verifyRollbackTableDataFunc(bc *BlockChain, curHeight int64) error {
	logx.Infof("verify rollback start,height:%s", curHeight)
	blocks, err := bc.BlockModel.GetBlockByStatus([]int{block.StatusPacked})
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get statusPacked block height failed: %s", err.Error())
	}
	if blocks != nil {
		return fmt.Errorf("there are some block which status is statusPacked")
	}

	count, err := bc.TxPoolModel.GetCountByGreaterHeight(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get count from TxPool by greater height failed: %s", err.Error())
	}
	if count > 0 {
		return fmt.Errorf("roll back for TxPool failed")
	}

	count, err = bc.TxModel.GetCountByGreaterHeight(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get count from Tx by greater height failed: %s", err.Error())
	}
	if count > 0 {
		return fmt.Errorf("roll back for Tx failed")
	}

	count, err = bc.TxDetailModel.GetCountByGreaterHeight(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get count from TxDetail by greater height failed: %s", err.Error())
	}
	if count > 0 {
		return fmt.Errorf("roll back for TxDetail failed")
	}

	count, err = bc.AccountModel.GetCountByGreaterHeight(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get count from Account by greater height failed: %s", err.Error())
	}
	if count > 0 {
		return fmt.Errorf("roll back for Account failed")
	}

	count, err = bc.AccountHistoryModel.GetCountByGreaterHeight(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get count from AccountHistory by greater height failed: %s", err.Error())
	}
	if count > 0 {
		return fmt.Errorf("roll back for AccountHistory failed")
	}

	count, err = bc.L2NftModel.GetCountByGreaterHeight(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get count from L2Nft by greater height failed: %s", err.Error())
	}
	if count > 0 {
		return fmt.Errorf("roll back for L2Nft failed")
	}

	count, err = bc.L2NftHistoryModel.GetCountByGreaterHeight(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get count by greater from L2NftHistory height failed: %s", err.Error())
	}
	if count > 0 {
		return fmt.Errorf("roll back for L2NftHistory failed")
	}

	count, err = bc.CompressedBlockModel.GetCountByGreaterHeight(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get count from CompressedBlock by greater height failed: %s", err.Error())
	}
	if count > 0 {
		return fmt.Errorf("roll back for CompressedBlock failed")
	}

	maxAccountIndex, err := bc.AccountModel.GetMaxAccountIndex()
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("GetMaxAccountIndex from Account failed: %s", err.Error())
	}

	maxAccountHistoryIndex, err := bc.AccountHistoryModel.GetMaxAccountIndex(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("GetMaxAccountIndex from AccountHistory failed: %s", err.Error())
	}
	if maxAccountIndex != maxAccountHistoryIndex {
		return fmt.Errorf("maxAccountIndex=%d not equal maxAccountHistoryIndex=%d failed", maxAccountIndex, maxAccountHistoryIndex)
	}
	maxNftIndex, err := bc.L2NftModel.GetMaxNftIndex()
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("GetMaxNftIndex from Nft failed: %s", err.Error())
	}

	maxNftHistoryIndex, err := bc.L2NftHistoryModel.GetMaxNftIndex(curHeight)
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("GetMaxNftIndex from NftHistory failed: %s", err.Error())
	}
	if maxNftIndex != maxNftHistoryIndex {
		return fmt.Errorf("maxNftIndex=%d not equal maxNftHistoryIndex=%d failed", maxNftIndex, maxNftHistoryIndex)
	}
	return nil
}

func verifyRollbackTreesFunc(bc *BlockChain, currentBlock *block.Block) error {
	hFunc := poseidon.NewPoseidon()
	hFunc.Write(bc.Statedb.AccountTree.Root())
	hFunc.Write(bc.Statedb.NftTree.Root())
	stateRoot := common.Bytes2Hex(hFunc.Sum(nil))
	logx.Infof("smt tree stateRoot=%s equal currentBlock.StateRoot=%s,height:%s,AccountTree.Root=%s,NftTree.Root=%s", stateRoot, currentBlock.StateRoot, currentBlock.BlockHeight, common.Bytes2Hex(bc.Statedb.AccountTree.Root()), common.Bytes2Hex(bc.Statedb.NftTree.Root()))
	if stateRoot != currentBlock.StateRoot {
		return fmt.Errorf("smt tree stateRoot=%s not equal currentBlock.StateRoot=%s,height:%d,AccountTree.Root=%s,NftTree.Root=%s", stateRoot, currentBlock.StateRoot, currentBlock.BlockHeight, common.Bytes2Hex(bc.Statedb.AccountTree.Root()), common.Bytes2Hex(bc.Statedb.NftTree.Root()))
	}
	return nil
}

func (bc *BlockChain) ApplyTransaction(tx *tx.Tx) error {
	return bc.processor.Process(tx)
}

func (bc *BlockChain) InitNewBlock() (*block.Block, error) {
	newBlock := bc.rollbackBlockMap[bc.currentBlock.BlockHeight+1]
	if newBlock == nil {
		newBlock = &block.Block{
			Model: gorm.Model{
				// The block timestamp will be set when the first transaction executed.
				CreatedAt: time.Time{},
			},
			BlockHeight: bc.currentBlock.BlockHeight + 1,
			StateRoot:   bc.currentBlock.StateRoot,
			BlockStatus: block.StatusProposing,
		}
	}
	bc.currentBlock = newBlock
	bc.Statedb.PurgeCache(bc.currentBlock.StateRoot)
	err := bc.Statedb.MarkGasAccountAsPending()
	return newBlock, err
}

func (bc *BlockChain) CurrentBlock() *block.Block {
	return bc.currentBlock
}

func (bc *BlockChain) ClearRollbackBlockMap() {
	bc.rollbackBlockMap = make(map[int64]*block.Block, 0)
}

// UpdateAssetTree compute account asset hash, commit asset smt,compute account leaf hash, compute nft leaf hash
func (bc *BlockChain) UpdateAssetTree(stateDataCopy *statedb.StateDataCopy) error {
	start := time.Now()
	// Intermediate state root.
	err := bc.Statedb.UpdateAssetTree(stateDataCopy)
	if err != nil {
		return err
	}
	metrics.UpdateAssetTreeMetrics.Set(float64(time.Since(start).Milliseconds()))
	return nil
}

// UpdateAccountAndNftTree multi set account tree with version,multi set nft tree with version
// commit account and nft tree
// build Block CompressedBlock PendingAccount PendingAccountHistory PendingNft PendingNftHistory
func (bc *BlockChain) UpdateAccountAndNftTree(blockSize int, stateDataCopy *statedb.StateDataCopy) (*block.BlockStates, error) {
	newBlock := stateDataCopy.CurrentBlock
	err := bc.Statedb.SetAccountAndNftTree(stateDataCopy)
	if err != nil {
		return nil, err
	}

	// Align pub data.
	bc.Statedb.AlignPubData(blockSize, stateDataCopy)

	commitment := chain.CreateBlockCommitment(newBlock.BlockHeight, newBlock.CreatedAt.UnixMilli(),
		common.FromHex(bc.Statedb.PreviousStateRootImmutable), common.FromHex(stateDataCopy.StateCache.StateRoot),
		stateDataCopy.StateCache.PubData, int64(len(stateDataCopy.StateCache.PubDataOffset)))

	newBlock.BlockSize = uint16(blockSize)
	newBlock.BlockCommitment = commitment
	newBlock.StateRoot = stateDataCopy.StateCache.StateRoot
	newBlock.PriorityOperations = stateDataCopy.StateCache.PriorityOperations
	newBlock.PendingOnChainOperationsHash = common.Bytes2Hex(stateDataCopy.StateCache.PendingOnChainOperationsHash)
	newBlock.Txs = stateDataCopy.StateCache.Txs
	for _, executedTx := range newBlock.Txs {
		executedTx.TxStatus = tx.StatusPacked
	}
	if len(stateDataCopy.StateCache.PendingOnChainOperationsPubData) > 0 {
		onChainOperationsPubDataBytes, err := json.Marshal(stateDataCopy.StateCache.PendingOnChainOperationsPubData)
		if err != nil {
			return nil, fmt.Errorf("marshal pending onChain operation pubData failed: %v", err)
		}
		newBlock.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
	}

	offsetBytes, err := json.Marshal(stateDataCopy.StateCache.PubDataOffset)
	if err != nil {
		return nil, fmt.Errorf("marshal pubData offset failed: %v", err)
	}

	newCompressedBlock := &compressedblock.CompressedBlock{
		BlockSize:         uint16(blockSize),
		BlockHeight:       newBlock.BlockHeight,
		StateRoot:         newBlock.StateRoot,
		PublicData:        common.Bytes2Hex(stateDataCopy.StateCache.PubData),
		Timestamp:         newBlock.CreatedAt.UnixMilli(),
		PublicDataOffsets: string(offsetBytes),
		RealBlockSize:     uint16(len(stateDataCopy.StateCache.Txs)),
	}
	bc.Statedb.PreviousStateRootImmutable = stateDataCopy.StateCache.StateRoot
	currentHeight := stateDataCopy.CurrentBlock.BlockHeight

	start := time.Now()
	logx.Infof("CommitAccountAndNftTree,latestVersion=%d,prunedBlockHeight=%d", uint64(bc.Statedb.AccountTree.LatestVersion()), uint64(bc.StateDB().GetPrunedBlockHeight()))
	prunedVersion := bc.StateDB().GetPrunedBlockHeight()
	err = tree.CommitAccountAndNftTree(uint64(prunedVersion), stateDataCopy.CurrentBlock.BlockHeight, bc.Statedb.AccountTree, bc.Statedb.NftTree)
	if err != nil {
		return nil, err
	}

	metrics.CommitAccountTreeMetrics.Set(float64(time.Since(start).Milliseconds()))

	pendingAccount, pendingAccountHistory, err := bc.Statedb.GetPendingAccount(currentHeight, stateDataCopy)
	if err != nil {
		return nil, err
	}

	pendingNft, pendingNftHistory, err := bc.Statedb.GetPendingNft(currentHeight, stateDataCopy)
	if err != nil {
		return nil, err
	}
	return &block.BlockStates{
		Block:                 newBlock,
		CompressedBlock:       newCompressedBlock,
		PendingAccount:        pendingAccount,
		PendingAccountHistory: pendingAccountHistory,
		PendingNft:            pendingNft,
		PendingNftHistory:     pendingNftHistory,
	}, nil
}

func (bc *BlockChain) VerifyExpiredAt(expiredAt int64) error {
	if !bc.dryRun {
		if expiredAt < bc.currentBlock.CreatedAt.UnixMilli() {
			return types.AppErrInvalidExpireTime
		}
	} else {
		if expiredAt < time.Now().UnixMilli() {
			return types.AppErrInvalidExpireTime
		}
	}
	return nil
}

func (bc *BlockChain) VerifyNonce(accountIndex int64, nonce int64) error {
	if !bc.dryRun {
		expectNonce, err := bc.Statedb.GetCommittedNonce(accountIndex)
		if err != nil {
			return err
		}
		logx.Infof("committer verify nonce start,accountIndex=%d,nonce=%d,expectNonce=%d", accountIndex, nonce, expectNonce)
		if nonce != expectNonce {
			logx.Infof("committer verify nonce failed,accountIndex=%d,nonce=%d,expectNonce=%d", accountIndex, nonce, expectNonce)
			bc.Statedb.SetPendingNonceToRedisCache(accountIndex, expectNonce-1)
			return types.AppErrInvalidNonce
		} else {
			logx.Infof("committer verify nonce success,accountIndex=%d,nonce=%d,expectNonce=%d", accountIndex, nonce, expectNonce)
		}
	} else {
		pendingNonce, err := bc.Statedb.GetPendingNonceFromCache(accountIndex)
		if err != nil {
			return err
		}
		if pendingNonce != nonce {
			logx.Infof("clear pending nonce from redis cache,accountIndex=%d,pendingNonce=%d,nonce=%d", accountIndex, pendingNonce, nonce)
			bc.Statedb.ClearPendingNonceFromRedisCache(accountIndex)
			return types.AppErrInvalidNonce
		}
	}
	return nil
}

func (bc *BlockChain) VerifyGas(gasAccountIndex, gasFeeAssetId int64, txType int, gasFeeAmount *big.Int, skipGasAmtChk bool) error {
	cfgGasAccountIndex, err := bc.Statedb.GetGasAccountIndex()
	if err != nil {
		return err
	}
	if gasAccountIndex != cfgGasAccountIndex {
		return types.AppErrInvalidGasFeeAccount
	}

	cfgGasFee, err := bc.Statedb.GetGasConfig()
	if err != nil {
		return err
	}

	gasAsset, ok := cfgGasFee[uint32(gasFeeAssetId)]
	if !ok {
		logx.Errorf("cannot find gas config for asset id: %d", gasFeeAssetId)
		return types.AppErrInvalidGasFeeAsset
	}

	if !skipGasAmtChk {
		gasFee, ok := gasAsset[txType]
		if !ok {
			return types.AppErrInvalidTxType
		}
		if gasFeeAmount.Cmp(big.NewInt(gasFee)) < 0 {
			return types.AppErrInvalidGasFeeAmount
		}
	}
	return nil
}

func (bc *BlockChain) StateDB() *sdb.StateDB {
	return bc.Statedb
}

func (bc *BlockChain) DB() *sdb.ChainDB {
	return bc.ChainDB
}

func (bc *BlockChain) setCurrentBlockTimeStamp() {
	if bc.currentBlock.CreatedAt.IsZero() && len(bc.Statedb.Txs) == 0 {
		creatAt := time.Now().UnixMilli()
		bc.currentBlock.CreatedAt = time.UnixMilli(creatAt)
	}
}

func (bc *BlockChain) resetCurrentBlockTimeStamp() {
	if len(bc.Statedb.Txs) > 0 || bc.Statedb.NeedRestoreExecutedTxs() {
		return
	}

	bc.currentBlock.CreatedAt = time.Time{}
}
