package statedb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	lru "github.com/hashicorp/golang-lru"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/common/chain"
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
	pending, exist := s.StateCache.GetPendingAccount(accountIndex)
	if exist {
		return pending, nil
	}

	cached, exist := s.AccountCache.Get(accountIndex)
	if exist {
		return cached.(*types.AccountInfo), nil
	}

	account, err := s.chainDb.AccountModel.GetAccountByIndex(accountIndex)
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

func (s *StateDB) syncPendingAccount(pendingAccount map[int64]*types.AccountInfo) error {
	for index, formatAccount := range pendingAccount {
		account, err := chain.FromFormatAccountInfo(formatAccount)
		if err != nil {
			return err
		}
		err = s.redisCache.Set(context.Background(), dbcache.AccountKeyByIndex(index), account)
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		s.AccountCache.Add(index, formatAccount)
	}

	return nil
}

func (s *StateDB) syncPendingNft(pendingNft map[int64]*nft.L2Nft) error {
	for index, nft := range pendingNft {
		err := s.redisCache.Set(context.Background(), dbcache.NftKeyByIndex(index), nft)
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		s.NftCache.Add(index, nft)
	}
	return nil
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
	// Sync pending to cache.
	err := s.syncPendingAccount(s.PendingAccountMap)
	if err != nil {
		return err
	}
	err = s.syncPendingNft(s.PendingNftMap)
	if err != nil {
		return err
	}

	return nil
}

func (s *StateDB) PurgeCache(stateRoot string) {
	s.StateCache = NewStateCache(stateRoot)
}

func (s *StateDB) GetPendingAccount(blockHeight int64) ([]*account.Account, []*account.AccountHistory, error) {
	pendingAccount := make([]*account.Account, 0)
	pendingAccountHistory := make([]*account.AccountHistory, 0)

	for _, formatAccount := range s.PendingAccountMap {
		if formatAccount.AccountIndex == types.GasAccount {
			s.applyGasUpdate(formatAccount)
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

func (s *StateDB) applyGasUpdate(formatAccount *types.AccountInfo) {
	for assetId, delta := range s.StateCache.PendingGasMap {
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

func (s *StateDB) GetPendingNft(blockHeight int64) ([]*nft.L2Nft, []*nft.L2NftHistory, error) {
	pendingNft := make([]*nft.L2Nft, 0)
	pendingNftHistory := make([]*nft.L2NftHistory, 0)

	for _, newNft := range s.PendingNftMap {
		pendingNft = append(pendingNft, newNft)
		pendingNftHistory = append(pendingNftHistory, &nft.L2NftHistory{
			NftIndex:            newNft.NftIndex,
			CreatorAccountIndex: newNft.CreatorAccountIndex,
			OwnerAccountIndex:   newNft.OwnerAccountIndex,
			NftContentHash:      newNft.NftContentHash,
			NftL1Address:        newNft.NftL1Address,
			NftL1TokenId:        newNft.NftL1TokenId,
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

func (s *StateDB) IntermediateRoot(taskPool *ants.Pool, cleanDirty bool) error {
	taskNum := 0
	resultChan := make(chan *treeUpdateResp, 1)

	for accountIndex, assetsMap := range s.dirtyAccountsAndAssetsMap {
		assets := make([]int64, 0, len(assetsMap))
		for assetIndex, isDirty := range assetsMap {
			if !isDirty {
				continue
			}
			assets = append(assets, assetIndex)
		}
		taskNum++
		func(accountIndex int64, assets []int64) {
			taskPool.Submit(func() {
				index, leaf, err := s.updateAccountTree(accountIndex, assets)
				resultChan <- &treeUpdateResp{
					role:  accountTreeRole,
					index: index,
					leaf:  leaf,
					err:   err,
				}
			})
		}(accountIndex, assets)
	}

	for nftIndex, isDirty := range s.dirtyNftMap {
		if !isDirty {
			continue
		}
		taskNum++
		func(nftIndex int64) {
			taskPool.Submit(func() {
				index, leaf, err := s.updateNftTree(nftIndex)
				resultChan <- &treeUpdateResp{
					role:  nftTreeRole,
					index: index,
					leaf:  leaf,
					err:   err,
				}
			})
		}(nftIndex)
	}

	if cleanDirty {
		s.dirtyAccountsAndAssetsMap = make(map[int64]map[int64]bool, 0)
		s.dirtyNftMap = make(map[int64]bool, 0)
	}

	pendingAccountItem := make([]bsmt.Item, 0, len(s.dirtyAccountsAndAssetsMap))
	pendingNftItem := make([]bsmt.Item, 0, len(s.dirtyNftMap))
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
	errChan := make(chan error, 2)
	taskPool.Submit(func() {
		errChan <- s.AccountTree.MultiSet(pendingAccountItem...)
	})
	taskPool.Submit(func() {
		errChan <- s.NftTree.MultiSet(pendingNftItem...)
	})
	for i := 0; i < 2; i++ {
		err := <-errChan
		if err != nil {
			return fmt.Errorf("update tree failed, %v", err)
		}
	}

	hFunc := mimc.NewMiMC()
	hFunc.Write(s.AccountTree.Root())
	hFunc.Write(s.NftTree.Root())
	s.StateRoot = common.Bytes2Hex(hFunc.Sum(nil))
	return nil
}

func (s *StateDB) updateAccountTree(accountIndex int64, assets []int64) (int64, []byte, error) {
	account, err := s.GetFormatAccount(accountIndex)
	if err != nil {
		return accountIndex, nil, err
	}
	isGasAccount := accountIndex == types.GasAccount
	pendingUpdateAssetItem := make([]bsmt.Item, 0, len(assets))
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
		assetLeaf, err := tree.ComputeAccountAssetLeafHash(
			balance.String(),
			account.AssetInfo[assetId].OfferCanceledOrFinalized.String(),
		)
		if err != nil {
			return accountIndex, nil, fmt.Errorf("compute new account asset leaf failed: %v", err)
		}
		pendingUpdateAssetItem = append(pendingUpdateAssetItem, bsmt.Item{Key: uint64(assetId), Val: assetLeaf})
	}

	err = s.AccountAssetTrees.Get(accountIndex).MultiSet(pendingUpdateAssetItem...)
	if err != nil {
		return accountIndex, nil, fmt.Errorf("update asset tree failed: %v", err)
	}

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

	return accountIndex, nAccountLeafHash, nil
}

func (s *StateDB) updateNftTree(nftIndex int64) (int64, []byte, error) {
	nft, err := s.GetNft(nftIndex)
	if err != nil {
		return nftIndex, nil, err
	}
	nftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
		nft.CreatorAccountIndex,
		nft.OwnerAccountIndex,
		nft.NftContentHash,
		nft.NftL1Address,
		nft.NftL1TokenId,
		nft.CreatorTreasuryRate,
		nft.CollectionId,
	)
	if err != nil {
		return nftIndex, nil, fmt.Errorf("unable to compute nft leaf: %v", err)
	}

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

func (s *StateDB) GetGasAccountIndex() (int64, error) {
	gasAccountIndex := int64(-1)
	_, err := s.redisCache.Get(context.Background(), dbcache.GasAccountKey, &gasAccountIndex)
	if err == nil {
		return gasAccountIndex, nil
	}
	logx.Errorf("fail to get gas account from cache, error: %s", err.Error())

	gasAccountConfig, err := s.chainDb.SysConfigModel.GetSysConfigByName(types.GasAccountIndex)
	if err != nil {
		logx.Errorf("cannot find config for: %s", types.GasAccountIndex)
		return -1, errors.New("internal error")
	}
	gasAccountIndex, err = strconv.ParseInt(gasAccountConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("invalid account index: %s", gasAccountConfig.Value)
		return -1, errors.New("internal error")
	}
	_ = s.redisCache.Set(context.Background(), dbcache.GasAccountKey, gasAccountIndex)
	return gasAccountIndex, nil
}

func (s *StateDB) GetGasConfig() (map[uint32]map[int]int64, error) {
	gasFeeValue := ""
	_, err := s.redisCache.Get(context.Background(), dbcache.GasConfigKey, &gasFeeValue)
	if err != nil {
		logx.Errorf("fail to get gas config from cache, error: %s", err.Error())

		cfgGasFee, err := s.chainDb.SysConfigModel.GetSysConfigByName(types.SysGasFee)
		if err != nil {
			logx.Errorf("cannot find gas asset: %s", err.Error())
			return nil, errors.New("invalid gas fee asset")
		}
		gasFeeValue = cfgGasFee.Value
		_ = s.redisCache.Set(context.Background(), dbcache.GasConfigKey, gasFeeValue)
	}

	m := make(map[uint32]map[int]int64)
	err = json.Unmarshal([]byte(gasFeeValue), &m)
	if err != nil {
		logx.Errorf("fail to unmarshal gas fee config, err: %s", err.Error())
		return nil, errors.New("internal error")
	}

	return m, nil
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
