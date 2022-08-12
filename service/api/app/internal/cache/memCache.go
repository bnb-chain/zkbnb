package cache

import (
	"fmt"
	"time"

	"github.com/bnb-chain/zkbas/common/model/sysconfig"

	"github.com/zeromicro/go-zero/core/logx"

	gocache "github.com/patrickmn/go-cache"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/asset"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/tx"
)

const cacheDefaultExpiration = time.Millisecond * 500
const cacheDefaultPurgeInterval = time.Second * 60

const AccountIndexPrefix = "i:"
const AccountNamePrefix = "n:"
const AccountPkPrefix = "k:"
const LatestAccountPrefix = "l:"
const AccountPrefix = "a:"
const AccountCountPrefix = "ac:"
const BlockHeightPrefix = "h:"
const BlockCommitmentPrefix = "c:"
const BlockCountPrefix = "bc:"
const TxPrefix = "t:"
const TxCountPrefix = "tc:"
const AssetsPrefix = "A:"
const AssetIdPrefix = "I:"
const AssetSymbolPrefix = "S:"
const PricePrefix = "p:"
const SysConfigPrefix = "s:"

type fallback func() (interface{}, error)

type MemCache struct {
	goCache      *gocache.Cache
	accountModel account.AccountModel
}

func NewMemCache(accountModel account.AccountModel) *MemCache {
	memCache := &MemCache{
		goCache:      gocache.New(cacheDefaultExpiration, cacheDefaultPurgeInterval),
		accountModel: accountModel,
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
	m.goCache.Set(key, result, duration)
	return result, nil
}

func (m *MemCache) setAccount(accountIndex int64, accountName, accountPk string) {
	m.goCache.Set(fmt.Sprintf("%s%d", AccountIndexPrefix, accountIndex), accountName, gocache.NoExpiration)
	m.goCache.Set(fmt.Sprintf("%s%s", AccountNamePrefix, accountName), accountIndex, gocache.NoExpiration)
	m.goCache.Set(fmt.Sprintf("%s%s", AccountPkPrefix, accountPk), accountIndex, gocache.NoExpiration)
}

func (m *MemCache) PreloadAccounts() {
	offset := 0
	limit := 2000
	for {
		logx.Infof("to preload accounts, offset: %d, limit: %d", offset, limit)
		accounts, err := m.accountModel.GetAccountsList(limit, int64(offset))
		if err != nil {
			logx.Errorf("fail to preload accounts, offset: %d, limit: %d, err: %s", offset, limit, err.Error())
		}
		for _, acc := range accounts {
			m.setAccount(acc.AccountIndex, acc.AccountName, acc.PublicKey)
		}
		offset += limit
	}

}

func (m *MemCache) GetAccountIndexByName(accountName string) (int64, error) {
	index, found := m.goCache.Get(fmt.Sprintf("%s%s", AccountNamePrefix, accountName))
	if found {
		return index.(int64), nil
	}
	account, err := m.accountModel.GetAccountByAccountName(accountName)
	if err != nil {
		return 0, err
	}
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account.AccountIndex, nil
}

func (m *MemCache) GetAccountIndexByPk(accountPk string) (int64, error) {
	index, found := m.goCache.Get(fmt.Sprintf("%s%s", AccountPkPrefix, accountPk))
	if found {
		return index.(int64), nil
	}
	account, err := m.accountModel.GetAccountByPk(accountPk)
	if err != nil {
		return 0, err
	}
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account.AccountIndex, nil
}

func (m *MemCache) GetAccountNameByIndex(accountIndex int64) (string, error) {
	name, found := m.goCache.Get(fmt.Sprintf("%s%d", AccountIndexPrefix, accountIndex))
	if found {
		return name.(string), nil
	}
	account, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		return "", err
	}
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account.AccountName, nil
}

func (m *MemCache) GetLatestAccountWithFallback(accountIndex int64, f fallback) (*commonAsset.AccountInfo, error) {
	key := fmt.Sprintf("%s%d", LatestAccountPrefix, accountIndex)
	a, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}
	return a.(*commonAsset.AccountInfo), nil
}

func (m *MemCache) GetAccountWithFallback(accountIndex int64, f fallback) (*account.Account, error) {
	key := fmt.Sprintf("%s%d", AccountPrefix, accountIndex)
	a, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}

	account := a.(*account.Account)
	m.setAccount(account.AccountIndex, account.AccountName, account.PublicKey)
	return account, nil
}

func (m *MemCache) GetAccountTotalCountWiltFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(AccountCountPrefix, gocache.NoExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetBlockByHeightWithFallback(blockHeight int64, f fallback) (*block.Block, error) {
	key := fmt.Sprintf("%s%d", BlockHeightPrefix, blockHeight)
	b, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}

	block := b.(*block.Block)
	key = fmt.Sprintf("%s%s", BlockCommitmentPrefix, block.BlockCommitment)
	m.goCache.Set(key, block, gocache.NoExpiration)
	return block, nil
}

func (m *MemCache) GetBlockByCommitmentWithFallback(blockCommitment string, f fallback) (*block.Block, error) {
	key := fmt.Sprintf("%s%s", BlockCommitmentPrefix, blockCommitment)
	b, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}

	block := b.(*block.Block)
	key = fmt.Sprintf("%s%d", BlockHeightPrefix, block.BlockHeight)
	m.goCache.Set(key, block, gocache.NoExpiration)
	return block, nil
}

func (m *MemCache) GetBlockTotalCountWithFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(BlockCountPrefix, gocache.NoExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetTxByHashWithFallback(txHash string, f fallback) (*tx.Tx, error) {
	key := fmt.Sprintf("%s%s", TxPrefix, txHash)
	t, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}
	return t.(*tx.Tx), nil
}

func (m *MemCache) GetTxTotalCountWithFallback(f fallback) (int64, error) {
	count, err := m.getWithSet(TxCountPrefix, gocache.NoExpiration, f)
	if err != nil {
		return 0, err
	}
	return count.(int64), nil
}

func (m *MemCache) GetAssetsWithFallback(f fallback) ([]*asset.Asset, error) {
	key := fmt.Sprintf("%s", AssetsPrefix)
	t, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}

	assets := t.([]*asset.Asset)
	for _, asset := range assets {
		keyForId := fmt.Sprintf("%s%d", AssetIdPrefix, asset.AssetId)
		m.goCache.Set(keyForId, asset, gocache.NoExpiration)
		keyForSymbol := fmt.Sprintf("%s%s", AssetSymbolPrefix, asset.AssetSymbol)
		m.goCache.Set(keyForSymbol, asset, gocache.NoExpiration)
	}
	return assets, nil
}

func (m *MemCache) GetAssetByIdWithFallback(assetId int64, f fallback) (*asset.Asset, error) {
	key := fmt.Sprintf("%s%d", AssetIdPrefix, assetId)
	a, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}

	asset := a.(*asset.Asset)
	key = fmt.Sprintf("%s%s", AssetSymbolPrefix, asset.AssetSymbol)
	m.goCache.Set(key, asset, gocache.NoExpiration)
	return asset, nil
}

func (m *MemCache) GetAssetBySymbolWithFallback(assetSymbol string, f fallback) (*asset.Asset, error) {
	key := fmt.Sprintf("%s%s", AssetSymbolPrefix, assetSymbol)
	a, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}

	asset := a.(*asset.Asset)
	key = fmt.Sprintf("%s%d", AssetIdPrefix, asset.AssetId)
	m.goCache.Set(key, asset, gocache.NoExpiration)
	return asset, nil
}

func (m *MemCache) GetPriceWithFallback(symbol string, f fallback) (float64, error) {
	key := fmt.Sprintf("%s%s", PricePrefix, symbol)
	price, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return 0, err
	}
	return price.(float64), nil
}

func (m *MemCache) GetSysConfigWithFallback(configName string, f fallback) (*sysconfig.SysConfig, error) {
	key := fmt.Sprintf("%s%s", SysConfigPrefix, configName)
	c, err := m.getWithSet(key, gocache.NoExpiration, f)
	if err != nil {
		return nil, err
	}
	return c.(*sysconfig.SysConfig), nil
}
