package commglobalmap

import (
	"context"
	"strconv"

	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/errcode"
)

type model struct {
	mempoolModel         mempool.MempoolModel
	mempoolTxDetailModel mempool.MempoolTxDetailModel
	accountModel         account.AccountModel
	liquidityModel       liquidity.LiquidityModel
	redisConnection      *redis.Redis
	offerModel           nft.OfferModel
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
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), accountInfo, 1, f)
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
	txDetails, err := m.mempoolTxDetailModel.GetMempoolTxDetailsByAccountIndex(accountIndex)
	if err != nil && err != mempool.ErrNotFound {
		logx.Errorf("[GetPendingMempoolTxsByAccountIndex]param:%v, err:%v", accountIndex, err)
		return nil, err
	}
	txMap := make(map[int64]bool)
	for _, txDetail := range txDetails {
		if !txMap[txDetail.TxId] {
			mempoolTx, err := m.mempoolModel.GetMempoolTxByTxId(uint(txDetail.TxId))
			if err != nil {
				logx.Errorf("[GetLatestAccountInfo] unable to get mempool tx by tx id: %s", err.Error())
				return nil, err
			}
			txMap[txDetail.TxId] = true
			if mempoolTx.Status != mempool.PendingTxStatus {
				continue
			}
			if mempoolTx.Nonce != commonConstant.NilNonce {
				accountInfo.Nonce = mempoolTx.Nonce
			}
		}
		switch txDetail.AssetType {
		case commonAsset.GeneralAssetType:
			if accountInfo.AssetInfo[txDetail.AssetId] == nil {
				accountInfo.AssetInfo[txDetail.AssetId] = &commonAsset.AccountAsset{
					AssetId:                  txDetail.AssetId,
					Balance:                  util.ZeroBigInt,
					LpAmount:                 util.ZeroBigInt,
					OfferCanceledOrFinalized: util.ZeroBigInt,
				}
			}
			nBalance, err := commonAsset.ComputeNewBalance(commonAsset.GeneralAssetType,
				accountInfo.AssetInfo[txDetail.AssetId].String(), txDetail.BalanceDelta)
			if err != nil {
				logx.Errorf("[ComputeNewBalance] err:%v", err)
				return nil, err
			}
			accountInfo.AssetInfo[txDetail.AssetId], err = commonAsset.ParseAccountAsset(nBalance)
			if err != nil {
				logx.Errorf("[ParseAccountAsset]param:%v, err:%v", nBalance, err)
				return nil, err
			}
		case commonAsset.CollectionNonceAssetType:
			accountInfo.CollectionNonce, err = strconv.ParseInt(txDetail.BalanceDelta, 10, 64)
			if err != nil {
				logx.Errorf("[ParseInt] unable to parse int: err:%v", err)
				return nil, err
			}
		case commonAsset.LiquidityAssetType:
		case commonAsset.NftAssetType:
		default:
			logx.Errorf("invalid asset type")
			return nil, errcode.ErrInvalidAssetType
		}
	}
	accountInfo.Nonce = accountInfo.Nonce + 1
	accountInfo.CollectionNonce = accountInfo.CollectionNonce + 1
	return accountInfo, nil
}

func (l *model) GetLatestLiquidityInfoForRead(pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error) {
	return commGlobalmapHandler.GetLatestLiquidityInfoForRead(l.liquidityModel, l.mempoolModel, l.redisConnection, pairIndex)
}

func (l *model) GetLatestOfferIdForWrite(accountIndex int64) (nftIndex int64, err error) {
	redisLock, offerId, err := commGlobalmapHandler.GetLatestOfferIdForWrite(l.offerModel, l.redisConnection, accountIndex)
	if err != nil {
		return 0, err
	}
	defer redisLock.Release()
	return offerId, nil
}
