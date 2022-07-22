package failtx

import (
	"fmt"

	table "github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/errcode"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"gorm.io/gorm"
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
		return errcode.ErrSqlOperation.RefineError(fmt.Sprint("CreateFailTx:", dbTx.Error.Error()))
	}
	if dbTx.RowsAffected == 0 {
		return errcode.ErrInvalidFailTx
	}
	return nil
}
