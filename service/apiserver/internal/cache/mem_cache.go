package cache

import (
	"context"
	"fmt"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/zeromicro/go-zero/core/logx"
	"time"

	"github.com/dgraph-io/ristretto"

	accdao "github.com/bnb-chain/zkbnb/dao/account"
	assetdao "github.com/bnb-chain/zkbnb/dao/asset"
	blockdao "github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
)

const (
	cacheDefaultExpiration = time.Hour * 1 //gocache default expiration

	AccountIndexL1AddressKeyPrefix = "in:"                    //key for cache: accountIndex -> l1Address
	AccountL1AddressKeyPrefix      = "n:"                     //key for cache: l1Address -> accountIndex
	AccountByIndexKeyPrefix        = "a:"                     //key for cache: accountIndex -> account
	AccountCountKeyPrefix          = "ac"                     //key for cache: total account count
	BlockByHeightKeyPrefix         = "h:"                     //key for cache: blockHeight -> block
	BlockByCommitmentKeyPrefix     = "c:"                     //key for cache: blockCommitment -> block
	BlockCountKeyPrefix            = "bc"                     //key for cache: total block count
	TxByHashKeyPrefix              = "h:"                     //key for cache: txHash -> tx
	TxCountKeyPrefix               = "tc"                     //key for cache: total tx count
	AssetCountKeyKeyPrefix         = "AC"                     //key for cache: total asset count
	AssetIdNameKeyPrefix           = "IN:"                    //key for cache: assetId -> assetName
	AssetIdSymbolKeyPrefix         = "IS:"                    //key for cache: assetId -> assetName
	AssetByIdKeyPrefix             = "I:"                     //key for cache: assetId -> asset
	AssetBySymbolKeyPrefix         = "S:"                     //key for cache: assetSymbol -> asset
	PriceKeyPrefix                 = "p:"                     //key for cache: symbol -> price
	SysConfigKeyPrefix             = "s:"                     //key for cache: configName -> sysconfig
	TxPendingCountKeyPrefix        = "tpc"                    //key for cache: total tx pending count
	GetCommittedBlocksCountPrefix  = "CommittedBlocksCount"   // key for cache: GetCommittedBlocksCountPrefix
	GetVerifiedBlocksCountPrefix   = "VerifiedBlocksCount"    // key for cache: GetVerifiedBlocksCountPrefix
	TxsTotalCountYesterdayPrefix   = "TxsTotalCountYesterday" // key for cache: TxsTotalCountYesterday
	TxsTotalCountTodayPrefix       = "TxsTotalCountToday"     // key for cache: TxsTotalCountToday
	AccountsCountYesterdayPrefix   = "AccountsCountYesterday" // key for cache: AccountsCountYesterday
	AccountsCountTodayPrefix       = "AccountsCountToday"     // key for cache: AccountsCountToday
)

type fallback func() (interface{}, error)

type MemCache struct {
	goCache             *ristretto.Cache
	accountModel        accdao.AccountModel
	assetModel          assetdao.AssetModel
	accountExpiration   time.Duration
	blockExpiration     time.Duration
	txExpiration        time.Duration
	assetExpiration     time.Duration
	txPendingExpiration time.Duration
	priceExpiration     time.Duration
	redisCache          dbcache.Cache
}

func MustNewMemCache(accountModel accdao.AccountModel, assetModel assetdao.AssetModel,
	accountExpiration, blockExpiration, txExpiration,
	assetExpiration, txPendingExpiration, priceExpiration int, maxCounterNum, maxKeyNum int64, redisCache dbcache.Cache) *MemCache {

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: maxCounterNum,
		MaxCost:     maxKeyNum,
		BufferItems: 64, // official recommended value

		// Called when setting cost to 0 in `Set/SetWithTTL`
		Cost: func(value interface{}) int64 {
			return 1
		},
		OnEvict: func(item *ristretto.Item) {
			//logx.Infof("OnEvict %d, %d, %v, %v", item.Key, item.Cost, item.Value, item.Expiration)
		},
	})
	if err != nil {
		logx.Severe("MemCache init failed")
		panic("MemCache init failed")
	}

	memCache := &MemCache{
		goCache:             cache,
		accountModel:        accountModel,
		assetModel:          assetModel,
		accountExpiration:   time.Duration(accountExpiration) * time.Millisecond,
		blockExpiration:     time.Duration(blockExpiration) * time.Millisecond,
		txExpiration:        time.Duration(txExpiration) * time.Millisecond,
		assetExpiration:     time.Duration(assetExpiration) * time.Millisecond,
		txPendingExpiration: time.Duration(txPendingExpiration) * time.Millisecond,
		priceExpiration:     time.Duration(priceExpiration) * time.Millisecond,
		redisCache:          redisCache,
	}
	return memCache
}

func (m *MemCache) getWithSet(key string, duration time.Duration, f fallback) (interface{}, error) {
	result, found := m.goCache.Get(key)
	if found {
		return result, nil
	}
	result, err := f()
	if err != nil {
		return nil, err
	}
	m.goCache.SetWithTTL(key, result, 0, duration)
	return result, nil
}

func (m *MemCache) getWithSetFromCache(key string, fromCache bool, duration time.Duration, f fallback) (interface{}, error) {
	if fromCache {
		result, found := m.goCache.Get(key)
		if found {
			return result, nil
		}
	}

	result, err := f()
	if err != nil {
		return nil, err
	}
	m.goCache.SetWithTTL(key, result, 0, duration)
	return result, nil
}

func (m *MemCache) setAccount(accountIndex int64, l1Address string) {
	m.goCache.SetWithTTL(fmt.Sprintf("%s%d", AccountIndexL1AddressKeyPrefix, accountIndex), l1Address, 0, cacheDefaultExpiration)
	m.goCache.SetWithTTL(fmt.Sprintf("%s%s", AccountL1AddressKeyPrefix, l1Address), accountIndex, 0, cacheDefaultExpiration)
}

func (m *MemCache) GetAccountIndexByL1Address(l1Address string) (int64, error) {
	index, found := m.goCache.Get(fmt.Sprintf("%s%s", AccountL1AddressKeyPrefix, l1Address))
	if found {
		return index.(int64), nil
	}
	account, err := m.accountModel.GetAccountByL1Address(l1Address)
	if err != nil && err != types.DbErrNotFound {
		return 0, err
	}
	if err == types.DbErrNotFound {
		var accountIndex interface{}
		var redisAccount interface{}
		redisAccount, err = m.redisCache.Get(context.Background(), dbcache.AccountKeyByL1Address(l1Address), &accountIndex)
		if err == nil && redisAccount != nil {
			m.setAccount(accountIndex.(int64), l1Address)
			return accountIndex.(int64), nil
		} else {
			return 0, types.DbErrNotFound
		}
	} else {
		m.setAccount(account.AccountIndex, account.L1Address)
		return account.AccountIndex, nil
	}
}

func (m *MemCache) GetL1AddressByIndex(accountIndex int64) (string, error) {
	name, found := m.goCache.Get(fmt.Sprintf("%s%d", AccountIndexL1AddressKeyPrefix, accountIndex))
	if found {
		return name.(string), nil
	}
	account, err := m.accountModel.GetAccountByIndex(accountIndex)
	if err != nil && err != types.DbErrNotFound {
		return "", err
	}
	if err == types.DbErrNotFound {
		var l1Address interface{}
		var redisAccount interface{}
		redisAccount, err = m.redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex), &l1Address)
		if err == nil && redisAccount != nil {
			m.setAccount(accountIndex, l1Address.(string))
			return l1Address.(string), nil
		} else {
			return "", types.DbErrNotFound
		}
	} else {
		m.setAccount(account.AccountIndex, account.L1Address)
		return account.L1Address, nil
	}
}

func (m *MemCache) GetAccountL1AddressByIndex(accountIndex int64) (string, error) {
	l1Address, found := m.goCache.Get(fmt.Sprintf("%s%d", AccountIndexL1AddressKeyPrefix, accountIndex))
	if found {
		return l1Address.(string), nil
	}
	account, err := m.accountModel.GetAccountByIndex(accountIndex)
	if err != nil {
		return "", err
	}
	m.setAccount(account.AccountIndex, account.L1Address)
	return account.L1Address, nil
}

func (m *MemCache) GetAccountWithFallback(accountIndex int64, f fallback) (*accdao.Account, error) {
	key := fmt.Sprintf("%s%d", AccountByIndexKeyPrefix, accountIndex)
	a, err := m.getWithSet(key, m.accountExpiration, f)
	if err != nil {
		return nil, err
	}

	account := a.(*accdao.Account)
	m.setAccount(account.AccountIndex, account.L1Address)
	return account, nil
}

func (m *MemCache) GetAccountTotalCountWiltFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(AccountCountKeyPrefix, m.accountExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetBlockByHeightWithFallback(blockHeight int64, f fallback) (*blockdao.Block, error) {
	key := fmt.Sprintf("%s%d", BlockByHeightKeyPrefix, blockHeight)
	b, err := m.getWithSet(key, m.blockExpiration, f)
	if err != nil {
		return nil, err
	}

	block := b.(*blockdao.Block)
	key = fmt.Sprintf("%s%s", BlockByCommitmentKeyPrefix, block.BlockCommitment)
	m.goCache.SetWithTTL(key, block, 0, m.blockExpiration)
	return block, nil
}

func (m *MemCache) GetBlockByCommitmentWithFallback(blockCommitment string, f fallback) (*blockdao.Block, error) {
	key := fmt.Sprintf("%s%s", BlockByCommitmentKeyPrefix, blockCommitment)
	b, err := m.getWithSet(key, m.blockExpiration, f)
	if err != nil {
		return nil, err
	}

	block := b.(*blockdao.Block)
	key = fmt.Sprintf("%s%d", BlockByHeightKeyPrefix, block.BlockHeight)
	m.goCache.SetWithTTL(key, block, 0, m.blockExpiration)
	return block, nil
}

func (m *MemCache) GetBlockTotalCountWithFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(BlockCountKeyPrefix, m.blockExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetTxByHashWithFallback(txHash string, f fallback) (*tx.Tx, error) {
	key := fmt.Sprintf("%s%s", TxByHashKeyPrefix, txHash)
	t, err := m.getWithSet(key, m.txExpiration, f)
	if err != nil {
		return nil, err
	}
	return t.(*tx.Tx), nil
}

func (m *MemCache) GetTxTotalCountWithFallback(fromCache bool, f fallback) (int64, error) {
	count, err := m.getWithSetFromCache(TxCountKeyPrefix, fromCache, m.txExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetTxsTotalCountYesterdayBetweenWithFallback(fromCache bool, f fallback) (int64, error) {
	count, err := m.getWithSetFromCache(TxsTotalCountYesterdayPrefix, fromCache, time.Duration(30)*time.Minute, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetTxsTotalCountTodayBetweenWithFallback(fromCache bool, f fallback) (int64, error) {
	count, err := m.getWithSetFromCache(TxsTotalCountTodayPrefix, fromCache, time.Duration(30)*time.Minute, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetDistinctAccountsCountYesterdayBetweenWithFallback(fromCache bool, f fallback) (int64, error) {
	count, err := m.getWithSetFromCache(AccountsCountYesterdayPrefix, fromCache, time.Duration(30)*time.Minute, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetDistinctAccountsCountTodayBetweenWithFallback(fromCache bool, f fallback) (int64, error) {
	count, err := m.getWithSetFromCache(AccountsCountTodayPrefix, fromCache, time.Duration(30)*time.Minute, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetVerifiedBlocksCountWithFallback(fromCache bool, f fallback) (int64, error) {
	count, err := m.getWithSetFromCache(GetVerifiedBlocksCountPrefix, fromCache, time.Duration(30)*time.Minute, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetCommittedBlocksCountWithFallback(fromCache bool, f fallback) (int64, error) {
	count, err := m.getWithSetFromCache(GetCommittedBlocksCountPrefix, fromCache, time.Duration(30)*time.Minute, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetTxPendingCountKeyPrefix(f fallback) (int64, error) {
	count, err := m.getWithSet(TxPendingCountKeyPrefix, m.txPendingExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) SetTxPendingCountKeyPrefix(f fallback) (int64, error) {
	result, err := f()
	if err != nil {
		return 0, err
	}
	m.goCache.SetWithTTL(TxPendingCountKeyPrefix, result, 0, m.txPendingExpiration)
	return result.(int64), nil
}

func (m *MemCache) GetAssetTotalCountWithFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(AssetCountKeyKeyPrefix, m.txExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetAssetByIdWithFallback(assetId int64, f fallback) (*assetdao.Asset, error) {
	key := fmt.Sprintf("%s%d", AssetByIdKeyPrefix, assetId)
	a, err := m.getWithSet(key, m.assetExpiration, f)
	if err != nil {
		return nil, err
	}

	asset := a.(*assetdao.Asset)
	key = fmt.Sprintf("%s%s", AssetBySymbolKeyPrefix, asset.AssetSymbol)
	m.goCache.SetWithTTL(key, asset, 0, m.assetExpiration)

	key = fmt.Sprintf("%s%d", AssetIdNameKeyPrefix, assetId)
	m.goCache.SetWithTTL(key, asset.AssetName, 0, cacheDefaultExpiration)
	return asset, nil
}

func (m *MemCache) GetAssetBySymbolWithFallback(assetSymbol string, f fallback) (*assetdao.Asset, error) {
	key := fmt.Sprintf("%s%s", AssetBySymbolKeyPrefix, assetSymbol)
	a, err := m.getWithSet(key, m.assetExpiration, f)
	if err != nil {
		return nil, err
	}

	asset := a.(*assetdao.Asset)
	key = fmt.Sprintf("%s%d", AssetByIdKeyPrefix, asset.AssetId)
	m.goCache.SetWithTTL(key, asset, 0, m.assetExpiration)

	key = fmt.Sprintf("%s%d", AssetIdNameKeyPrefix, asset.AssetId)
	m.goCache.SetWithTTL(key, asset.AssetName, 0, cacheDefaultExpiration)
	return asset, nil
}

func (m *MemCache) GetAssetNameById(assetId int64) (string, error) {
	keyForName := fmt.Sprintf("%s%d", AssetIdNameKeyPrefix, assetId)
	name, found := m.goCache.Get(keyForName)
	if found {
		return name.(string), nil
	}
	asset, err := m.assetModel.GetAssetById(assetId)
	if err != nil {
		return "", err
	}

	m.goCache.SetWithTTL(keyForName, asset.AssetName, 0, cacheDefaultExpiration)
	keyForSymbol := fmt.Sprintf("%s%d", AssetIdSymbolKeyPrefix, assetId)
	m.goCache.SetWithTTL(keyForSymbol, asset.AssetSymbol, 0, cacheDefaultExpiration)

	return asset.AssetName, nil
}

func (m *MemCache) GetAssetSymbolById(assetId int64) (string, error) {
	keyForSymbol := fmt.Sprintf("%s%d", AssetIdSymbolKeyPrefix, assetId)
	name, found := m.goCache.Get(keyForSymbol)
	if found {
		return name.(string), nil
	}
	asset, err := m.assetModel.GetAssetById(assetId)
	if err != nil {
		return "", err
	}

	m.goCache.SetWithTTL(keyForSymbol, asset.AssetSymbol, 0, cacheDefaultExpiration)
	keyForName := fmt.Sprintf("%s%d", AssetIdNameKeyPrefix, assetId)
	m.goCache.SetWithTTL(keyForName, asset.AssetName, 0, cacheDefaultExpiration)

	return asset.AssetSymbol, nil
}

func (m *MemCache) GetPriceWithFallback(symbol string, f fallback) (float64, error) {
	key := fmt.Sprintf("%s%s", PriceKeyPrefix, symbol)
	price, err := m.getWithSet(key, m.priceExpiration, f)
	if err != nil {
		return 0, err
	}
	return price.(float64), nil
}

func (m *MemCache) SetPrice(symbol string, price float64) {
	key := fmt.Sprintf("%s%s", PriceKeyPrefix, symbol)
	m.goCache.SetWithTTL(key, price, int64(len(key)), m.priceExpiration)
}

func (m *MemCache) GetSysConfigWithFallback(configName string, fromCache bool, f fallback) (*sysconfig.SysConfig, error) {
	key := fmt.Sprintf("%s%s", SysConfigKeyPrefix, configName)
	c, err := m.getWithSetFromCache(key, fromCache, cacheDefaultExpiration, f)
	if err != nil {
		return nil, err
	}
	return c.(*sysconfig.SysConfig), nil
}

func (m *MemCache) GetCache() *ristretto.Cache {
	return m.goCache
}
