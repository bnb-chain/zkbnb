package liquidity

import (
	table "github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"gorm.io/gorm"
)

type liquidityModel struct {
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
func (m *liquidityModel) GetLiquidityByPairIndex(pairIndex int64) (entity *table.Liquidity, err error) {
	dbTx := m.db.Table(m.table).Where("pair_index = ?", pairIndex).Find(&entity)
	if dbTx.Error != nil {
		return entity, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.RepoErrNotFound
	}
	return entity, nil
}

func (m *liquidityModel) GetAllLiquidityAssets() (entity []*table.Liquidity, err error) {
	dbTx := m.db.Table(m.table).Raw("SELECT * FROM liquidity").Find(&entity)
	if dbTx.Error != nil {
		return entity, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.RepoErrNotFound
	}
	return entity, nil
}
