package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/util"
)

//go:generate mockgen -source api.go -destination api_mock.go -package state

//TODO: replace with committer code when merge
const (
	AccountKeyPrefix   = "cache:account_"
	LiquidityKeyPrefix = "cache:liquidity_"
	NftKeyPrefix       = "cache:nft_"
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

// Fetcher will fetch the latest states (account,nft,liquidity) from redis, which is written by committer;
// and if the required data cannot be found then database will be used.
type Fetcher interface {
	GetLatestAccount(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
	GetLatestLiquidity(ctx context.Context, pairIndex int64) (liquidityInfo *commonAsset.LiquidityInfo, err error)
	GetLatestOfferId(ctx context.Context, accountIndex int64) (offerId int64, err error)
	GetLatestNft(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error)
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

func (f *fetcher) GetLatestAccount(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	var formatAccount *commonAsset.AccountInfo

	redisAccount, err := f.redisConnection.Get(AccountKeyByIndex(accountIndex))
	if err != nil {
		account, err := f.accountModel.GetAccountByAccountIndex(accountIndex)
		if err != nil {
			return nil, err
		}
		formatAccount, err = commonAsset.ToFormatAccountInfo(account)
		if err != nil {
			return nil, err
		}
	} else {
		err = json.Unmarshal([]byte(redisAccount), &formatAccount)
		if err != nil {
			return nil, err
		}
	}

	mempoolTxs, err := f.mempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil && err != errorcode.DbErrNotFound {
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		if mempoolTx.Nonce != commonConstant.NilNonce {
			formatAccount.Nonce = mempoolTx.Nonce
		}
		for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
			if mempoolTxDetail.AccountIndex != accountIndex {
				continue
			}
			switch mempoolTxDetail.AssetType {
			case commonAsset.GeneralAssetType:
				if formatAccount.AssetInfo[mempoolTxDetail.AssetId] == nil {
					formatAccount.AssetInfo[mempoolTxDetail.AssetId] = &commonAsset.AccountAsset{
						AssetId:                  mempoolTxDetail.AssetId,
						Balance:                  util.ZeroBigInt,
						LpAmount:                 util.ZeroBigInt,
						OfferCanceledOrFinalized: util.ZeroBigInt,
					}
				}
				nBalance, err := commonAsset.ComputeNewBalance(commonAsset.GeneralAssetType,
					formatAccount.AssetInfo[mempoolTxDetail.AssetId].String(), mempoolTxDetail.BalanceDelta)
				if err != nil {
					return nil, err
				}
				formatAccount.AssetInfo[mempoolTxDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
				if err != nil {
					return nil, err
				}
			case commonAsset.CollectionNonceAssetType:
				formatAccount.CollectionNonce, err = strconv.ParseInt(mempoolTxDetail.BalanceDelta, 10, 64)
				if err != nil {
					return nil, err
				}
			case commonAsset.LiquidityAssetType:
			case commonAsset.NftAssetType:
			default:
				return nil, errors.New("invalid asset type")
			}
		}
	}
	formatAccount.Nonce = formatAccount.Nonce + 1
	formatAccount.CollectionNonce = formatAccount.CollectionNonce + 1
	return formatAccount, nil
}

func (f *fetcher) GetLatestLiquidity(ctx context.Context, pairIndex int64) (liquidityInfo *commonAsset.LiquidityInfo, err error) {
	var liquidity *liquidity.Liquidity

	redisLiquidity, err := f.redisConnection.Get(LiquidityKeyByIndex(pairIndex))
	if err != nil {
		liquidity, err = f.liquidityModel.GetLiquidityByPairIndex(pairIndex)
		if err != nil {
			return nil, err
		}
	} else {
		err = json.Unmarshal([]byte(redisLiquidity), &liquidity)
		if err != nil {
			return nil, err
		}
	}

	mempoolTxs, err := f.mempoolModel.GetPendingLiquidityTxs()
	if err != nil && err != errorcode.DbErrNotFound {
		return nil, err
	}
	liquidityInfo, err = commonAsset.ConstructLiquidityInfo(
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
	if err != nil {
		logx.Errorf("[ConstructLiquidityInfo] err: %s", err.Error())
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.LiquidityAssetType || liquidityInfo.PairIndex != txDetail.AssetId {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.LiquidityAssetType, liquidityInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				return nil, err
			}
			liquidityInfo, err = commonAsset.ParseLiquidityInfo(nBalance)
			if err != nil {
				return nil, err
			}
		}
	}
	return liquidityInfo, nil
}

func (f *fetcher) GetLatestNft(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error) {
	var nft *nft.L2Nft

	redisNft, err := f.redisConnection.Get(NftKeyByIndex(nftIndex))
	if err != nil {
		nft, err = f.nftModel.GetNftAsset(nftIndex)
		if err != nil {
			return nil, err
		}
	} else {
		err = json.Unmarshal([]byte(redisNft), &nft)
		if err != nil {
			return nil, err
		}
	}

	mempoolTxs, err := f.mempoolModel.GetPendingNftTxs()
	if err != nil && err != errorcode.DbErrNotFound {
		return nil, err
	}
	nftInfo := commonAsset.ConstructNftInfo(nftIndex, nft.CreatorAccountIndex, nft.OwnerAccountIndex, nft.NftContentHash,
		nft.NftL1TokenId, nft.NftL1Address, nft.CreatorTreasuryRate, nft.CollectionId)
	for _, mempoolTx := range mempoolTxs {
		for _, txDetail := range mempoolTx.MempoolDetails {
			if txDetail.AssetType != commonAsset.NftAssetType || txDetail.AssetId != nftInfo.NftIndex {
				continue
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.NftAssetType, nftInfo.String(), txDetail.BalanceDelta)
			if err != nil {
				return nil, err
			}
			nftInfo, err = commonAsset.ParseNftInfo(nBalance)
			if err != nil {
				return nil, err
			}
		}
	}
	return nftInfo, nil
	//TODO: from redis
}

func (f *fetcher) GetLatestOfferId(ctx context.Context, accountIndex int64) (int64, error) {
	lastOfferId, err := f.offerModel.GetLatestOfferId(accountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return 0, nil
		}
		return -1, err
	}
	return lastOfferId, nil
}
