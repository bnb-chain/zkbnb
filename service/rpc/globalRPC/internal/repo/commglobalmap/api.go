package commglobalmap

import (
	"sync"

	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	commGlobalmapHandler "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GlobalAssetInfo struct {
	AccountIndex   int64
	AssetId        int64
	AssetType      int64
	ChainId        int64
	BaseBalanceEnc string
}

type Commglobalmap interface {
	GetLatestAccountInfo(accountIndex int64) (accountInfo *commGlobalmapHandler.AccountInfo, err error)
	GetLatestLiquidityInfoForRead(pairIndex int64) (liquidityInfo *commGlobalmapHandler.LiquidityInfo, err error)
}

var singletonValue *commglobalmap
var once sync.Once

func withRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func New(svcCtx *svc.ServiceContext) Commglobalmap {
	once.Do(func() {
		conn := sqlx.NewSqlConn("postgres", svcCtx.Config.Postgres.DataSource)
		redisConn := redis.New(svcCtx.Config.CacheRedis[0].Host, withRedis(svcCtx.Config.CacheRedis[0].Type, svcCtx.Config.CacheRedis[0].Pass))
		singletonValue = &commglobalmap{
			mempoolTxDetailModel: mempool.NewMempoolDetailModel(conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
			mempoolModel:         mempool.NewMempoolModel(conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
			AccountModel:         account.NewAccountModel(conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
			liquidityModel:       liquidity.NewLiquidityModel(conn, svcCtx.Config.CacheRedis, svcCtx.GormPointer),
			redisConnection:      redisConn,
		}
	})
	return singletonValue
}
