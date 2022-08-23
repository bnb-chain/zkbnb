package state

import (
	"context"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/dbcache"
	accdao "github.com/bnb-chain/zkbas/common/model/account"
	liqdao "github.com/bnb-chain/zkbas/common/model/liquidity"
	nftdao "github.com/bnb-chain/zkbas/common/model/nft"
)

//go:generate mockgen -source api.go -destination api_mock.go -package state

// Fetcher will fetch the latest states (account,nft,liquidity) from redis, which is written by committer;
// and if the required data cannot be found then database will be used.
type Fetcher interface {
	GetLatestAccount(accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
	GetLatestLiquidity(pairIndex int64) (liquidityInfo *commonAsset.LiquidityInfo, err error)
	GetLatestNft(nftIndex int64) (*commonAsset.NftInfo, error)
}

func NewFetcher(redisCache dbcache.Cache,
	accountModel accdao.AccountModel,
	liquidityModel liqdao.LiquidityModel,
	nftModel nftdao.L2NftModel) Fetcher {
	return &fetcher{
		redisCache:     redisCache,
		accountModel:   accountModel,
		liquidityModel: liquidityModel,
		nftModel:       nftModel,
	}
}

type fetcher struct {
	redisCache     dbcache.Cache
	accountModel   accdao.AccountModel
	liquidityModel liqdao.LiquidityModel
	nftModel       nftdao.L2NftModel
}

func (f *fetcher) GetLatestAccount(accountIndex int64) (*commonAsset.AccountInfo, error) {
	var fa *commonAsset.AccountInfo

	redisAccount, err := f.redisCache.Get(context.Background(), dbcache.AccountKeyByIndex(accountIndex))
	if err == nil && redisAccount != nil {
		fa = redisAccount.(*commonAsset.AccountInfo)
	} else {
		account, err := f.accountModel.GetAccountByIndex(accountIndex)
		if err != nil {
			return nil, err
		}
		fa, err = commonAsset.ToFormatAccountInfo(account)
		if err != nil {
			return nil, err
		}
	}
	return fa, nil
}

func (f *fetcher) GetLatestLiquidity(pairIndex int64) (liquidityInfo *commonAsset.LiquidityInfo, err error) {
	var l *liqdao.Liquidity

	redisLiquidity, err := f.redisCache.Get(context.Background(), dbcache.LiquidityKeyByIndex(pairIndex))
	if err == nil && redisLiquidity != "" {
		l = redisLiquidity.(*liqdao.Liquidity)
	} else {
		l, err = f.liquidityModel.GetLiquidityByPairIndex(pairIndex)
		if err != nil {
			return nil, err
		}
	}

	return commonAsset.ConstructLiquidityInfo(
		pairIndex,
		l.AssetAId,
		l.AssetA,
		l.AssetBId,
		l.AssetB,
		l.LpAmount,
		l.KLast,
		l.FeeRate,
		l.TreasuryAccountIndex,
		l.TreasuryRate,
	)
}

func (f *fetcher) GetLatestNft(nftIndex int64) (*commonAsset.NftInfo, error) {
	var n *nftdao.L2Nft

	redisNft, err := f.redisCache.Get(context.Background(), dbcache.NftKeyByIndex(nftIndex))
	if err == nil && redisNft != "" {
		n = redisNft.(*nftdao.L2Nft)
	} else {
		n, err = f.nftModel.GetNftAsset(nftIndex)
		if err != nil {
			return nil, err
		}
	}

	return commonAsset.ConstructNftInfo(nftIndex,
		n.CreatorAccountIndex,
		n.OwnerAccountIndex,
		n.NftContentHash,
		n.NftL1TokenId,
		n.NftL1Address,
		n.CreatorTreasuryRate,
		n.CollectionId), nil
}
