package liquidity

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"gorm.io/gorm"
)

type liquidity struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
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
		return entity, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return entity, ErrNotExistInSql
	}
	return entity, nil
}

func (m *liquidity) GetAllLiquidityAssets() (entity []*table.Liquidity, err error) {
	return entity, nil
}
