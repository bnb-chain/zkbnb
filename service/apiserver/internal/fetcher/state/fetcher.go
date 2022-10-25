package state

import (
	"context"

	"github.com/bnb-chain/zkbnb/common/chain"
	accdao "github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	nftdao "github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/types"
)

//go:generate mockgen -source api.go -destination api_mock.go -package state

// Fetcher will fetch the latest states (account,nft) from redis, which is written by committer;
// and if the required data cannot be found then database will be used.
type Fetcher interface {
	GetLatestAccount(accountIndex int64) (accountInfo *types.AccountInfo, err error)
	GetLatestNft(nftIndex int64) (*types.NftInfo, error)
}

func NewFetcher(redisCache dbcache.Cache,
	accountModel accdao.AccountModel,
	nftModel nftdao.L2NftModel) Fetcher {
	return &fetcher{
		redisCache:   redisCache,
		accountModel: accountModel,
		nftModel:     nftModel,
	}
}

type fetcher struct {
	redisCache   dbcache.Cache
	accountModel accdao.AccountModel
	nftModel     nftdao.L2NftModel
}

func (f *fetcher) GetLatestAccount(accountIndex int64) (*types.AccountInfo, error) {
	var fa *types.AccountInfo
	account := &accdao.Account{}

	redisAccount, err := f.redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex), account)
	if err == nil && redisAccount != nil {
		fa, err = chain.ToFormatAccountInfo(account)
		if err == nil {
			return fa, nil
		}
	} else {
		dbAccount, err := f.accountModel.GetAccountByIndex(accountIndex)
		if err != nil {
			return nil, err
		}
		fa, err = chain.ToFormatAccountInfo(dbAccount)
		if err != nil {
			return nil, err
		}
	}
	return fa, nil
}

func (f *fetcher) GetLatestNft(nftIndex int64) (*types.NftInfo, error) {
	n := &nftdao.L2Nft{}

	redisNft, err := f.redisCache.Get(context.Background(), dbcache.NftKeyByIndex(nftIndex), n)
	if err == nil && redisNft != "" {
	} else {
		n, err = f.nftModel.GetNft(nftIndex)
		if err != nil {
			return nil, err
		}
	}

	return types.ConstructNftInfo(nftIndex,
		n.CreatorAccountIndex,
		n.OwnerAccountIndex,
		n.NftContentHash,
		n.CreatorTreasuryRate,
		n.CollectionId), nil
}
