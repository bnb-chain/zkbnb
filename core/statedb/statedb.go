package statedb

import (
	"context"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	bsmt "github.com/bnb-chain/zkbas-smt"
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/dbcache"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/tree"
	"github.com/bnb-chain/zkbas/types"
)

type StateDB struct {
	dryRun bool
	// State cache
	*StateCache
	chainDb    *ChainDB
	redisCache dbcache.Cache

	// Flat state
	AccountMap   map[int64]*types.AccountInfo
	LiquidityMap map[int64]*liquidity.Liquidity
	NftMap       map[int64]*nft.L2Nft

	// Tree state
	AccountTree       bsmt.SparseMerkleTree
	LiquidityTree     bsmt.SparseMerkleTree
	NftTree           bsmt.SparseMerkleTree
	AccountAssetTrees []bsmt.SparseMerkleTree
	TreeCtx           *tree.Context
}

func NewStateDB(treeCtx *tree.Context, chainDb *ChainDB, redisCache dbcache.Cache, stateRoot string, curHeight int64) (*StateDB, error) {
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
	return &StateDB{
		StateCache:   NewStateCache(stateRoot),
		chainDb:      chainDb,
		redisCache:   redisCache,
		AccountMap:   make(map[int64]*types.AccountInfo),
		LiquidityMap: make(map[int64]*liquidity.Liquidity),
		NftMap:       make(map[int64]*nft.L2Nft),

		AccountTree:       accountTree,
		LiquidityTree:     liquidityTree,
		NftTree:           nftTree,
		AccountAssetTrees: accountAssetTrees,
		TreeCtx:           treeCtx,
	}, nil
}

func NewStateDBForDryRun(redisCache dbcache.Cache, chainDb *ChainDB) *StateDB {
	return &StateDB{
		dryRun:       true,
		redisCache:   redisCache,
		chainDb:      chainDb,
		AccountMap:   make(map[int64]*types.AccountInfo),
		LiquidityMap: make(map[int64]*liquidity.Liquidity),
		NftMap:       make(map[int64]*nft.L2Nft),
		StateCache:   NewStateCache(""),
	}
}

func (s *StateDB) GetAccount(accountIndex int64) interface{} {
	// to save account to cache, we need to convert it
	account, err := chain.FromFormatAccountInfo(s.AccountMap[accountIndex])
	if err != nil {
		return nil
	}
	return account
}

func (s *StateDB) GetLiquidity(pairIndex int64) interface{} {
	return s.LiquidityMap[pairIndex]
}

func (s *StateDB) GetNft(nftIndex int64) interface{} {
	return s.NftMap[nftIndex]
}

func (s *StateDB) syncPendingStateToRedis(pendingMap map[int64]int, getKey func(int64) string, getValue func(int64) interface{}) error {
	for index, status := range pendingMap {
		if status != StateCachePending {
			continue
		}

		err := s.redisCache.Set(context.Background(), getKey(index), getValue(index))
		if err != nil {
			return fmt.Errorf("cache to redis failed: %v", err)
		}
		pendingMap[index] = StateCacheCached
	}

	return nil
}

func (s *StateDB) SyncStateCacheToRedis() error {

	// Sync new create to cache.
	err := s.syncPendingStateToRedis(s.PendingNewAccountIndexMap, dbcache.AccountKeyByIndex, s.GetAccount)
	if err != nil {
		return err
	}
	err = s.syncPendingStateToRedis(s.PendingNewLiquidityIndexMap, dbcache.LiquidityKeyByIndex, s.GetLiquidity)
	if err != nil {
		return err
	}
	err = s.syncPendingStateToRedis(s.PendingNewNftIndexMap, dbcache.NftKeyByIndex, s.GetNft)
	if err != nil {
		return err
	}

	// Sync pending update to cache.
	err = s.syncPendingStateToRedis(s.PendingUpdateAccountIndexMap, dbcache.AccountKeyByIndex, s.GetAccount)
	if err != nil {
		return err
	}
	err = s.syncPendingStateToRedis(s.PendingUpdateLiquidityIndexMap, dbcache.LiquidityKeyByIndex, s.GetLiquidity)
	if err != nil {
		return err
	}
	err = s.syncPendingStateToRedis(s.PendingUpdateNftIndexMap, dbcache.NftKeyByIndex, s.GetNft)
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

	for index, status := range s.PendingNewAccountIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in Statedb cache")
			continue
		}

		newAccount, err := chain.FromFormatAccountInfo(s.AccountMap[index])
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

	for index, status := range s.PendingUpdateAccountIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in Statedb cache")
			continue
		}

		if _, exist := s.PendingNewAccountIndexMap[index]; exist {
			continue
		}

		newAccount, err := chain.FromFormatAccountInfo(s.AccountMap[index])
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

	for index, status := range s.PendingNewLiquidityIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in Statedb cache")
			continue
		}

		newLiquidity := s.LiquidityMap[index]
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

	for index, status := range s.PendingUpdateLiquidityIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in Statedb cache")
			continue
		}

		if _, exist := s.PendingNewLiquidityIndexMap[index]; exist {
			continue
		}

		newLiquidity := s.LiquidityMap[index]
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

	for index, status := range s.PendingNewNftIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in Statedb cache")
			continue
		}

		newNft := s.NftMap[index]
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

	for index, status := range s.PendingUpdateNftIndexMap {
		if status < StateCachePending {
			logx.Errorf("unexpected 0 status in Statedb cache")
			continue
		}

		if _, exist := s.PendingNewNftIndexMap[index]; exist {
			continue
		}

		newNft := s.NftMap[index]
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

		accountCopy, err := s.AccountMap[accountId].DeepCopy()
		if err != nil {
			return nil, err
		}
		accounts[accountId] = accountCopy
	}

	return accounts, nil
}

func (s *StateDB) PrepareAccountsAndAssets(accounts []int64, assets []int64) error {
	for _, accountIndex := range accounts {
		if s.dryRun {
			account := &account.Account{}
			redisAccount, err := s.redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex), account)
			if err == nil && redisAccount != nil {
				formatAccount, err := chain.ToFormatAccountInfo(account)
				if err == nil {
					s.AccountMap[accountIndex] = formatAccount
				}
			}
		}

		if s.AccountMap[accountIndex] == nil {
			accountInfo, err := s.chainDb.AccountModel.GetAccountByIndex(accountIndex)
			if err != nil {
				return err
			}
			s.AccountMap[accountIndex], err = chain.ToFormatAccountInfo(accountInfo)
			if err != nil {
				return fmt.Errorf("convert to format account info failed: %v", err)
			}
		}
		if s.AccountMap[accountIndex].AssetInfo == nil {
			s.AccountMap[accountIndex].AssetInfo = make(map[int64]*types.AccountAsset)
		}
		for _, assetId := range assets {
			if s.AccountMap[accountIndex].AssetInfo[assetId] == nil {
				s.AccountMap[accountIndex].AssetInfo[assetId] = &types.AccountAsset{
					AssetId:                  assetId,
					Balance:                  types.ZeroBigInt,
					LpAmount:                 types.ZeroBigInt,
					OfferCanceledOrFinalized: types.ZeroBigInt,
				}
			}
		}
	}

	return nil
}

func (s *StateDB) PrepareLiquidity(pairIndex int64) error {
	if s.dryRun {
		l := &liquidity.Liquidity{}
		redisLiquidity, err := s.redisCache.Get(context.Background(), dbcache.LiquidityKeyByIndex(pairIndex), l)
		if err == nil && redisLiquidity != nil {
			s.LiquidityMap[pairIndex] = l
		}
	}

	if s.LiquidityMap[pairIndex] == nil {
		liquidityInfo, err := s.chainDb.LiquidityModel.GetLiquidityByIndex(pairIndex)
		if err != nil {
			return err
		}
		s.LiquidityMap[pairIndex] = liquidityInfo
	}
	return nil
}

func (s *StateDB) PrepareNft(nftIndex int64) error {
	if s.dryRun {
		n := &nft.L2Nft{}
		redisNft, err := s.redisCache.Get(context.Background(), dbcache.NftKeyByIndex(nftIndex), n)
		if err == nil && redisNft != nil {
			s.NftMap[nftIndex] = n
		}
	}

	if s.NftMap[nftIndex] == nil {
		nftAsset, err := s.chainDb.L2NftModel.GetNft(nftIndex)
		if err != nil {
			return err
		}
		s.NftMap[nftIndex] = nftAsset
	}
	return nil
}

func (s *StateDB) UpdateAccountTree(accounts []int64, assets []int64) error {
	for _, accountIndex := range accounts {
		for _, assetId := range assets {
			assetLeaf, err := tree.ComputeAccountAssetLeafHash(
				s.AccountMap[accountIndex].AssetInfo[assetId].Balance.String(),
				s.AccountMap[accountIndex].AssetInfo[assetId].LpAmount.String(),
				s.AccountMap[accountIndex].AssetInfo[assetId].OfferCanceledOrFinalized.String(),
			)
			if err != nil {
				return fmt.Errorf("compute new account asset leaf failed: %v", err)
			}
			err = s.AccountAssetTrees[accountIndex].Set(uint64(assetId), assetLeaf)
			if err != nil {
				return fmt.Errorf("update asset tree failed: %v", err)
			}
		}

		s.AccountMap[accountIndex].AssetRoot = common.Bytes2Hex(s.AccountAssetTrees[accountIndex].Root())
		nAccountLeafHash, err := tree.ComputeAccountLeafHash(
			s.AccountMap[accountIndex].AccountNameHash,
			s.AccountMap[accountIndex].PublicKey,
			s.AccountMap[accountIndex].Nonce,
			s.AccountMap[accountIndex].CollectionNonce,
			s.AccountAssetTrees[accountIndex].Root(),
		)
		if err != nil {
			return fmt.Errorf("unable to compute account leaf: %v", err)
		}
		err = s.AccountTree.Set(uint64(accountIndex), nAccountLeafHash)
		if err != nil {
			return fmt.Errorf("unable to update account tree: %v", err)
		}
	}

	return nil
}

func (s *StateDB) UpdateLiquidityTree(pairIndex int64) error {
	nLiquidityAssetLeaf, err := tree.ComputeLiquidityAssetLeafHash(
		s.LiquidityMap[pairIndex].AssetAId,
		s.LiquidityMap[pairIndex].AssetA,
		s.LiquidityMap[pairIndex].AssetBId,
		s.LiquidityMap[pairIndex].AssetB,
		s.LiquidityMap[pairIndex].LpAmount,
		s.LiquidityMap[pairIndex].KLast,
		s.LiquidityMap[pairIndex].FeeRate,
		s.LiquidityMap[pairIndex].TreasuryAccountIndex,
		s.LiquidityMap[pairIndex].TreasuryRate,
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

func (s *StateDB) UpdateNftTree(nftIndex int64) error {
	nftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
		s.NftMap[nftIndex].CreatorAccountIndex,
		s.NftMap[nftIndex].OwnerAccountIndex,
		s.NftMap[nftIndex].NftContentHash,
		s.NftMap[nftIndex].NftL1Address,
		s.NftMap[nftIndex].NftL1TokenId,
		s.NftMap[nftIndex].CreatorTreasuryRate,
		s.NftMap[nftIndex].CollectionId,
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

func (s *StateDB) GetStateRoot() string {
	hFunc := mimc.NewMiMC()
	hFunc.Write(s.AccountTree.Root())
	hFunc.Write(s.LiquidityTree.Root())
	hFunc.Write(s.NftTree.Root())
	return common.Bytes2Hex(hFunc.Sum(nil))
}

func (s *StateDB) GetCommittedNonce(accountIndex int64) (int64, error) {
	if acc, exist := s.AccountMap[accountIndex]; exist {
		return acc.Nonce, nil
	} else {
		return 0, fmt.Errorf("account does not exist")
	}
}

func (s *StateDB) GetPendingNonce(accountIndex int64) (int64, error) {
	nonce, err := s.chainDb.MempoolModel.GetMaxNonceByAccountIndex(accountIndex)
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
	return int64(len(s.AccountAssetTrees))
}

func (s *StateDB) GetNextNftIndex() int64 {
	if len(s.PendingNewNftIndexMap) == 0 {
		maxNftIndex, err := s.chainDb.L2NftModel.GetLatestNftIndex()
		if err != nil {
			panic("get latest nft index error: " + err.Error())
		}
		return maxNftIndex + 1
	}

	maxNftIndex := int64(-1)
	for index, status := range s.PendingNewNftIndexMap {
		if status >= StateCachePending && index > maxNftIndex {
			maxNftIndex = index
		}
	}
	return maxNftIndex + 1
}
