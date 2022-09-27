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
	"github.com/zeromicro/go-zero/core/logx"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

var (
	DefaultCacheConfig = CacheConfig{
		AccountCacheSize:   2048,
		LiquidityCacheSize: 2048,
		NftCacheSize:       2048,
	}
)

type CacheConfig struct {
	AccountCacheSize   int
	LiquidityCacheSize int
	NftCacheSize       int
}

func (c *CacheConfig) sanitize() *CacheConfig {
	if c.AccountCacheSize <= 0 {
		c.AccountCacheSize = DefaultCacheConfig.AccountCacheSize
	}

	if c.LiquidityCacheSize <= 0 {
		c.LiquidityCacheSize = DefaultCacheConfig.LiquidityCacheSize
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
	AccountCache   *lru.Cache
	LiquidityCache *lru.Cache
	NftCache       *lru.Cache

	// Tree state
	AccountTree       bsmt.SparseMerkleTree
	LiquidityTree     bsmt.SparseMerkleTree
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
		chainDb.AccountHistoryModel,
		curHeight,
		treeCtx,
		assetCacheSize,
	)
	if err != nil {
		logx.Error("dbinitializer account tree failed:", err)
		return nil, err
	}
	liquidityTree, err := tree.InitLiquidityTree(
		chainDb.LiquidityHistoryModel,
		curHeight,
		treeCtx,
	)
	if err != nil {
		logx.Error("dbinitializer liquidity tree failed:", err)
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
	liquidityCache, err := lru.New(cacheConfig.LiquidityCacheSize)
	if err != nil {
		logx.Error("init liquidity cache failed:", err)
		return nil, err
	}
	nftCache, err := lru.New(cacheConfig.NftCacheSize)
	if err != nil {
		logx.Error("init nft cache failed:", err)
		return nil, err
	}

	return &StateDB{
		StateCache:     NewStateCache(stateRoot),
		chainDb:        chainDb,
		redisCache:     redisCache,
		AccountCache:   accountCache,
		LiquidityCache: liquidityCache,
		NftCache:       nftCache,

		AccountTree:       accountTree,
		LiquidityTree:     liquidityTree,
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
	liquidityCache, err := lru.New(cacheConfig.LiquidityCacheSize)
	if err != nil {
		logx.Error("init liquidity cache failed:", err)
		return nil, err
	}
	nftCache, err := lru.New(cacheConfig.NftCacheSize)
	if err != nil {
		logx.Error("init nft cache failed:", err)
		return nil, err
	}
	return &StateDB{
		dryRun:         true,
		redisCache:     redisCache,
		chainDb:        chainDb,
		AccountCache:   accountCache,
		LiquidityCache: liquidityCache,
		NftCache:       nftCache,
		StateCache:     NewStateCache(""),
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
	if err != nil {
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

func (s *StateDB) GetLiquidity(pairIndex int64) (*liquidity.Liquidity, error) {
	pending, exist := s.StateCache.GetPendingLiquidity(pairIndex)
	if exist {
		return pending, nil
	}
	cached, exist := s.LiquidityCache.Get(pairIndex)
	if exist {
		return cached.(*liquidity.Liquidity), nil
	}
	liquidity, err := s.chainDb.LiquidityModel.GetLiquidityByIndex(pairIndex)
	if err != nil {
		return nil, err
	}
	s.LiquidityCache.Add(pairIndex, liquidity)
	return liquidity, nil
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
	if err != nil {
		return nil, err
	}
	s.NftCache.Add(nftIndex, nft)
	return nft, nil
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

func (s *StateDB) syncPendingLiquidity(pendingLiquidity map[int64]*liquidity.Liquidity) error {
	for index, liquidity := range pendingLiquidity {
		err := s.redisCache.Set(context.Background(), dbcache.LiquidityKeyByIndex(index), liquidity)
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		s.LiquidityCache.Add(index, liquidity)
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

func (s *StateDB) SyncStateCacheToRedis() error {
	// Sync new create to cache.
	err := s.syncPendingAccount(s.PendingNewAccountMap)
	if err != nil {
		return err
	}
	err = s.syncPendingLiquidity(s.PendingNewLiquidityMap)
	if err != nil {
		return err
	}
	err = s.syncPendingNft(s.PendingNewNftMap)
	if err != nil {
		return err
	}

	// Sync pending update to cache.
	err = s.syncPendingAccount(s.PendingUpdateAccountMap)
	if err != nil {
		return err
	}
	err = s.syncPendingLiquidity(s.PendingUpdateLiquidityMap)
	if err != nil {
		return err
	}
	err = s.syncPendingNft(s.PendingUpdateNftMap)
	if err != nil {
		return err
	}

	return nil
}

func (s *StateDB) PurgeCache(stateRoot string) {
	s.StateCache = NewStateCache(stateRoot)
}

func (s *StateDB) GetPendingAccount(blockHeight int64) ([]*account.Account, []*account.Account, []*account.AccountHistory, error) {
	pendingNewAccount := make([]*account.Account, 0)
	pendingUpdateAccount := make([]*account.Account, 0)
	pendingNewAccountHistory := make([]*account.AccountHistory, 0)

	for _, formatAccount := range s.PendingNewAccountMap {
		newAccount, err := chain.FromFormatAccountInfo(formatAccount)
		if err != nil {
			return nil, nil, nil, err
		}

		pendingNewAccount = append(pendingNewAccount, newAccount)
		pendingNewAccountHistory = append(pendingNewAccountHistory, &account.AccountHistory{
			AccountIndex:    newAccount.AccountIndex,
			Nonce:           newAccount.Nonce,
			CollectionNonce: newAccount.CollectionNonce,
			AssetInfo:       newAccount.AssetInfo,
			AssetRoot:       newAccount.AssetRoot,
			L2BlockHeight:   blockHeight, // TODO: ensure this should be the new block's height.
		})
	}

	for index, formatAccount := range s.PendingUpdateAccountMap {
		if _, exist := s.PendingNewAccountMap[index]; exist {
			continue
		}

		newAccount, err := chain.FromFormatAccountInfo(formatAccount)
		if err != nil {
			return nil, nil, nil, err
		}
		pendingUpdateAccount = append(pendingUpdateAccount, newAccount)
		pendingNewAccountHistory = append(pendingNewAccountHistory, &account.AccountHistory{
			AccountIndex:    newAccount.AccountIndex,
			Nonce:           newAccount.Nonce,
			CollectionNonce: newAccount.CollectionNonce,
			AssetInfo:       newAccount.AssetInfo,
			AssetRoot:       newAccount.AssetRoot,
			L2BlockHeight:   blockHeight, // TODO: ensure this should be the new block's height.
		})
	}

	return pendingNewAccount, pendingUpdateAccount, pendingNewAccountHistory, nil
}

func (s *StateDB) GetPendingLiquidity(blockHeight int64) ([]*liquidity.Liquidity, []*liquidity.Liquidity, []*liquidity.LiquidityHistory, error) {
	pendingNewLiquidity := make([]*liquidity.Liquidity, 0)
	pendingUpdateLiquidity := make([]*liquidity.Liquidity, 0)
	pendingNewLiquidityHistory := make([]*liquidity.LiquidityHistory, 0)

	for _, newLiquidity := range s.PendingNewLiquidityMap {
		pendingNewLiquidity = append(pendingNewLiquidity, newLiquidity)
		pendingNewLiquidityHistory = append(pendingNewLiquidityHistory, &liquidity.LiquidityHistory{
			PairIndex:            newLiquidity.PairIndex,
			AssetAId:             newLiquidity.AssetAId,
			AssetA:               newLiquidity.AssetA,
			AssetBId:             newLiquidity.AssetBId,
			AssetB:               newLiquidity.AssetB,
			LpAmount:             newLiquidity.LpAmount,
			KLast:                newLiquidity.KLast,
			FeeRate:              newLiquidity.FeeRate,
			TreasuryAccountIndex: newLiquidity.TreasuryAccountIndex,
			TreasuryRate:         newLiquidity.TreasuryRate,
			L2BlockHeight:        blockHeight,
		})
	}

	for index, newLiquidity := range s.PendingUpdateLiquidityMap {
		if _, exist := s.PendingNewLiquidityMap[index]; exist {
			continue
		}

		pendingUpdateLiquidity = append(pendingUpdateLiquidity, newLiquidity)
		pendingNewLiquidityHistory = append(pendingNewLiquidityHistory, &liquidity.LiquidityHistory{
			PairIndex:            newLiquidity.PairIndex,
			AssetAId:             newLiquidity.AssetAId,
			AssetA:               newLiquidity.AssetA,
			AssetBId:             newLiquidity.AssetBId,
			AssetB:               newLiquidity.AssetB,
			LpAmount:             newLiquidity.LpAmount,
			KLast:                newLiquidity.KLast,
			FeeRate:              newLiquidity.FeeRate,
			TreasuryAccountIndex: newLiquidity.TreasuryAccountIndex,
			TreasuryRate:         newLiquidity.TreasuryRate,
			L2BlockHeight:        blockHeight,
		})
	}

	return pendingNewLiquidity, pendingUpdateLiquidity, pendingNewLiquidityHistory, nil
}

func (s *StateDB) GetPendingNft(blockHeight int64) ([]*nft.L2Nft, []*nft.L2Nft, []*nft.L2NftHistory, error) {
	pendingNewNft := make([]*nft.L2Nft, 0)
	pendingUpdateNft := make([]*nft.L2Nft, 0)
	pendingNewNftHistory := make([]*nft.L2NftHistory, 0)

	for _, newNft := range s.PendingNewNftMap {
		pendingNewNft = append(pendingNewNft, newNft)
		pendingNewNftHistory = append(pendingNewNftHistory, &nft.L2NftHistory{
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

	for index, newNft := range s.PendingUpdateNftMap {
		if _, exist := s.PendingNewNftMap[index]; exist {
			continue
		}
		pendingUpdateNft = append(pendingUpdateNft, newNft)
		pendingNewNftHistory = append(pendingNewNftHistory, &nft.L2NftHistory{
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

	return pendingNewNft, pendingUpdateNft, pendingNewNftHistory, nil
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
		accountCopy, err := account.DeepCopy()
		if err != nil {
			return nil, err
		}
		accounts[accountId] = accountCopy
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
					LpAmount:                 types.ZeroBigInt,
					OfferCanceledOrFinalized: types.ZeroBigInt,
				}
			}
		}
		s.AccountCache.Add(accountIndex, account)
	}

	return nil
}

func (s *StateDB) PrepareLiquidity(pairIndex int64) error {
	if s.dryRun {
		l := &liquidity.Liquidity{}
		redisLiquidity, err := s.redisCache.Get(context.Background(), dbcache.LiquidityKeyByIndex(pairIndex), l)
		if err == nil && redisLiquidity != nil {
			s.LiquidityCache.Add(pairIndex, l)
		}
	}

	_, err := s.GetLiquidity(pairIndex)
	if err != nil {
		return err
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

func (s *StateDB) IntermediateRoot(cleanDirty bool) error {
	for accountIndex, assetsMap := range s.dirtyAccountsAndAssetsMap {
		assets := make([]int64, 0, len(assetsMap))
		for assetIndex, isDirty := range assetsMap {
			if !isDirty {
				continue
			}
			assets = append(assets, assetIndex)
		}

		err := s.updateAccountTree(accountIndex, assets)
		if err != nil {
			return err
		}
	}

	for pairIndex, isDirty := range s.dirtyLiquidityMap {
		if !isDirty {
			continue
		}
		err := s.updateLiquidityTree(pairIndex)
		if err != nil {
			return err
		}
	}

	for nftIndex, isDirty := range s.dirtyNftMap {
		if !isDirty {
			continue
		}
		err := s.updateNftTree(nftIndex)
		if err != nil {
			return err
		}
	}

	if cleanDirty {
		s.dirtyAccountsAndAssetsMap = make(map[int64]map[int64]bool, 0)
		s.dirtyLiquidityMap = make(map[int64]bool, 0)
		s.dirtyNftMap = make(map[int64]bool, 0)
	}

	hFunc := mimc.NewMiMC()
	hFunc.Write(s.AccountTree.Root())
	hFunc.Write(s.LiquidityTree.Root())
	hFunc.Write(s.NftTree.Root())
	s.StateRoot = common.Bytes2Hex(hFunc.Sum(nil))
	return nil
}

func (s *StateDB) updateAccountTree(accountIndex int64, assets []int64) error {
	account, err := s.GetFormatAccount(accountIndex)
	if err != nil {
		return err
	}
	for _, assetId := range assets {
		assetLeaf, err := tree.ComputeAccountAssetLeafHash(
			account.AssetInfo[assetId].Balance.String(),
			account.AssetInfo[assetId].LpAmount.String(),
			account.AssetInfo[assetId].OfferCanceledOrFinalized.String(),
		)
		if err != nil {
			return fmt.Errorf("compute new account asset leaf failed: %v", err)
		}
		err = s.AccountAssetTrees.Get(accountIndex).Set(uint64(assetId), assetLeaf)
		if err != nil {
			return fmt.Errorf("update asset tree failed: %v", err)
		}
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
		return fmt.Errorf("unable to compute account leaf: %v", err)
	}
	err = s.AccountTree.Set(uint64(accountIndex), nAccountLeafHash)
	if err != nil {
		return fmt.Errorf("unable to update account tree: %v", err)
	}

	return nil
}

func (s *StateDB) updateLiquidityTree(pairIndex int64) error {
	liquidity, err := s.GetLiquidity(pairIndex)
	if err != nil {
		return err
	}
	nLiquidityAssetLeaf, err := tree.ComputeLiquidityAssetLeafHash(
		liquidity.AssetAId,
		liquidity.AssetA,
		liquidity.AssetBId,
		liquidity.AssetB,
		liquidity.LpAmount,
		liquidity.KLast,
		liquidity.FeeRate,
		liquidity.TreasuryAccountIndex,
		liquidity.TreasuryRate,
	)
	if err != nil {
		return fmt.Errorf("unable to compute liquidity leaf: %v", err)
	}
	err = s.LiquidityTree.Set(uint64(pairIndex), nLiquidityAssetLeaf)
	if err != nil {
		return fmt.Errorf("unable to update liquidity tree: %v", err)
	}

	return nil
}

func (s *StateDB) updateNftTree(nftIndex int64) error {
	nft, err := s.GetNft(nftIndex)
	if err != nil {
		return err
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
		return fmt.Errorf("unable to compute nft leaf: %v", err)
	}
	err = s.NftTree.Set(uint64(nftIndex), nftAssetLeaf)
	if err != nil {
		return fmt.Errorf("unable to update nft tree: %v", err)
	}

	return nil
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
	if len(s.PendingNewNftMap) == 0 {
		maxNftIndex, err := s.chainDb.L2NftModel.GetLatestNftIndex()
		if err != nil {
			panic("get latest nft index error: " + err.Error())
		}
		return maxNftIndex + 1
	}

	maxNftIndex := int64(-1)
	for index := range s.PendingNewNftMap {
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
