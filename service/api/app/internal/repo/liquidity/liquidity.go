package liquidity

import (
	"fmt"

	table "github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

var (
	cacheAccountLiquidityIdPrefix                  = "cache::accountLiquidity:id:"
	cacheAccountLiquidityPairAndAccountIndexPrefix = "cache::accountLiquidity:pairAndAccountIndex:"
)

type liquidity struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	redisConn  *redis.Redis
	cache      multcache.MultCache
}

/*
	Func: GetAccountLiquidityByPairIndex
	Params: pairIndex int64
	Return: entities []*Liquidity, err error
	Description: get account liquidity entities by account index
*/
func (m *liquidity) GetLiquidityByPairIndex(pairIndex int64) (entity *table.Liquidity, err error) {
	dbTx := m.db.Table(m.table).Where("pair_index = ?", pairIndex).Find(&entity)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[liquidity.GetLiquidityByPairIndex] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[liquidity.GetLiquidityByPairIndex] %s", ErrNotExistInSql)
		logx.Error(err)
		return nil, ErrNotExistInSql
	}
	return entity, nil
}
