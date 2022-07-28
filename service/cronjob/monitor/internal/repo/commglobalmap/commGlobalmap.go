package commglobalmap

import (
	"context"
	"errors"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	mempoolModel         mempool.MempoolModel
	mempoolTxDetailModel mempool.MempoolTxDetailModel
	accountModel         account.AccountModel
	liquidityModel       liquidity.LiquidityModel
	redisConnection      *redis.Redis
	offerModel           nft.OfferModel
	nftModel             nft.L2NftModel
	cache                multcache.MultCache
}

func (m *model) GetLatestAccountInfoWithCache(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	f := func() (interface{}, error) {
		accountInfo, err := m.GetLatestAccountInfo(ctx, accountIndex)
		if err != nil {
			return nil, err
		}
		account, err := commonAsset.FromFormatAccountInfo(accountInfo)
		if err != nil {
			return nil, err
		}
		return account, nil
	}
	accountInfo := &account.Account{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), accountInfo, 10, f)
	if err != nil {
		return nil, err
	}
	account, _ := value.(*account.Account)
	res, err := commonAsset.ToFormatAccountInfo(account)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *model) SetLatestAccountInfoInToCache(ctx context.Context, accountIndex int64) error {
	accountInfo, err := m.GetLatestAccountInfo(ctx, accountIndex)
	if err != nil {
		return err
	}
	account, err := commonAsset.FromFormatAccountInfo(accountInfo)
	if err != nil {
		return err
	}
	if err := m.cache.Set(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), account, 10); err != nil {
		return err
	}
	return nil
}

func (m *model) DeleteLatestAccountInfoInCache(ctx context.Context, accountIndex int64) error {
	return m.cache.Delete(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex))
}

func (m *model) GetLatestAccountInfo(ctx context.Context, accountIndex int64) (*commonAsset.AccountInfo, error) {
	oAccountInfo, err := m.accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		logx.Errorf("[GetAccountByAccountIndex]param:%v, err:%v", accountIndex, err)
		return nil, err
	}
	accountInfo, err := commonAsset.ToFormatAccountInfo(oAccountInfo)
	if err != nil {
		logx.Errorf("[ToFormatAccountInfo]param:%v, err:%v", oAccountInfo, err)
		return nil, err
	}
	mempoolTxs, err := m.mempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil && err != mempool.RepoErrNotFound {
		logx.Errorf("[GetPendingMempoolTxsByAccountIndex]param:%v, err:%v", accountIndex, err)
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
					logx.Errorf("[ComputeNewBalance] err:%v", err)
					return nil, err
				}
				accountInfo.AssetInfo[mempoolTxDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
				if err != nil {
					logx.Errorf("[ParseAccountAsset]param:%v, err:%v", nBalance, err)
					return nil, err
				}
			case commonAsset.CollectionNonceAssetType:
				accountInfo.CollectionNonce, err = strconv.ParseInt(mempoolTxDetail.BalanceDelta, 10, 64)
				if err != nil {
					logx.Errorf("[ParseInt] unable to parse int: err:%v", err)
					return nil, err
				}
			case commonAsset.LiquidityAssetType:
			case commonAsset.NftAssetType:
			default:
				logx.Errorf("invalid asset type")
				return nil, errors.New("invalid asset type")
			}
		}
	}
	accountInfo.Nonce = accountInfo.Nonce + 1
	accountInfo.CollectionNonce = accountInfo.CollectionNonce + 1
	return accountInfo, nil
}
