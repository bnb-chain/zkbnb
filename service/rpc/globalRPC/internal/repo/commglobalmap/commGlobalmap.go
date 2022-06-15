package commglobalmap

import (
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"

	"github.com/zeromicro/go-zero/core/logx"
)

type commglobalmap struct {
	mempoolTxDetailModel    mempool.MempoolTxDetailModel
	mempoolModel mempool.MempoolModel
	AccountModel    account.AccountModel
	liquidityModel liquidity.LiquidityModel
	redisConnection *redis.Redis
}

func (l *commglobalmap) GetLatestAccountInfo( accountIndex int64) (accountInfo *commGlobalmapHandler.AccountInfo, err error){
	return  commGlobalmapHandler.GetLatestAccountInfo(l.AccountModel,
		l.mempoolModel, l.redisConnection, accountIndex)
}

func (l *commglobalmap) GetLatestLiquidityInfoForRead( pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error){
	res,err := commGlobalmapHandler.GetLatestLiquidityInfoForRead(l.liquidityModel,l.mempoolTxDetailModel, l.redisConnection, pairIndex)
	logx.Errorf("[GetLatestAccountInfo] err:%v", err)
	return res,err
}
