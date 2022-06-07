package commglobalmap

import (
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
)

type commglobalmap struct {
	MempoolModel    mempool.MempoolModel
	AccountModel    account.AccountModel
	liquidityModel liquidity.LiquidityModel,
	RedisConnection *redis.Redis
}

func (l *commglobalmap) GetLatestAccountInfo( accountIndex int64) (accountInfo *commGlobalmapHandler.AccountInfo, err error){
	return  commGlobalmapHandler.GetLatestAccountInfo(l.AccountModel,
		l.MempoolModel, l.RedisConnection, accountIndex)
}

func (l *commglobalmap) GetLatestLiquidityInfoForRead( pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error){
	return  commGlobalmapHandler.GetLatestLiquidityInfoForRead(l.AccountModel,
		l.MempoolModel, l.RedisConnection, pairIndex)
}


