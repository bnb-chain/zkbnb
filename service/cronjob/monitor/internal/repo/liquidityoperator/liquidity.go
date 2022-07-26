package liquidityoperator

import (
	"errors"

	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) CreateLiquidities(pendingNewLiquidityInfos []*liquidity.Liquidity) (err error) {
	if len(pendingNewLiquidityInfos) == 0 {
		return nil
	}
	dbTx := m.db.Table(m.table).CreateInBatches(pendingNewLiquidityInfos, len(pendingNewLiquidityInfos))
	if dbTx.Error != nil {
		logx.Errorf("[CreateInBatches] unable to create pending new liquidity infos: %s", dbTx.Error.Error())
		return dbTx.Error
	}
	if dbTx.RowsAffected != int64(len(pendingNewLiquidityInfos)) {
		logx.Errorf("[CreateMempoolAndActiveAccount] invalid new liquidity infos")
		return errors.New("[CreateMempoolAndActiveAccount] invalid new liquidity infos")
	}
	return nil
}
