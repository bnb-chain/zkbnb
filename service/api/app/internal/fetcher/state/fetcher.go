package state

import (
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
)

//go:generate mockgen -source api.go -destination api_mock.go -package state

//TODO: replace with committer code when merge
const (
	AccountKeyPrefix   = "cache:account_"
	LiquidityKeyPrefix = "cache:liquidity_"
	NftKeyPrefix       = "cache:nft_"
	OfferIdKeyPrefix   = "cache:offerId_"
)

func AccountKeyByIndex(accountIndex int64) string {
	return AccountKeyPrefix + fmt.Sprintf("%d", accountIndex)
}

func LiquidityKeyByIndex(pairIndex int64) string {
	return LiquidityKeyPrefix + fmt.Sprintf("%d", pairIndex)
}

func NftKeyByIndex(nftIndex int64) string {
	return NftKeyPrefix + fmt.Sprintf("%d", nftIndex)
}

func OfferIdKeyByIndex(nftIndex int64) string {
	return OfferIdKeyPrefix + fmt.Sprintf("%d", nftIndex)
}

// Fetcher will fetch the latest states (account,nft,liquidity) from redis, which is written by committer;
// and if the required data cannot be found then database will be used.
type Fetcher interface {
	GetLatestAccount(accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
	GetLatestLiquidity(pairIndex int64) (liquidityInfo *commonAsset.LiquidityInfo, err error)
	GetLatestOfferId(accountIndex int64) (offerId int64, err error)
	GetLatestNft(nftIndex int64) (*commonAsset.NftInfo, error)
	GetPendingNonce(accountIndex int64) (nonce int64, err error)
}

func NewFetcher(redisConn *redis.Redis,
	mempoolModel mempool.MempoolModel,
	mempoolDetailModel mempool.MempoolTxDetailModel,
	accountModel account.AccountModel,
	liquidityModel liquidity.LiquidityModel,
	nftModel nft.L2NftModel,
	offerModel nft.OfferModel) Fetcher {
	return &fetcher{
		redisConnection:      redisConn,
		mempoolModel:         mempoolModel,
		mempoolTxDetailModel: mempoolDetailModel,
		accountModel:         accountModel,
		liquidityModel:       liquidityModel,
		nftModel:             nftModel,
		offerModel:           offerModel,
	}
}

type fetcher struct {
	redisConnection      *redis.Redis
	mempoolModel         mempool.MempoolModel
	mempoolTxDetailModel mempool.MempoolTxDetailModel
	accountModel         account.AccountModel
	liquidityModel       liquidity.LiquidityModel
	offerModel           nft.OfferModel
	nftModel             nft.L2NftModel
}

func (f *fetcher) GetLatestAccount(accountIndex int64) (*commonAsset.AccountInfo, error) {
	var formatAccount *commonAsset.AccountInfo

	redisAccount, err := f.redisConnection.Get(AccountKeyByIndex(accountIndex))
	if err == nil && redisAccount != "" {
		err = json.Unmarshal([]byte(redisAccount), &formatAccount)
		if err != nil {
			return nil, err
		}
	} else {
		account, err := f.accountModel.GetAccountByAccountIndex(accountIndex)
		if err != nil {
			return nil, err
		}
		formatAccount, err = commonAsset.ToFormatAccountInfo(account)
		if err != nil {
			return nil, err
		}
	}
	return formatAccount, nil
}

func (f *fetcher) GetLatestLiquidity(pairIndex int64) (liquidityInfo *commonAsset.LiquidityInfo, err error) {
	var liquidity *liquidity.Liquidity

	redisLiquidity, err := f.redisConnection.Get(LiquidityKeyByIndex(pairIndex))
	if err == nil && redisLiquidity != "" {
		err = json.Unmarshal([]byte(redisLiquidity), &liquidity)
		if err != nil {
			return nil, err
		}
	} else {
		liquidity, err = f.liquidityModel.GetLiquidityByPairIndex(pairIndex)
		if err != nil {
			return nil, err
		}
	}

	return commonAsset.ConstructLiquidityInfo(
		pairIndex,
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
}

func (f *fetcher) GetLatestNft(nftIndex int64) (*commonAsset.NftInfo, error) {
	var nft *nft.L2Nft

	redisNft, err := f.redisConnection.Get(NftKeyByIndex(nftIndex))
	if err == nil && redisNft != "" {
		err = json.Unmarshal([]byte(redisNft), &nft)
		if err != nil {
			return nil, err
		}
	} else {
		nft, err = f.nftModel.GetNftAsset(nftIndex)
		if err != nil {
			return nil, err
		}
	}

	return commonAsset.ConstructNftInfo(nftIndex,
		nft.CreatorAccountIndex,
		nft.OwnerAccountIndex,
		nft.NftContentHash,
		nft.NftL1TokenId,
		nft.NftL1Address,
		nft.CreatorTreasuryRate,
		nft.CollectionId), nil
}

func (f *fetcher) GetLatestOfferId(accountIndex int64) (int64, error) {
	lastOfferId := int64(0)

	redisOfferId, err := f.redisConnection.Get(OfferIdKeyByIndex(accountIndex))
	if err == nil && redisOfferId != "" {
		err = json.Unmarshal([]byte(redisOfferId), &lastOfferId)
		if err != nil {
			return -1, err
		}
	} else {
		lastOfferId, err = f.offerModel.GetLatestOfferId(accountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return 0, nil
			}
			return -1, err
		}
	}

	return lastOfferId, nil
}

func (f *fetcher) GetPendingNonce(accountIndex int64) (nonce int64, err error) {
	nonce, err = f.mempoolModel.GetMaxNonceByAccountIndex(accountIndex)
	if err == nil {
		return nonce + 1, nil
	}
	redisAccount, err := f.redisConnection.Get(AccountKeyByIndex(accountIndex))
	if err == nil {
		var formatAccount *commonAsset.AccountInfo
		err = json.Unmarshal([]byte(redisAccount), &formatAccount)
		if err != nil {
			return 0, err
		}
		return formatAccount.Nonce, nil
	}
	dbAccount, err := f.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		return dbAccount.Nonce + 1, nil
	}
	return 0, err
}
