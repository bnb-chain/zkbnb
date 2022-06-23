package sysconf

import (
	"fmt"

	table "github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/errcode"

	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type sysconf struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

/*
	Func: GetSysconfigByName
	Params: name string
	Return: info *Sysconfig, err error
	Description: get sysconfig by config name
*/
func (m *sysconf) GetSysconfigByName(name string) (config *table.Sysconfig, err error) {
	dbTx := m.db.Table(m.table).Where("name = ?", name).Find(&config)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return nil, errcode.ErrDataNotExist
	}
	return config, nil
}

/*
	Func: CreateSysconfig
	Params: config *Sysconfig
	Return: error
	Description: Insert New Sysconfig
*/

func (m *sysconf) CreateSysconfig(config *table.Sysconfig) error {
	dbTx := m.db.Table(m.table).Create(config)
	if dbTx.Error != nil {
		logx.Error("[sysconfig.sysconfig] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[sysconfig.sysconfig] Create Invalid Sysconfig")
		return errcode.ErrInvalidSysconfig
	}
	return nil
}

func (m *sysconf) CreateSysconfigInBatches(configs []*table.Sysconfig) (rowsAffected int64, err error) {
	dbTx := m.db.Table(m.table).CreateInBatches(configs, len(configs))
	if dbTx.Error != nil {
		logx.Error("[sysconfig.CreateSysconfigInBatches] %s", dbTx.Error)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[sysconfig.CreateSysconfigInBatches] Create Invalid Sysconfig Batches")
		return 0, errcode.ErrInvalidSysconfig
	}
	return dbTx.RowsAffected, nil
}

/*
	Func: UpdateSysconfigByName
	Params: config *Sysconfig
	Return: err error
	Description: update sysconfig by config name
*/
func (m *sysconf) UpdateSysconfig(config *table.Sysconfig) error {
	dbTx := m.db.Table(m.table).Where("name = ?", config.Name).Select("name", "value", "value_type", "comment").
		Updates(config)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[sysconfig.UpdateSysconfig] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return errcode.ErrDataNotExist
	}
	return nil
}
