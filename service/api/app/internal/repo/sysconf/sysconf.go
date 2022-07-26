package sysconf

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/pkg/multcache"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/errcode"
	"github.com/zeromicro/go-zero/core/logx"
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
func (m *sysconf) GetSysconfigByName(ctx context.Context, name string) (*table.Sysconfig, error) {
	logx.Errorf("[GetSysconfigByName] name:%v", name)

	f := func() (interface{}, error) {
		var config table.Sysconfig
		dbTx := m.db.Table(m.table).Where("name = ?", name).Find(&config)
		if dbTx.Error != nil {
			return nil, errcode.ErrSqlOperation.RefineError(fmt.Sprintf("GetSysconfigByName:%v", dbTx.Error))
		} else if dbTx.RowsAffected == 0 {
			return nil, errcode.ErrDataNotExist.RefineError(name)
		}
		return &config, nil
	}
	var config table.Sysconfig
	value, err := m.cache.GetWithSet(ctx, multcache.KeyGetSysconfigByName+name, &config, 5, f)
	if err != nil {
		return &config, err
	}
	config1, ok := value.(*table.Sysconfig)
	if !ok {
		return nil, errcode.ErrTypeAsset
	}
	return config1, nil
}

/*
	Func: CreateSysconfig
	Params: config *Sysconfig
	Return: error
	Description: Insert New Sysconfig
*/

func (m *sysconf) CreateSysconfig(_ context.Context, config *table.Sysconfig) error {
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

func (m *sysconf) CreateSysconfigInBatches(_ context.Context, configs []*table.Sysconfig) (rowsAffected int64, err error) {
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
func (m *sysconf) UpdateSysconfig(_ context.Context, config *table.Sysconfig) error {
	dbTx := m.db.Table(m.table).Where("name = ?", config.Name).Select(NameColumn, ValueColumn, ValueTypeColumn, CommentColumn).
		Updates(config)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[sysconfig.UpdateSysconfig] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[sysconfig.UpdateSysconfig] %s", ErrNotFound)
		logx.Error(err)
		return ErrNotFound
	}
	return nil
}
