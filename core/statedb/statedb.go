package statedb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbnb/common/metrics"
	"github.com/dgraph-io/ristretto"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	lru "github.com/hashicorp/golang-lru"
	"github.com/zeromicro/go-zero/core/logx"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/common/gopool"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

var (
	DefaultCacheConfig = CacheConfig{
		AccountCacheSize: 2048,
		NftCacheSize:     2048,
		MemCacheSize:     204800,
	}
)

type CacheConfig struct {
	AccountCacheSize int
	NftCacheSize     int
	MemCacheSize     int `json:",optional"`
}

func (c *CacheConfig) sanitize() *CacheConfig {
	if c.AccountCacheSize <= 0 {
		c.AccountCacheSize = DefaultCacheConfig.AccountCacheSize
	}

	if c.NftCacheSize <= 0 {
		c.NftCacheSize = DefaultCacheConfig.NftCacheSize
	}

	if c.MemCacheSize <= 0 {
		c.MemCacheSize = DefaultCacheConfig.MemCacheSize
	}

	return c
}

type StateDB struct {
	DryRun bool
	// State cache
	*StateCache
	chainDb    *ChainDB
	redisCache dbcache.Cache

	// Flat state
	AccountCache *lru.Cache
	NftCache     *lru.Cache
	MemCache     *ristretto.Cache

	// Tree state
	AccountTree                  bsmt.SparseMerkleTree
	NftTree                      bsmt.SparseMerkleTree
	AccountAssetTrees            *tree.AssetTreeCache
	TreeCtx                      *tree.Context
	prunedBlockHeight            int64
	prunedBlockHeightLock        sync.RWMutex
	PreviousStateRootImmutable   string
	MaxPollTxIdRollbackImmutable uint

	needRestoreExecutedTxs     bool
	needRestoreExecutedTxsLock sync.RWMutex

	maxPoolTxIdFinished     uint
	maxPoolTxIdFinishedLock sync.RWMutex
	nextNftIndex            int64
	nextNftIndexLock        sync.RWMutex
}

func NewStateDB(treeCtx *tree.Context, chainDb *ChainDB,
	redisCache dbcache.Cache, cacheConfig *CacheConfig, assetCacheSize int,
	stateRoot string, accountIndexList []int64, curHeight int64) (*StateDB, error) {
	err := tree.SetupTreeDB(treeCtx)
	if err != nil {
		logx.Error("setup tree db failed: ", err)
		return nil, err
	}
	accountTree, accountAssetTrees, err := tree.InitAccountTree(
		chainDb.AccountModel,
		chainDb.AccountHistoryModel,
		accountIndexList,
		curHeight,
		treeCtx,
		assetCacheSize,
	)

	if err != nil {
		logx.Error("dbinitializer account tree failed:", err)
		return nil, err
	}
	nftTree, err := tree.InitNftTree(
		chainDb.L2NftHistoryModel,
		curHeight,
		treeCtx,
	)
	if err != nil {
		logx.Error("dbinitializer nft tree failed:", err)
		return nil, err
	}

	cacheConfig.sanitize()
	accountCache, err := lru.New(cacheConfig.AccountCacheSize)
	if err != nil {
		logx.Error("init account cache failed:", err)
		return nil, err
	}
	nftCache, err := lru.New(cacheConfig.NftCacheSize)
	if err != nil {
		logx.Error("init nft cache failed:", err)
		return nil, err
	}

	memCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(cacheConfig.MemCacheSize) * 10,
		MaxCost:     int64(cacheConfig.MemCacheSize),
		BufferItems: 64, // official recommended value

		// Called when setting cost to 0 in `Set/SetWithTTL`
		Cost: func(value interface{}) int64 {
			return 1
		},
	})
	if err != nil {
		logx.Error("MemCache init failed:", err)
		return nil, err
	}

	return &StateDB{
		StateCache:   NewStateCache(stateRoot),
		chainDb:      chainDb,
		redisCache:   redisCache,
		AccountCache: accountCache,
		NftCache:     nftCache,
		MemCache:     memCache,

		AccountTree:       accountTree,
		NftTree:           nftTree,
		AccountAssetTrees: accountAssetTrees,
		TreeCtx:           treeCtx,
	}, nil
}

func NewStateDBForDryRun(redisCache dbcache.Cache, cacheConfig *CacheConfig, chainDb *ChainDB, memCache *ristretto.Cache) (*StateDB, error) {
	accountCache, err := lru.New(cacheConfig.AccountCacheSize)
	if err != nil {
		logx.Error("init account cache failed:", err)
		return nil, err
	}
	nftCache, err := lru.New(cacheConfig.NftCacheSize)
	if err != nil {
		logx.Error("init nft cache failed:", err)
		return nil, err
	}

	return &StateDB{
		DryRun:       true,
		redisCache:   redisCache,
		chainDb:      chainDb,
		AccountCache: accountCache,
		MemCache:     memCache,
		NftCache:     nftCache,
		StateCache:   NewStateCache(""),
	}, nil
}

func (s *StateDB) GetFormatAccount(accountIndex int64) (*types.AccountInfo, error) {
	var start time.Time
	start = time.Now()
	if metrics.GetAccountCounter != nil {
		metrics.GetAccountCounter.Inc()
	}
	pending, exist := s.StateCache.GetPendingAccount(accountIndex)
	if exist {
		return pending, nil
	}

	cached, exist := s.AccountCache.Get(accountIndex)
	if exist {
		return cached.(*types.AccountInfo), nil
	}

	startGauge := time.Now()
	account, err := s.chainDb.AccountModel.GetAccountByIndex(accountIndex)
	if metrics.AccountFromDbGauge != nil {
		metrics.AccountFromDbGauge.Set(float64(time.Since(startGauge).Milliseconds()))
	}
	if metrics.GetAccountFromDbCounter != nil {
		metrics.GetAccountFromDbCounter.Inc()
	}

	if err == types.DbErrNotFound {
		return nil, types.AppErrAccountNotFound
	} else if err != nil {
		return nil, err
	}
	formatAccount, err := chain.ToFormatAccountInfo(account)
	if err != nil {
		return nil, err
	}
	s.AccountCache.Add(accountIndex, formatAccount)
	if metrics.AccountGauge != nil {
		metrics.AccountGauge.Set(float64(time.Since(start).Milliseconds()))
	}

	return formatAccount, nil
}

func (s *StateDB) GetAccount(accountIndex int64) (*account.Account, error) {
	pending, exist := s.StateCache.GetPendingAccount(accountIndex)
	if exist {
		account, err := chain.FromFormatAccountInfo(pending)
		if err != nil {
			return nil, err
		}
		return account, nil
	}

	cached, exist := s.AccountCache.Get(accountIndex)
	if exist {
		// to save account to cache, we need to convert it
		account, err := chain.FromFormatAccountInfo(cached.(*types.AccountInfo))
		if err == nil {
			return account, nil
		}
	}

	account, err := s.chainDb.AccountModel.GetAccountByIndex(accountIndex)
	if err != nil {
		return nil, err
	}
	formatAccount, err := chain.ToFormatAccountInfo(account)
	if err != nil {
		return nil, err
	}
	s.AccountCache.Add(accountIndex, formatAccount)
	return account, nil
}

// GetAccountByName get the account by its name.
// Firstly, try to find the account in the current state cache, it iterates the pending
// account map, not performance friendly, please take care when use this API.
// Secondly, if not found in the current state cache, then try to find the account from database.
func (s *StateDB) GetAccountByName(accountName string) (*account.Account, error) {
	for _, accountInfo := range s.PendingAccountMap {
		if accountInfo.AccountName == accountName {
			account, err := chain.FromFormatAccountInfo(accountInfo)
			if err != nil {
				return nil, err
			}

			return account, nil
		}
	}

	account, err := s.chainDb.AccountModel.GetAccountByName(accountName)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountByNameHash get the account by its name hash.
// Firstly, try to find the account in the current state cache, it iterates the pending
// account map, not performance friendly, please take care when use this API.
// Secondly, if not found in the current state cache, then try to find the account from database.
func (s *StateDB) GetAccountByNameHash(accountNameHash string) (*account.Account, error) {
	for _, accountInfo := range s.PendingAccountMap {
		if accountInfo.AccountNameHash == accountNameHash {
			account, err := chain.FromFormatAccountInfo(accountInfo)
			if err != nil {
				return nil, err
			}

			return account, nil
		}
	}

	account, err := s.chainDb.AccountModel.GetAccountByNameHash(accountNameHash)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *StateDB) GetNft(nftIndex int64) (*nft.L2Nft, error) {
	pending, exist := s.StateCache.GetPendingNft(nftIndex)
	if exist {
		return pending, nil
	}
	cached, exist := s.NftCache.Get(nftIndex)
	if exist {
		return cached.(*nft.L2Nft), nil
	}
	nft, err := s.chainDb.L2NftModel.GetNft(nftIndex)
	if err == types.DbErrNotFound {
		return nil, types.AppErrNftNotFound
	} else if err != nil {
		return nil, err
	}
	s.NftCache.Add(nftIndex, nft)
	return nft, nil
}

// MarkGasAccountAsPending will mark gas account as pending account. Putting gas account is pending
// account will unify many codes and remove some tricky logics.
func (s *StateDB) MarkGasAccountAsPending() error {
	gasAccount, err := s.GetFormatAccount(types.GasAccount)
	if err != nil && err != types.AppErrAccountNotFound {
		return err
	}
	if err == nil {
		s.PendingAccountMap[types.GasAccount] = gasAccount
	}
	return nil
}

func (s *StateDB) SyncPendingAccountToRedis(pendingAccount map[int64]*types.AccountInfo) {
	for index, formatAccount := range pendingAccount {
		account, err := chain.FromFormatAccountInfo(formatAccount)
		if err != nil {
			logx.Errorf("format accountInfo error, err=%v,formatAccount=%v", err, formatAccount)
			continue
		}
		err = s.redisCache.Set(context.Background(), dbcache.AccountKeyByIndex(index), account)
		if err != nil {
			logx.Errorf("cache to redis failed: %v,formatAccount=%v", err, formatAccount)
		}
		if formatAccount.AccountId == 0 {
			err = s.redisCache.Set(context.Background(), dbcache.AccountKeyByPK(formatAccount.PublicKey), account)
			if err != nil {
				logx.Errorf("cache to redis failed: %v,formatAccount=%v", err, formatAccount)
			}
			err = s.redisCache.Set(context.Background(), dbcache.AccountKeyByName(formatAccount.AccountName), account)
			if err != nil {
				logx.Errorf("cache to redis failed: %v,formatAccount=%v", err, formatAccount)
			}
		}
	}
}

func (s *StateDB) SyncPendingNftToRedis(pendingNft map[int64]*nft.L2Nft) {
	for index, nft := range pendingNft {
		err := s.redisCache.Set(context.Background(), dbcache.NftKeyByIndex(index), nft)
		if err != nil {
			logx.Errorf("cache to redis failed: %v,nft=%v", err, nft)
			continue
		}
	}
}

func (s *StateDB) SyncPendingAccountToMemoryCache(pendingAccount map[int64]*types.AccountInfo) {
	for index, formatAccount := range pendingAccount {
		s.AccountCache.Add(index, formatAccount)
	}
}

func (s *StateDB) SyncPendingNftToMemoryCache(pendingNft map[int64]*nft.L2Nft) {
	for index, nft := range pendingNft {
		s.NftCache.Add(index, nft)
	}
}

func (s *StateDB) SyncGasAccountToRedis() error {
	if cacheAccount, ok := s.AccountCache.Get(types.GasAccount); ok {
		formatAccount := cacheAccount.(*types.AccountInfo)
		account, err := chain.FromFormatAccountInfo(formatAccount)
		if err != nil {
			return err
		}
		err = s.redisCache.Set(context.Background(), dbcache.AccountKeyByIndex(account.AccountIndex), account)
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
	}
	return nil
}

func (s *StateDB) PurgeCache(stateRoot string) {
	s.StateCache = NewStateCache(stateRoot)
}

func (s *StateDB) GetPendingAccount(blockHeight int64, stateDataCopy *StateDataCopy) ([]*account.Account, []*account.AccountHistory, error) {
	pendingAccount := make([]*account.Account, 0)
	pendingAccountHistory := make([]*account.AccountHistory, 0)

	for _, formatAccount := range stateDataCopy.StateCache.PendingAccountMap {
		newAccount, err := chain.FromFormatAccountInfo(formatAccount)
		if err != nil {
			return nil, nil, err
		}
		newAccount.L2BlockHeight = blockHeight
		pendingAccount = append(pendingAccount, newAccount)
		pendingAccountHistory = append(pendingAccountHistory, &account.AccountHistory{
			AccountIndex:    newAccount.AccountIndex,
			Nonce:           newAccount.Nonce,
			CollectionNonce: newAccount.CollectionNonce,
			AssetInfo:       newAccount.AssetInfo,
			AssetRoot:       newAccount.AssetRoot,
			L2BlockHeight:   blockHeight,
		})
	}

	return pendingAccount, pendingAccountHistory, nil
}

func (s *StateDB) GetPendingNft(blockHeight int64, stateDataCopy *StateDataCopy) ([]*nft.L2Nft, []*nft.L2NftHistory, error) {
	pendingNft := make([]*nft.L2Nft, 0)
	pendingNftHistory := make([]*nft.L2NftHistory, 0)

	for _, newNft := range stateDataCopy.StateCache.PendingNftMap {
		pendingNft = append(pendingNft, newNft)
		pendingNftHistory = append(pendingNftHistory, &nft.L2NftHistory{
			NftIndex:            newNft.NftIndex,
			CreatorAccountIndex: newNft.CreatorAccountIndex,
			OwnerAccountIndex:   newNft.OwnerAccountIndex,
			NftContentHash:      newNft.NftContentHash,
			CreatorTreasuryRate: newNft.CreatorTreasuryRate,
			CollectionId:        newNft.CollectionId,
			L2BlockHeight:       blockHeight,
		})
	}

	return pendingNft, pendingNftHistory, nil
}

func (s *StateDB) DeepCopyAccounts(accountIds []int64) (map[int64]*types.AccountInfo, error) {
	accounts := make(map[int64]*types.AccountInfo)
	if len(accountIds) == 0 {
		return accounts, nil
	}

	for _, accountId := range accountIds {
		if _, ok := accounts[accountId]; ok {
			continue
		}
		account, err := s.GetFormatAccount(accountId)
		if err != nil {
			return nil, err
		}
		accounts[accountId] = account.DeepCopy()
	}

	return accounts, nil
}

func (s *StateDB) PrepareAccountsAndAssets(accountAssetsMap map[int64]map[int64]bool) error {
	for accountIndex, assets := range accountAssetsMap {
		if s.DryRun {
			account := &account.Account{}
			redisAccount, err := s.redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex), account)
			if err == nil && redisAccount != nil {
				formatAccount, err := chain.ToFormatAccountInfo(account)
				if err == nil {
					s.AccountCache.Add(accountIndex, formatAccount)
				}
			}
		}

		account, err := s.GetFormatAccount(accountIndex)
		if err != nil {
			return err
		}
		if account.AssetInfo == nil {
			account.AssetInfo = make(map[int64]*types.AccountAsset)
		}
		for assetId := range assets {
			if account.AssetInfo[assetId] == nil {
				account.AssetInfo[assetId] = &types.AccountAsset{
					AssetId:                  assetId,
					Balance:                  types.ZeroBigInt,
					OfferCanceledOrFinalized: types.ZeroBigInt,
				}
			}
		}
		s.AccountCache.Add(accountIndex, account)
	}

	return nil
}

func (s *StateDB) PrepareNft(nftIndex int64) (*nft.L2Nft, error) {
	if s.DryRun {
		n := &nft.L2Nft{}
		redisNft, err := s.redisCache.Get(context.Background(), dbcache.NftKeyByIndex(nftIndex), n)
		if err == nil && redisNft != nil {
			s.NftCache.Add(nftIndex, n)
		}
	}

	return s.GetNft(nftIndex)
}

const (
	accountTreeRole = "account"
	nftTreeRole     = "nft"
)

type treeUpdateResp struct {
	role  string
	index int64
	leaf  []byte
	err   error
}

func (s *StateDB) UpdateAssetTree(cleanDirty bool, stateDataCopy *StateDataCopy) error {
	taskNum := 0
	resultChan := make(chan *treeUpdateResp, 1)
	defer close(resultChan)

	for accountIndex, assetsMap := range stateDataCopy.StateCache.dirtyAccountsAndAssetsMap {
		assets := make([]int64, 0, len(assetsMap))
		for assetIndex, isDirty := range assetsMap {
			if !isDirty {
				continue
			}
			assets = append(assets, assetIndex)
		}
		taskNum++
		err := func(accountIndex int64, assets []int64) error {
			return gopool.Submit(func() {
				index, leaf, err := s.SetAndCommitAssetTree(accountIndex, assets, stateDataCopy)
				resultChan <- &treeUpdateResp{
					role:  accountTreeRole,
					index: index,
					leaf:  leaf,
					err:   err,
				}
			})
		}(accountIndex, assets)
		if err != nil {
			return err
		}
	}

	for nftIndex, isDirty := range stateDataCopy.StateCache.dirtyNftMap {
		if !isDirty {
			continue
		}
		taskNum++
		err := func(nftIndex int64) error {
			return gopool.Submit(func() {
				index, leaf, err := s.computeNftLeafHash(nftIndex, stateDataCopy)
				resultChan <- &treeUpdateResp{
					role:  nftTreeRole,
					index: index,
					leaf:  leaf,
					err:   err,
				}
			})
		}(nftIndex)
		if err != nil {
			return err
		}
	}

	if cleanDirty {
		stateDataCopy.StateCache.dirtyAccountsAndAssetsMap = make(map[int64]map[int64]bool, 0)
		stateDataCopy.StateCache.dirtyNftMap = make(map[int64]bool, 0)
	}

	pendingAccountItem := make([]bsmt.Item, 0, len(stateDataCopy.StateCache.dirtyAccountsAndAssetsMap))
	pendingNftItem := make([]bsmt.Item, 0, len(stateDataCopy.StateCache.dirtyNftMap))
	for i := 0; i < taskNum; i++ {
		result := <-resultChan
		if result.err != nil {
			return result.err
		}

		switch result.role {
		case accountTreeRole:
			pendingAccountItem = append(pendingAccountItem, bsmt.Item{Key: uint64(result.index), Val: result.leaf})
		case nftTreeRole:
			pendingNftItem = append(pendingNftItem, bsmt.Item{Key: uint64(result.index), Val: result.leaf})
		}
	}
	stateDataCopy.pendingAccountSmtItem = pendingAccountItem
	stateDataCopy.pendingNftSmtItem = pendingNftItem
	return nil
}

func (s *StateDB) SetAccountAndNftTree(stateDataCopy *StateDataCopy) error {
	start := time.Now()
	resultChan := make(chan *treeUpdateResp, 1)
	defer close(resultChan)
	err := gopool.Submit(func() {
		resultChan <- &treeUpdateResp{
			role: accountTreeRole,
			err:  s.AccountTree.MultiSetWithVersion(stateDataCopy.pendingAccountSmtItem, bsmt.Version(stateDataCopy.CurrentBlock.BlockHeight)),
		}
	})
	if err != nil {
		return err
	}
	err = gopool.Submit(func() {
		resultChan <- &treeUpdateResp{
			role: nftTreeRole,
			err:  s.NftTree.MultiSetWithVersion(stateDataCopy.pendingNftSmtItem, bsmt.Version(stateDataCopy.CurrentBlock.BlockHeight)),
		}
	})
	if err != nil {
		return err
	}

	for i := 0; i < 2; i++ {
		result := <-resultChan
		if result.err != nil {
			return fmt.Errorf("update %s tree failed, %v", result.role, result.err)
		}
	}
	metrics.AccountTreeMultiSetGauge.Set(float64(time.Since(start).Milliseconds()))

	//hFunc := poseidon.NewPoseidon()
	hFunc := tree.NewGMimc()
	hFunc.Write(s.AccountTree.Root())
	hFunc.Write(s.NftTree.Root())
	stateDataCopy.StateCache.StateRoot = common.Bytes2Hex(hFunc.Sum(nil))
	return nil
}

func (s *StateDB) SetAndCommitAssetTree(accountIndex int64, assets []int64, stateCopy *StateDataCopy) (int64, []byte, error) {
	start := time.Now()
	account, exist := stateCopy.StateCache.GetPendingAccount(accountIndex)
	metrics.AccountTreeTimeGauge.WithLabelValues("cache_get_account").Set(float64(time.Since(start).Milliseconds()))
	if !exist {
		logx.Infof("update account tree failed,not exist accountIndex=%s", accountIndex)
	}
	start = time.Now()
	pendingUpdateAssetItem := make([]bsmt.Item, 0, len(assets))
	metrics.AccountTreeTimeGauge.WithLabelValues("assets_count").Set(float64(len(assets)))
	for _, assetId := range assets {
		balance := account.AssetInfo[assetId].Balance
		startItem := time.Now()
		assetLeaf, err := tree.ComputeAccountAssetLeafHash(
			balance.String(),
			account.AssetInfo[assetId].OfferCanceledOrFinalized.String(), accountIndex, assetId, stateCopy.CurrentBlock.BlockHeight,
		)
		metrics.AccountTreeTimeGauge.WithLabelValues("compute_poseidon").Set(float64(time.Since(startItem).Milliseconds()))
		if err != nil {
			return accountIndex, nil, fmt.Errorf("compute new account asset leaf failed: %v", err)
		}
		pendingUpdateAssetItem = append(pendingUpdateAssetItem, bsmt.Item{Key: uint64(assetId), Val: assetLeaf})
	}
	metrics.AccountTreeTimeGauge.WithLabelValues("for_assets").Set(float64(time.Since(start).Milliseconds()))

	start = time.Now()
	err := s.AccountAssetTrees.Get(accountIndex).MultiSetWithVersion(pendingUpdateAssetItem, bsmt.Version(stateCopy.CurrentBlock.BlockHeight))
	if err != nil {
		return accountIndex, nil, fmt.Errorf("update asset tree failed: %v", err)
	}
	metrics.AccountTreeTimeGauge.WithLabelValues("multiSet").Set(float64(time.Since(start).Milliseconds()))

	account.AssetRoot = common.Bytes2Hex(s.AccountAssetTrees.Get(accountIndex).Root())
	nAccountLeafHash, err := tree.ComputeAccountLeafHash(
		account.AccountNameHash,
		account.PublicKey,
		account.Nonce,
		account.CollectionNonce,
		s.AccountAssetTrees.Get(accountIndex).Root(),
		accountIndex,
		stateCopy.CurrentBlock.BlockHeight,
	)
	if err != nil {
		return accountIndex, nil, fmt.Errorf("unable to compute account leaf: %v", err)
	}
	asset := s.AccountAssetTrees.Get(accountIndex)
	prunedVersion := bsmt.Version(tree.GetAssetLatestVerifiedHeight(s.GetPrunedBlockHeight(), asset.Versions()))
	latestVersion := asset.LatestVersion()
	if prunedVersion > latestVersion {
		prunedVersion = latestVersion
	}
	newVersion := bsmt.Version(stateCopy.CurrentBlock.BlockHeight)
	logx.Infof("asset.CommitWithNewVersion:blockHeight=%s,accountIndex=%s,prunedVersion=%s:", stateCopy.CurrentBlock.BlockHeight, accountIndex, prunedVersion)
	ver, err := asset.CommitWithNewVersion(&prunedVersion, &newVersion)
	if err != nil {
		logx.Error("asset.Commit failed:", err)
		return accountIndex, nil, fmt.Errorf("unable to commit asset tree [%d], tree ver: %d, prune ver: %d,error:%s", accountIndex, ver, prunedVersion, err.Error())
	}
	assetOne, err := asset.Get(0, nil)
	if err == nil {
		logx.Infof("asset.CommitWithNewVersion:blockHeight=%s,accountIndex=%s,assetId=0,hash=%s:", stateCopy.CurrentBlock.BlockHeight, accountIndex, common.Bytes2Hex(assetOne))
	}
	return accountIndex, nAccountLeafHash, nil
}

func (s *StateDB) computeNftLeafHash(nftIndex int64, stateCopy *StateDataCopy) (int64, []byte, error) {
	start := time.Now()
	nftInfo, exist := stateCopy.StateCache.GetPendingNft(nftIndex)
	if !exist {
		logx.Error("computeNftLeafHash failed,No NFT found in GetPendingNft nftIndex=%s", nftIndex)
		return nftIndex, nil, fmt.Errorf("computeNftLeafHash failed,No NFT found in GetPendingNft nftIndex=%v", nftIndex)
	}
	nftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
		nftInfo.CreatorAccountIndex,
		nftInfo.OwnerAccountIndex,
		nftInfo.NftContentHash,
		nftInfo.CreatorTreasuryRate,
		nftInfo.CollectionId,
		nftInfo.NftIndex,
		stateCopy.CurrentBlock.BlockHeight,
	)
	if err != nil {
		return nftIndex, nil, fmt.Errorf("unable to compute nftInfo leaf: %v", err)
	}
	metrics.NftTreeTimeGauge.WithLabelValues("nftInfo").Set(float64(time.Since(start).Milliseconds()))
	return nftIndex, nftAssetLeaf, nil
}

func (s *StateDB) GetCommittedNonce(accountIndex int64) (int64, error) {
	account, err := s.GetFormatAccount(accountIndex)
	if err != nil {
		return 0, err
	}
	return account.Nonce, nil
}

func (s *StateDB) GetPendingNonce(accountIndex int64) (int64, error) {
	nonce, err := s.chainDb.TxPoolModel.GetMaxNonceByAccountIndex(accountIndex)
	if err == nil {
		return nonce + 1, nil
	}
	account := &account.Account{}
	redisAccount, err := s.redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex), account)
	if err == nil && redisAccount != nil {
		return account.Nonce, nil
	}
	dbAccount, err := s.chainDb.AccountModel.GetAccountByIndex(accountIndex)
	if err == nil {
		return dbAccount.Nonce, nil
	}
	return 0, err
}

func (s *StateDB) GetPendingNonceFromCache(accountIndex int64) (int64, error) {
	accountNonce := int64(-2)
	_, err := s.redisCache.Get(context.Background(), dbcache.AccountNonceKeyByIndex(accountIndex), &accountNonce)
	if err == nil && accountNonce != -2 {
		return accountNonce + 1, nil
	}
	pendingNonce, err := s.GetPendingNonce(accountIndex)
	if err == nil {
		_ = s.redisCache.Set(context.Background(), dbcache.AccountNonceKeyByIndex(accountIndex), pendingNonce-1)
		return pendingNonce, err
	}
	return 0, err
}

func (s *StateDB) ClearPendingNonceFromRedisCache(accountIndex int64) {
	_ = s.redisCache.Delete(context.Background(), dbcache.AccountNonceKeyByIndex(accountIndex))
}

func (s *StateDB) SetPendingNonceToRedisCache(accountIndex int64, nonce int64) {
	_ = s.redisCache.Set(context.Background(), dbcache.AccountNonceKeyByIndex(accountIndex), nonce)
}

func (s *StateDB) GetNextAccountIndex() int64 {
	return s.AccountAssetTrees.GetNextAccountIndex()
}

func (s *StateDB) GetCurrentAccountIndex() int64 {
	return s.AccountAssetTrees.GetCurrentAccountIndex()
}

func (c *StateDB) UpdateNftIndex(nftIndex int64) {
	c.nextNftIndexLock.Lock()
	if c.nextNftIndex < nftIndex {
		c.nextNftIndex = nftIndex
	}
	c.nextNftIndexLock.Unlock()
}

func (c *StateDB) GetNextNftIndex() int64 {
	c.nextNftIndexLock.RLock()
	defer c.nextNftIndexLock.RUnlock()
	return c.nextNftIndex + 1
}

func (c *StateDB) GetCurrentNftIndex() int64 {
	c.nextNftIndexLock.RLock()
	defer c.nextNftIndexLock.RUnlock()
	return c.nextNftIndex
}

func (s *StateDB) GetGasAccountIndex() (int64, error) {
	result, found := s.MemCache.Get(dbcache.GasAccountKey)
	if found {
		return result.(int64), nil
	}
	logx.Infof("GetGasAccountIndex mem cache expired")
	gasAccountConfig, err := s.chainDb.SysConfigModel.GetSysConfigByName(types.GasAccountIndex)
	if err != nil {
		logx.Errorf("cannot find config for: %s", types.GasAccountIndex)
		return -1, errors.New("internal error")
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("invalid account index: %s", gasAccountConfig.Value)
		return -1, errors.New("internal error")
	}
	s.MemCache.SetWithTTL(dbcache.GasAccountKey, gasAccountIndex, 0, time.Duration(24)*time.Hour)
	return gasAccountIndex, nil
}

func (s *StateDB) GetGasConfig() (map[uint32]map[int]int64, error) {
	gasFeeValue := ""
	result, found := s.MemCache.Get(dbcache.GasConfigKey)
	if found {
		gasFeeValue = result.(string)
	} else {
		logx.Infof("fail to get gas config from cache")
		cfgGasFee, err := s.chainDb.SysConfigModel.GetSysConfigByName(types.SysGasFee)
		if err != nil {
			logx.Errorf("cannot find gas asset: %s", err.Error())
			return nil, errors.New("invalid gas fee asset")
		}
		gasFeeValue = cfgGasFee.Value
		s.MemCache.SetWithTTL(dbcache.GasConfigKey, gasFeeValue, 0, time.Duration(15)*time.Minute)
	}
	m := make(map[uint32]map[int]int64)
	err := json.Unmarshal([]byte(gasFeeValue), &m)
	if err != nil {
		logx.Errorf("fail to unmarshal gas fee config, err: %s", err.Error())
		return nil, errors.New("internal error")
	}
	return m, nil
}

func (c *StateDB) UpdatePrunedBlockHeight(latestBlock int64) {
	c.prunedBlockHeightLock.Lock()
	if c.prunedBlockHeight < latestBlock {
		c.prunedBlockHeight = latestBlock
	}
	c.prunedBlockHeightLock.Unlock()
}

func (c *StateDB) GetPrunedBlockHeight() int64 {
	c.prunedBlockHeightLock.RLock()
	defer c.prunedBlockHeightLock.RUnlock()
	return c.prunedBlockHeight
}

func (c *StateDB) UpdateNeedRestoreExecutedTxs(need bool) {
	c.needRestoreExecutedTxsLock.Lock()
	c.needRestoreExecutedTxs = need
	c.needRestoreExecutedTxsLock.Unlock()
}

func (c *StateDB) NeedRestoreExecutedTxs() bool {
	c.needRestoreExecutedTxsLock.RLock()
	defer c.needRestoreExecutedTxsLock.RUnlock()
	return c.needRestoreExecutedTxs
}

func (c *StateDB) UpdateMaxPoolTxIdFinished(maxPoolTxId uint) {
	c.maxPoolTxIdFinishedLock.Lock()
	if maxPoolTxId > c.maxPoolTxIdFinished {
		c.maxPoolTxIdFinished = maxPoolTxId
	}
	c.maxPoolTxIdFinishedLock.Unlock()
}

func (c *StateDB) GetMaxPoolTxIdFinished() uint {
	c.maxPoolTxIdFinishedLock.RLock()
	defer c.maxPoolTxIdFinishedLock.RUnlock()
	return c.maxPoolTxIdFinished
}

func (s *StateDB) Close() {
	sqlDB, err := s.chainDb.DB.DB()
	if err == nil && sqlDB != nil {
		err = sqlDB.Close()
	}
	if err != nil {
		logx.Errorf("close db error: %s", err.Error())
	}

	err = s.redisCache.Close()
	if err != nil {
		logx.Errorf("close redis error: %s", err.Error())
	}

	err = s.TreeCtx.TreeDB.Close()
	if err != nil {
		logx.Errorf("close treedb error: %s", err.Error())
	}
}
