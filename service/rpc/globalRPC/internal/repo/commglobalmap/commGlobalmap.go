package commglobalmap

import (
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
)

type commglobalmap struct {
	mempoolTxDetailModel    mempool.MempoolTxDetailModel
	mempoolModel mempool.MempoolModel
	AccountModel    account.AccountModel
	liquidityModel liquidity.LiquidityModel
	redisConnection *redis.Redis
	offerModel nft.OfferModel
}

func (l *commglobalmap) GetLatestAccountInfo( accountIndex int64) (accountInfo *commGlobalmapHandler.AccountInfo, err error){
	return  commGlobalmapHandler.GetLatestAccountInfo(l.AccountModel,
		l.mempoolModel, l.redisConnection, accountIndex)
}

func (l *commglobalmap) GetLatestLiquidityInfoForRead( pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error){
	return  commGlobalmapHandler.GetLatestLiquidityInfoForRead(l.liquidityModel,l.mempoolModel, l.redisConnection, pairIndex)
}


func (l *commglobalmap) GetLatestOfferIdForWrite(accountIndex int64) (nftIndex int64,err error) {
	redisLock, offerId, err:=  commGlobalmapHandler.GetLatestOfferIdForWrite(l.offerModel, l.redisConnection,accountIndex)
	if err != nil {
		return 0, err
	}
	defer redisLock.Release()
	return offerId,nil
}