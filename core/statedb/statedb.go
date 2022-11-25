package statedb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bnb-chain/zkbnb/common/zkbnbprometheus"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	lru "github.com/hashicorp/golang-lru"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
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
	}
)

type CacheConfig struct {
	AccountCacheSize int
	NftCacheSize     int
}

func (c *CacheConfig) sanitize() *CacheConfig {
	if c.AccountCacheSize <= 0 {
		c.AccountCacheSize = DefaultCacheConfig.AccountCacheSize
	}

	if c.NftCacheSize <= 0 {
		c.NftCacheSize = DefaultCacheConfig.NftCacheSize
	}

	return c
}

type StateDB struct {
	dryRun bool
	// State cache
	*StateCache
	chainDb    *ChainDB
	redisCache dbcache.Cache

	// Flat state
	AccountCache *lru.Cache
	NftCache     *lru.Cache

	// Tree state
	AccountTree       bsmt.SparseMerkleTree
	NftTree           bsmt.SparseMerkleTree
	AccountAssetTrees *tree.AssetTreeCache
	TreeCtx           *tree.Context
	mainLock          sync.RWMutex
	prunedBlockHeight int64
	PreviousStateRoot string
	Metrics           *zkbnbprometheus.StateDBMetrics
}

func NewStateDB(treeCtx *tree.Context, chainDb *ChainDB,
	redisCache dbcache.Cache, cacheConfig *CacheConfig, assetCacheSize int,
	stateRoot string, curHeight int64) (*StateDB, error) {
	err := tree.SetupTreeDB(treeCtx)
	if err != nil {
		logx.Error("setup tree db failed: ", err)
		return nil, err
	}
	accountTree, accountAssetTrees, err := tree.InitAccountTree(
		chainDb.AccountModel,
		chainDb.AccountHistoryModel,
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

	return &StateDB{
		StateCache:   NewStateCache(stateRoot),
		chainDb:      chainDb,
		redisCache:   redisCache,
		AccountCache: accountCache,
		NftCache:     nftCache,

		AccountTree:       accountTree,
		NftTree:           nftTree,
		AccountAssetTrees: accountAssetTrees,
		TreeCtx:           treeCtx,
	}, nil
}

func NewStateDBForDryRun(redisCache dbcache.Cache, cacheConfig *CacheConfig, chainDb *ChainDB) (*StateDB, error) {
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
		dryRun:       true,
		redisCache:   redisCache,
		chainDb:      chainDb,
		AccountCache: accountCache,
		NftCache:     nftCache,
		StateCache:   NewStateCache(""),
	}, nil
}

func (s *StateDB) GetFormatAccount(accountIndex int64) (*types.AccountInfo, error) {
	var start time.Time
	start = time.Now()
	if s.Metrics != nil && s.Metrics.GetAccountCounter != nil {
		s.Metrics.GetAccountCounter.Inc()
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
	if s.Metrics != nil && s.Metrics.GetAccountFromDbGauge != nil {
		s.Metrics.GetAccountFromDbGauge.Set(float64(time.Since(startGauge).Milliseconds()))
	}
	if s.Metrics != nil && s.Metrics.GetAccountFromDbCounter != nil {
		s.Metrics.GetAccountFromDbCounter.Inc()
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
	if s.Metrics != nil && s.Metrics.GetAccountGauge != nil {
		s.Metrics.GetAccountGauge.Set(float64(time.Since(start).Milliseconds()))
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
			continue
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

func (s *StateDB) SyncStateCacheToRedis() error {
	//todo
	// Sync pending to cache.
	//err := s.syncPendingAccount(s.PendingAccountMap)
	//if err != nil {
	//	return err
	//}
	//err = s.syncPendingNft(s.PendingNftMap)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (s *StateDB) PurgeCache(stateRoot string) {
	s.StateCache = NewStateCache(stateRoot)
}

func (s *StateDB) GetPendingAccount(blockHeight int64, stateDataCopy *StateDataCopy) ([]*account.Account, []*account.AccountHistory, error) {
	pendingAccount := make([]*account.Account, 0)
	pendingAccountHistory := make([]*account.AccountHistory, 0)

	for _, formatAccount := range stateDataCopy.StateCache.PendingAccountMap {
		if formatAccount.AccountIndex == types.GasAccount {
			s.applyGasUpdate(formatAccount, stateDataCopy)
		}

		newAccount, err := chain.FromFormatAccountInfo(formatAccount)
		if err != nil {
			return nil, nil, err
		}

		pendingAccount = append(pendingAccount, newAccount)
		pendingAccountHistory = append(pendingAccountHistory, &account.AccountHistory{
			AccountIndex:    newAccount.AccountIndex,
			Nonce:           newAccount.Nonce,
			CollectionNonce: newAccount.CollectionNonce,
			AssetInfo:       newAccount.AssetInfo,
			AssetRoot:       newAccount.AssetRoot,
			L2BlockHeight:   blockHeight, // TODO: ensure this should be the new block's height.
		})
	}

	return pendingAccount, pendingAccountHistory, nil
}

func (s *StateDB) applyGasUpdate(formatAccount *types.AccountInfo, stateDataCopy *StateDataCopy) {
	for assetId, delta := range stateDataCopy.StateCache.PendingGasMap {
		if asset, ok := formatAccount.AssetInfo[assetId]; ok {
			formatAccount.AssetInfo[assetId].Balance = ffmath.Add(asset.Balance, delta)
		} else {
			formatAccount.AssetInfo[assetId] = &types.AccountAsset{
				Balance:                  delta,
				OfferCanceledOrFinalized: types.ZeroBigInt,
			}
		}
	}
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
		if s.dryRun {
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
	if s.dryRun {
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

func (s *StateDB) IntermediateRoot(cleanDirty bool, stateDataCopy *StateDataCopy) error {
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
				index, leaf, err := s.updateAccountTree(accountIndex, assets, stateDataCopy)
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
				index, leaf, err := s.updateNftTree(nftIndex, stateDataCopy)
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
	//todo for stress
	stateDataCopy.pendingAccountSmtItem = pendingAccountItem
	stateDataCopy.pendingNftSmtItem = pendingNftItem

	//err := gopool.Submit(func() {
	//	resultChan <- &treeUpdateResp{
	//		role: accountTreeRole,
	//		err:  s.AccountTree.MultiSet(pendingAccountItem),
	//	}
	//})
	//if err != nil {
	//	return err
	//}
	//err = gopool.Submit(func() {
	//	resultChan <- &treeUpdateResp{
	//		role: nftTreeRole,
	//		err:  s.NftTree.MultiSet(pendingNftItem),
	//	}
	//})
	//if err != nil {
	//	return err
	//}
	//
	//start := time.Now()
	//for i := 0; i < 2; i++ {
	//	result := <-resultChan
	//	if result.err != nil {
	//		return fmt.Errorf("update %s tree failed, %v", result.role, result.err)
	//	}
	//}
	//s.Metrics.AccountTreeMultiSetGauge.Set(float64(time.Since(start).Milliseconds()))
	//
	//hFunc := poseidon.NewPoseidon()
	//hFunc.Write(s.AccountTree.Root())
	//hFunc.Write(s.NftTree.Root())
	//stateDataCopy.StateCache.StateRoot = common.Bytes2Hex(hFunc.Sum(nil))
	return nil
}

func (s *StateDB) AccountTreeAndNftTreeMultiSet(stateDataCopy *StateDataCopy) error {
	start := time.Now()
	resultChan := make(chan *treeUpdateResp, 1)
	defer close(resultChan)
	err := gopool.Submit(func() {
		resultChan <- &treeUpdateResp{
			role: accountTreeRole,
			err:  s.AccountTree.MultiSet(stateDataCopy.pendingAccountSmtItem),
		}
	})
	if err != nil {
		return err
	}
	err = gopool.Submit(func() {
		resultChan <- &treeUpdateResp{
			role: nftTreeRole,
			err:  s.NftTree.MultiSet(stateDataCopy.pendingNftSmtItem),
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
	s.Metrics.AccountTreeMultiSetGauge.Set(float64(time.Since(start).Milliseconds()))

	hFunc := poseidon.NewPoseidon()
	hFunc.Write(s.AccountTree.Root())
	hFunc.Write(s.NftTree.Root())
	stateDataCopy.StateCache.StateRoot = common.Bytes2Hex(hFunc.Sum(nil))
	return nil
}

func (s *StateDB) updateAccountTree(accountIndex int64, assets []int64, stateCopy *StateDataCopy) (int64, []byte, error) {
	start := time.Now()
	account, exist := stateCopy.StateCache.GetPendingAccount(accountIndex)
	s.Metrics.AccountTreeGauge.WithLabelValues("cache_get_account").Set(float64(time.Since(start).Milliseconds()))
	if !exist {
		//todo
	}
	start = time.Now()
	isGasAccount := accountIndex == types.GasAccount
	pendingUpdateAssetItem := make([]bsmt.Item, 0, len(assets))
	s.Metrics.AccountTreeGauge.WithLabelValues("assets_count").Set(float64(len(assets)))
	for _, assetId := range assets {
		isGasAsset := false
		if isGasAccount {
			for _, gasAssetId := range types.GasAssets {
				if assetId == gasAssetId {
					isGasAsset = true
					break
				}
			}
		}
		balance := account.AssetInfo[assetId].Balance
		if isGasAsset {
			balance = ffmath.Add(balance, s.GetPendingGas(assetId))
		}
		startItem := time.Now()
		assetLeaf, err := tree.ComputeAccountAssetLeafHash(
			balance.String(),
			account.AssetInfo[assetId].OfferCanceledOrFinalized.String(),
		)
		s.Metrics.AccountTreeGauge.WithLabelValues("compute_poseidon").Set(float64(time.Since(startItem).Milliseconds()))
		if err != nil {
			return accountIndex, nil, fmt.Errorf("compute new account asset leaf failed: %v", err)
		}
		pendingUpdateAssetItem = append(pendingUpdateAssetItem, bsmt.Item{Key: uint64(assetId), Val: assetLeaf})
	}
	s.Metrics.AccountTreeGauge.WithLabelValues("for_assets").Set(float64(time.Since(start).Milliseconds()))

	start = time.Now()
	err := s.AccountAssetTrees.Get(accountIndex).MultiSet(pendingUpdateAssetItem)
	if err != nil {
		return accountIndex, nil, fmt.Errorf("update asset tree failed: %v", err)
	}
	s.Metrics.AccountTreeGauge.WithLabelValues("multiSet").Set(float64(time.Since(start).Milliseconds()))

	account.AssetRoot = common.Bytes2Hex(s.AccountAssetTrees.Get(accountIndex).Root())
	nAccountLeafHash, err := tree.ComputeAccountLeafHash(
		account.AccountNameHash,
		account.PublicKey,
		account.Nonce,
		account.CollectionNonce,
		s.AccountAssetTrees.Get(accountIndex).Root(),
	)
	if err != nil {
		return accountIndex, nil, fmt.Errorf("unable to compute account leaf: %v", err)
	}
	//todo for tress
	asset := s.AccountAssetTrees.Get(accountIndex)
	version := bsmt.Version(uint64(asset.LatestVersion()) - 20)
	//todo recentVersion
	//todo check
	//version := bsmt.Version(s.GetPrunedBlockHeight())
	//logx.Infof("asset.Commit: %d", curHeight)

	ver, err := asset.Commit(&version)
	if err != nil {
		logx.Error("asset.Commit failed:", err)
		return accountIndex, nil, fmt.Errorf("unable to commit asset tree [%d], tree ver: %d, prune ver: %d,error:%s", accountIndex, ver, version, err.Error())
	}
	return accountIndex, nAccountLeafHash, nil
}

func (s *StateDB) updateNftTree(nftIndex int64, stateCopy *StateDataCopy) (int64, []byte, error) {
	start := time.Now()
	nft, exist := stateCopy.StateCache.GetPendingNft(nftIndex)
	if !exist {
		//todo
	}

	nftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
		nft.CreatorAccountIndex,
		nft.OwnerAccountIndex,
		nft.NftContentHash,
		nft.CreatorTreasuryRate,
		nft.CollectionId,
	)
	if err != nil {
		return nftIndex, nil, fmt.Errorf("unable to compute nft leaf: %v", err)
	}
	s.Metrics.NftTreeGauge.WithLabelValues("nft").Set(float64(time.Since(start).Milliseconds()))
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

func (s *StateDB) GetNextAccountIndex() int64 {
	return s.AccountAssetTrees.GetNextAccountIndex()
}

func (s *StateDB) GetNextNftIndex() int64 {
	maxNftIndex, err := s.chainDb.L2NftModel.GetLatestNftIndex()
	if err != nil {
		panic("get latest nft index error: " + err.Error())
	}

	for index := range s.PendingNftMap {
		if index > maxNftIndex {
			maxNftIndex = index
		}
	}
	return maxNftIndex + 1
}

//func (s *StateDB) GetGasAccountIndex() (int64, error) {
//	gasAccountIndex := int64(-1)
//	_, err := s.redisCache.Get(context.Background(), dbcache.GasAccountKey, &gasAccountIndex)
//	if err == nil {
//		return gasAccountIndex, nil
//	}
//	logx.Errorf("fail to get gas account from cache, error: %s", err.Error())
//
//	gasAccountConfig, err := s.chainDb.SysConfigModel.GetSysConfigByName(types.GasAccountIndex)
//	if err != nil {
//		logx.Errorf("cannot find config for: %s", types.GasAccountIndex)
//		return -1, errors.New("internal error")
//	}
//	gasAccountIndex, err = strconv.ParseInt(gasAccountConfig.Value, 10, 64)
//	if err != nil {
//		logx.Errorf("invalid account index: %s", gasAccountConfig.Value)
//		return -1, errors.New("internal error")
//	}
//	_ = s.redisCache.Set(context.Background(), dbcache.GasAccountKey, gasAccountIndex)
//	return gasAccountIndex, nil
//}

//todo for stress
func (s *StateDB) GetGasAccountIndex() (int64, error) {
	return int64(1), nil
}

//func (s *StateDB) GetGasConfig() (map[uint32]map[int]int64, error) {
//	gasFeeValue := ""
//	_, err := s.redisCache.Get(context.Background(), dbcache.GasConfigKey, &gasFeeValue)
//	if err != nil {
//		logx.Errorf("fail to get gas config from cache, error: %s", err.Error())
//
//		cfgGasFee, err := s.chainDb.SysConfigModel.GetSysConfigByName(types.SysGasFee)
//		if err != nil {
//			logx.Errorf("cannot find gas asset: %s", err.Error())
//			return nil, errors.New("invalid gas fee asset")
//		}
//		gasFeeValue = cfgGasFee.Value
//		_ = s.redisCache.Set(context.Background(), dbcache.GasConfigKey, gasFeeValue)
//	}
//
//	m := make(map[uint32]map[int]int64)
//	err = json.Unmarshal([]byte(gasFeeValue), &m)
//	if err != nil {
//		logx.Errorf("fail to unmarshal gas fee config, err: %s", err.Error())
//		return nil, errors.New("internal error")
//	}
//
//	return m, nil
//}

//todo for stress
func (s *StateDB) GetGasConfig() (map[uint32]map[int]int64, error) {
	gasFeeValue := "{\"0\":{\"10\":12000000000000,\"11\":20000000000000,\"4\":10000000000000,\"5\":20000000000000,\"6\":10000000000000,\"7\":10000000000000,\"8\":12000000000000,\"9\":18000000000000},\"1\":{\"10\":12000000000000,\"11\":20000000000000,\"4\":10000000000000,\"5\":20000000000000,\"6\":10000000000000,\"7\":10000000000000,\"8\":12000000000000,\"9\":18000000000000}}"

	m := make(map[uint32]map[int]int64)
	err := json.Unmarshal([]byte(gasFeeValue), &m)
	if err != nil {
		logx.Errorf("fail to unmarshal gas fee config, err: %s", err.Error())
		return nil, errors.New("internal error")
	}

	return m, nil
}

func (c *StateDB) UpdatePrunedBlockHeight(latestBlock int64) {
	c.mainLock.Lock()
	if c.prunedBlockHeight < latestBlock {
		c.prunedBlockHeight = latestBlock
	}
	c.mainLock.Unlock()
}

func (c *StateDB) GetPrunedBlockHeight() int64 {
	c.mainLock.RLock()
	defer c.mainLock.RUnlock()
	return c.prunedBlockHeight
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
