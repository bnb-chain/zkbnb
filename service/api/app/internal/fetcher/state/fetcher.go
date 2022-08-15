package state

import (
	"context"
	"errors"
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

// Fetcher will fetch the latest states (account,nft,liquidity) from redis, which is written by committer;
// and if the required data cannot be found then database will be used.
type Fetcher interface {
	GetBasicAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
	GetLatestAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error)
	GetLatestLiquidityInfo(ctx context.Context, pairIndex int64) (liquidityInfo *commonAsset.LiquidityInfo, err error)
	GetLatestOfferId(ctx context.Context, accountIndex int64) (offerId int64, err error)
	GetLatestNftInfo(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error)
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

func (m *fetcher) GetBasicAccountInfo(ctx context.Context, accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		return nil, err
	}
	accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		return nil, err
	}
	return accountInfo, nil
}

func (m *fetcher) GetLatestAccountInfo(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		return nil, err
	}
	accountInfo, err := commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil && err != errorcode.DbErrNotFound {
		return nil, err
	}
	for _, mempoolTx := range mempoolTxs {
		if mempoolTx.Nonce != commonConstant.NilNonce {
			accountInfo.Nonce = mempoolTx.Nonce
		}
		for _, mempoolTxDetail := range mempoolTx.MempoolDetails {
			if mempoolTxDetail.AccountIndex != accountIndex {
				continue
			}
			switch mempoolTxDetail.AssetType {
			case commonAsset.GeneralAssetType:
				if accountInfo.AssetInfo[mempoolTxDetail.AssetId] == nil {
					accountInfo.AssetInfo[mempoolTxDetail.AssetId] = &commonAsset.AccountAsset{
						AssetId:                  mempoolTxDetail.AssetId,
						Balance:                  util.ZeroBigInt,
						LpAmount:                 util.ZeroBigInt,
						OfferCanceledOrFinalized: util.ZeroBigInt,
					}
				}
				nBalance, err := commonAsset.ComputeNewBalance(commonAsset.GeneralAssetType,
					accountInfo.AssetInfo[mempoolTxDetail.AssetId].String(), mempoolTxDetail.BalanceDelta)
				if err != nil {
					return nil, err
				}
				accountInfo.AssetInfo[mempoolTxDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
				if err != nil {
					return nil, err
				}
			case commonAsset.CollectionNonceAssetType:
				accountInfo.CollectionNonce, err = strconv.ParseInt(mempoolTxDetail.BalanceDelta, 10, 64)
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
	accountInfo.Nonce = accountInfo.Nonce + 1
	accountInfo.CollectionNonce = accountInfo.CollectionNonce + 1
	return accountInfo, nil
}

func (m *fetcher) GetLatestLiquidityInfo(ctx context.Context, pairIndex int64) (liquidityInfo *commonAsset.LiquidityInfo, err error) {
	dbLiquidityInfo, err := m.liquidityModel.GetLiquidityByPairIndex(pairIndex)
	if err != nil {
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingLiquidityTxs()
	if err != nil && err != errorcode.DbErrNotFound {
		return nil, err
	}
	liquidityInfo, err = commonAsset.ConstructLiquidityInfo(
		pairIndex,
		dbLiquidityInfo.AssetAId,
		dbLiquidityInfo.AssetA,
		dbLiquidityInfo.AssetBId,
		dbLiquidityInfo.AssetB,
		dbLiquidityInfo.LpAmount,
		dbLiquidityInfo.KLast,
		dbLiquidityInfo.FeeRate,
		dbLiquidityInfo.TreasuryAccountIndex,
		dbLiquidityInfo.TreasuryRate,
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

func (m *fetcher) GetLatestNftInfo(ctx context.Context, nftIndex int64) (*commonAsset.NftInfo, error) {
	dbNftInfo, err := m.nftModel.GetNftAsset(nftIndex)
	if err != nil {
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingNftTxs()
	if err != nil && err != errorcode.DbErrNotFound {
		return nil, err
	}
	nftInfo := commonAsset.ConstructNftInfo(nftIndex, dbNftInfo.CreatorAccountIndex, dbNftInfo.OwnerAccountIndex, dbNftInfo.NftContentHash,
		dbNftInfo.NftL1TokenId, dbNftInfo.NftL1Address, dbNftInfo.CreatorTreasuryRate, dbNftInfo.CollectionId)
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
}

func (m *fetcher) GetLatestOfferId(ctx context.Context, accountIndex int64) (int64, error) {
	lastOfferId, err := m.offerModel.GetLatestOfferId(accountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return 0, nil
		}
		return -1, err
	}
	return lastOfferId, nil
}
