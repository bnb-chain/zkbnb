package commglobalmap

import (
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	commGlobalmapHandler "github.com/bnb-chain/zkbas/common/util/globalmapHandler"
)

type commglobalmap struct {
	mempoolTxDetailModel mempool.MempoolTxDetailModel
	mempoolModel         mempool.MempoolModel
	AccountModel         account.AccountModel
	liquidityModel       liquidity.LiquidityModel
	redisConnection      *redis.Redis
	offerModel           nft.OfferModel
}

func (l *commglobalmap) GetLatestAccountInfo(accountIndex int64) (accountInfo *commGlobalmapHandler.AccountInfo, err error) {
	return commGlobalmapHandler.GetLatestAccountInfo(l.AccountModel,
		l.mempoolModel, l.redisConnection, accountIndex)
}

func (l *commglobalmap) GetLatestLiquidityInfoForRead(pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error) {
	return commGlobalmapHandler.GetLatestLiquidityInfoForRead(l.liquidityModel, l.mempoolTxDetailModel, l.redisConnection, pairIndex)
}

func (l *commglobalmap) GetLatestOfferIdForWrite(accountIndex int64) (nftIndex int64, err error) {
	redisLock, offerId, err := commGlobalmapHandler.GetLatestOfferIdForWrite(l.offerModel, l.redisConnection, accountIndex)
	if err != nil {
		return 0, err
	}
	defer redisLock.Release()
	return offerId, nil
}
