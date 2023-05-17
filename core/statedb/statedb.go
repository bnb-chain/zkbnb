package statedb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb/common/log"
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
	IsFromApi bool
	// State cache
	*StateCache
	chainDb    *ChainDB
	redisCache dbcache.Cache

	// Flat state
	AccountCache   *lru.Cache
	L1AddressCache *lru.Cache
	NftCache       *lru.Cache
	MemCache       *ristretto.Cache

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
		true,
	)

	if err != nil {
		logx.Error("dbinitializer account tree failed:", err)
		return nil, err
	}
	nftTree, err := tree.InitNftTree(
		chainDb.L2NftModel,
		chainDb.L2NftHistoryModel,
		curHeight,
		treeCtx,
		true,
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
	l1AddressCache, err := lru.New(cacheConfig.AccountCacheSize)
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
		StateCache:     NewStateCache(stateRoot),
		chainDb:        chainDb,
		redisCache:     redisCache,
		AccountCache:   accountCache,
		L1AddressCache: l1AddressCache,
		NftCache:       nftCache,
		MemCache:       memCache,

		AccountTree:       accountTree,
		NftTree:           nftTree,
		AccountAssetTrees: accountAssetTrees,
		TreeCtx:           treeCtx,
	}, nil
}

func NewStateDBForDesertExit(redisCache dbcache.Cache, cacheConfig *CacheConfig, chainDb *ChainDB) (*StateDB, error) {
	accountCache, err := lru.New(cacheConfig.AccountCacheSize)
	if err != nil {
		logx.Error("init account cache failed:", err)
		return nil, err
	}
	l1AddressCache, err := lru.New(cacheConfig.AccountCacheSize)
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
		IsFromApi:      false,
		redisCache:     redisCache,
		chainDb:        chainDb,
		AccountCache:   accountCache,
		L1AddressCache: l1AddressCache,
		MemCache:       memCache,
		NftCache:       nftCache,
		StateCache:     NewStateCache(""),
	}, nil
}

func NewStateDBForDryRun(redisCache dbcache.Cache, cacheConfig *CacheConfig, chainDb *ChainDB, memCache *ristretto.Cache) (*StateDB, error) {
	accountCache, err := lru.New(cacheConfig.AccountCacheSize)
	if err != nil {
		logx.Error("init account cache failed:", err)
		return nil, err
	}
	l1AddressCache, err := lru.New(cacheConfig.AccountCacheSize)
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
		IsFromApi:      true,
		redisCache:     redisCache,
		chainDb:        chainDb,
		AccountCache:   accountCache,
		L1AddressCache: l1AddressCache,
		MemCache:       memCache,
		NftCache:       nftCache,
		StateCache:     NewStateCache(""),
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
		accountInfo := cached.(*types.AccountInfo)
		if accountInfo.AccountIndex != accountIndex {
			return nil, types.AppErrInvalidAccount
		}
		return accountInfo, nil
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
	s.L1AddressCache.Add(formatAccount.L1Address, accountIndex)

	if metrics.AccountGauge != nil {
		metrics.AccountGauge.Set(float64(time.Since(start).Milliseconds()))
	}

	return formatAccount, nil
}

func (s *StateDB) isAccountExistInCache(accountIndex int64) bool {
	_, exist := s.StateCache.GetPendingAccount(accountIndex)
	if exist {
		return true
	}

	_, exist = s.AccountCache.Get(accountIndex)
	if exist {
		return true
	}

	return false
}

// GetAccountByL1Address get the account by l1 address.
// Firstly, try to find the account in the current state cache, it iterates the pending
// account map, not performance friendly, please take care when use this API.
// Secondly, if not found in the current state cache, then try to find the account from database.
func (s *StateDB) GetAccountByL1Address(l1Address string) (*types.AccountInfo, error) {
	if s.IsFromApi {
		var accountIndex interface{}
		var redisAccount interface{}
		redisAccount, err := s.redisCache.Get(context.Background(), dbcache.AccountKeyByL1Address(l1Address), &accountIndex)
		if err == nil && redisAccount != nil {
			account := &account.Account{}
			redisAccount, err := s.redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex.(int64)), account)
			if err == nil && redisAccount != nil {
				formatAccount, err := chain.ToFormatAccountInfo(account)
				if err == nil {
					s.AccountCache.Add(accountIndex, formatAccount)
					s.L1AddressCache.Add(formatAccount.L1Address, accountIndex)
				}
			}
		}
	}
	var exist bool
	var accountIndex int64
	accountIndex, exist = s.StateCache.GetPendingAccountL1AddressMap(l1Address)
	if !exist {
		var accountIndexInterface interface{}
		accountIndexInterface, exist = s.L1AddressCache.Get(l1Address)
		if exist {
			accountIndex = accountIndexInterface.(int64)
		}
	}

	if exist {
		fromAccount, err := s.GetFormatAccount(accountIndex)
		if err != nil {
			return nil, err
		}
		if fromAccount.AccountIndex == accountIndex && fromAccount.L1Address == l1Address {
			return fromAccount, err
		} else {
			return nil, types.AppErrInvalidAccount
		}
	}

	accountInfo, err := s.chainDb.AccountModel.GetAccountByL1Address(l1Address)
	if err == types.DbErrNotFound {
		return nil, types.AppErrAccountNotFound
	} else if err != nil {
		return nil, err
	}
	formatAccount, err := chain.ToFormatAccountInfo(accountInfo)
	if err != nil {
		return nil, err
	}
	s.AccountCache.Add(accountInfo.AccountIndex, formatAccount)
	s.L1AddressCache.Add(formatAccount.L1Address, accountInfo.AccountIndex)
	return formatAccount, nil
}

func (s *StateDB) isAddressExistInCache(l1Address string) bool {
	_, exist := s.StateCache.GetPendingAccountL1AddressMap(l1Address)
	if exist {
		return true
	}

	_, exist = s.L1AddressCache.Get(l1Address)
	if exist {
		return true
	}

	return false
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

func (s *StateDB) isNftExistInCache(nftIndex int64) bool {
	_, exist := s.StateCache.GetPendingNft(nftIndex)
	if exist {
		return true
	}
	_, exist = s.NftCache.Get(nftIndex)
	if exist {
		return true
	}
	return false
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
			logx.Errorf("format accountInfo error, err=%v,formatAccount=%v", err, formatAccount.AccountIndex)
			continue
		}
		err = s.redisCache.Set(context.Background(), dbcache.AccountKeyByIndex(index), account)
		if err != nil {
			logx.Errorf("cache to redis failed: %v,formatAccount=%v", err, formatAccount.AccountIndex)
		}
		if formatAccount.AccountId == 0 {
			err = s.redisCache.Set(context.Background(), dbcache.AccountKeyByL1Address(formatAccount.L1Address), account.AccountIndex)
			if err != nil {
				logx.Errorf("cache to redis failed: %v,formatAccount=%v", err, formatAccount.AccountIndex)
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
		s.L1AddressCache.Add(formatAccount.L1Address, index)
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
			Status:          newAccount.Status,
			L1Address:       newAccount.L1Address,
			PublicKey:       newAccount.PublicKey,
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
			RoyaltyRate:         newNft.RoyaltyRate,
			CollectionId:        newNft.CollectionId,
			L2BlockHeight:       blockHeight,
			NftContentType:      newNft.NftContentType,
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

func (s *StateDB) PrepareAccountsAndAssets(accountAssetsMap map[int64]map[int64]bool, creatingAccountIndex int64) error {
	for accountIndex, assets := range accountAssetsMap {
		if creatingAccountIndex == accountIndex {
			continue
		}
		if s.IsFromApi {
			account := &account.Account{}
			redisAccount, err := s.redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex), account)
			if err == nil && redisAccount != nil {
				formatAccount, err := chain.ToFormatAccountInfo(account)
				if err == nil {
					s.AccountCache.Add(accountIndex, formatAccount)
					s.L1AddressCache.Add(formatAccount.L1Address, accountIndex)
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
		s.L1AddressCache.Add(account.L1Address, accountIndex)
	}

	return nil
}

func (s *StateDB) PrepareNft(nftIndex int64) (*nft.L2Nft, error) {
	if s.IsFromApi {
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

// UpdateAssetTree compute account asset hash, commit asset smt,compute account leaf hash, compute nft leaf hash
func (s *StateDB) UpdateAssetTree(stateDataCopy *StateDataCopy) error {
	taskNum := 0
	resultChan := make(chan *treeUpdateResp, len(stateDataCopy.StateCache.dirtyAccountsAndAssetsMap)+len(stateDataCopy.StateCache.dirtyNftMap))
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
				ctx := log.NewCtxWithKV(log.BlockHeightContext, stateDataCopy.CurrentBlock.BlockHeight, log.AccountIndexCtx, accountIndex)
				index, leaf, err := s.SetAndCommitAssetTree(accountIndex, assets, stateDataCopy, ctx)
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

// SetAccountAndNftTree multi set account tree with version,multi set nft tree with version
func (s *StateDB) SetAccountAndNftTree(stateDataCopy *StateDataCopy) error {
	start := time.Now()
	resultChan := make(chan *treeUpdateResp, 2)
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
	accountTreeRoot := s.AccountTree.Root()
	nftTreeRoot := s.NftTree.Root()
	hFunc := tree.NewGMimc()
	hFunc.Write(accountTreeRoot)
	hFunc.Write(nftTreeRoot)
	logx.Infof("committer smt blockHeight=%d, account tree root=%s,nft tree root=%s", stateDataCopy.CurrentBlock.BlockHeight, common.Bytes2Hex(accountTreeRoot), common.Bytes2Hex(nftTreeRoot))
	stateDataCopy.StateCache.StateRoot = common.Bytes2Hex(hFunc.Sum(nil))
	return nil
}

// SetAndCommitAssetTree compute account asset hash, commit asset smt,compute account leaf hash
func (s *StateDB) SetAndCommitAssetTree(accountIndex int64, assets []int64, stateCopy *StateDataCopy, ctx context.Context) (int64, []byte, error) {
	start := time.Now()
	account, exist := stateCopy.StateCache.GetPendingAccount(accountIndex)
	metrics.AccountTreeTimeGauge.WithLabelValues("cache_get_account").Set(float64(time.Since(start).Milliseconds()))
	if !exist {
		return accountIndex, nil, fmt.Errorf("update account tree failed,not exist accountIndex=%d", accountIndex)
	}

	start = time.Now()
	pendingUpdateAssetItem := make([]bsmt.Item, 0, len(assets))
	metrics.AccountTreeTimeGauge.WithLabelValues("assets_count").Set(float64(len(assets)))
	for _, assetId := range assets {
		balance := account.AssetInfo[assetId].Balance
		startItem := time.Now()
		ctx := log.UpdateCtxWithKV(ctx, log.AssetIdCtx, assetId)
		assetLeaf, err := tree.ComputeAccountAssetLeafHash(balance.String(), account.AssetInfo[assetId].OfferCanceledOrFinalized.String(), ctx)
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
		account.L1Address,
		account.PublicKey,
		account.Nonce,
		account.CollectionNonce,
		s.AccountAssetTrees.Get(accountIndex).Root(),
		ctx,
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
	logx.WithContext(ctx).Infof("asset.CommitWithNewVersion:blockHeight=%d,accountIndex=%d,prunedVersion=%d:", stateCopy.CurrentBlock.BlockHeight, accountIndex, prunedVersion)
	ver, err := asset.CommitWithNewVersion(&prunedVersion, &newVersion)
	if err != nil {
		return accountIndex, nil, fmt.Errorf("unable to commit asset tree [%d], tree ver: %d, prune ver: %d,error:%s", accountIndex, ver, prunedVersion, err.Error())
	}

	return accountIndex, nAccountLeafHash, nil
}

// compute nft leaf hash
func (s *StateDB) computeNftLeafHash(nftIndex int64, stateCopy *StateDataCopy) (int64, []byte, error) {
	start := time.Now()
	nftInfo, exist := stateCopy.StateCache.GetPendingNft(nftIndex)
	if !exist {
		return nftIndex, nil, fmt.Errorf("computeNftLeafHash failed,No NFT found in GetPendingNft nftIndex=%d", nftIndex)
	}
	ctx := log.NewCtxWithKV(log.BlockHeightContext, stateCopy.CurrentBlock.BlockHeight, log.NftIndexCtx, nftIndex)
	nftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
		nftInfo.CreatorAccountIndex,
		nftInfo.OwnerAccountIndex,
		nftInfo.NftContentHash,
		nftInfo.RoyaltyRate,
		nftInfo.CollectionId,
		ctx,
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

func (s *StateDB) ClearProtocolRateFromRedisCache() {
	_ = s.redisCache.Delete(context.Background(), dbcache.ProtocolRate)
}

func (s *StateDB) GetProtocolRateFromRedisCache() (int64, error) {
	var protocolRate int64
	rate, err := s.redisCache.Get(context.Background(), dbcache.ProtocolRate, &protocolRate)
	if err == nil && rate != nil {
		return protocolRate, nil
	}
	sysProtocolRate, err := s.chainDb.SysConfigModel.GetSysConfigByName(types.ProtocolRate)
	if err == nil {
		feeRate, err := strconv.ParseInt(sysProtocolRate.Value, 10, 64)
		if err != nil {
			return 0, err
		}
		_ = s.redisCache.Set(context.Background(), dbcache.ProtocolRate, feeRate)
		return feeRate, err
	}
	return 0, err
}

func (s *StateDB) GetNextAccountIndex() int64 {
	return s.AccountAssetTrees.GetNextAccountIndex()
}

func (s *StateDB) GetCurrentAccountIndex() int64 {
	return s.AccountAssetTrees.GetCurrentAccountIndex()
}
func (s *StateDB) UpdateAccountIndex(accountIndex int64) {
	s.AccountAssetTrees.UpdateAccountIndex(accountIndex)
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
		return -1, types.AppErrNotFindGasAccountConfig
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("invalid account index: %s", gasAccountConfig.Value)
		return -1, types.AppErrInvalidGasAccountIndex
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
			return nil, types.AppErrInvalidGasFeeAsset
		}
		gasFeeValue = cfgGasFee.Value
		s.MemCache.SetWithTTL(dbcache.GasConfigKey, gasFeeValue, 0, time.Duration(15)*time.Minute)
	}
	m := make(map[uint32]map[int]int64)
	err := json.Unmarshal([]byte(gasFeeValue), &m)
	if err != nil {
		logx.Errorf("fail to unmarshal gas fee config, err: %s", err.Error())
		return nil, types.AppErrFailUnmarshalGasFeeConfig
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

func (c *StateDB) PreLoadAccountAndNft(accountIndexMap map[int64]bool, nftIndexMap map[int64]bool, addressMap map[string]bool) {
	var nftIndexList []int64
	for nftIndex, _ := range nftIndexMap {
		if c.isNftExistInCache(nftIndex) {
			continue
		}
		nftIndexList = append(nftIndexList, nftIndex)
	}
	if len(nftIndexList) > 0 {
		nftAssets, err := c.chainDb.L2NftModel.GetNftsByNftIndexes(nftIndexList)
		if err != nil {
			for _, nftAsset := range nftAssets {
				c.NftCache.Add(nftAsset.NftIndex, nftAsset)

				accountIndexMap[nftAsset.OwnerAccountIndex] = true
				accountIndexMap[nftAsset.CreatorAccountIndex] = true
			}
		}
	}

	var accountIndexList []int64
	for accountIndex, _ := range accountIndexMap {
		if c.isAccountExistInCache(accountIndex) {
			continue
		}
		accountIndexList = append(accountIndexList, accountIndex)
	}
	if len(accountIndexList) > 0 {
		accounts, err := c.chainDb.AccountModel.GetAccountByIndexes(accountIndexList)
		if err != nil {
			c.syncToMemCache(accounts)
		}
	}

	var addressList []string
	for address, _ := range addressMap {
		if c.isAddressExistInCache(address) {
			continue
		}
		addressList = append(addressList, address)
	}
	if len(addressList) > 0 {
		accounts, err := c.chainDb.AccountModel.GetAccountByL1Addresses(addressList)
		if err != nil {
			c.syncToMemCache(accounts)
		}
	}
}

func (c *StateDB) syncToMemCache(accounts []*account.Account) {
	for _, accountInfo := range accounts {
		formatAccount, err := chain.ToFormatAccountInfo(accountInfo)
		if err != nil {
			continue
		}
		c.AccountCache.Add(accountInfo.AccountIndex, formatAccount)
		c.L1AddressCache.Add(formatAccount.L1Address, accountInfo.AccountIndex)
	}
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
