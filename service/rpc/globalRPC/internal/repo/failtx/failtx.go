package failtx

import (
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

/*
	Func: CreateFailTx
	Params: failTx *FailTx
	Return: err error
	Description: create fail txVerification
*/
func (m *model) CreateFailTx(failTx *table.FailTx) error {
	dbTx := m.db.Table(m.table).Create(failTx)
	if dbTx.Error != nil {
		logx.Errorf("fail to create failTx, error: %s", dbTx.Error.Error())
		return errorcode.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return errorcode.DbErrFailToCreateFailTx
	}
	return nil
}
