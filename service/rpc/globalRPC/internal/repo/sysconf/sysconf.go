package sysconf

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	table "github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/logx"
)

type model struct {
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
func (m *model) GetSysconfigByName(ctx context.Context, name string) (*table.Sysconfig, error) {
	f := func() (interface{}, error) {
		var config table.Sysconfig
		dbTx := m.db.Table(m.table).Where("name = ?", name).Find(&config)
		if dbTx.Error != nil {
			err := fmt.Sprintf("[sysconfig.GetSysconfigByName] %s", dbTx.Error)
			logx.Error(err)
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			err := fmt.Sprintf("[sysconfig.GetSysconfigByName] %s", ErrNotFound)
			logx.Error(err)
			return nil, ErrNotFound
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
		return nil, fmt.Errorf("[GetSysconfigByName] ErrConvertFail")
	}
	return config1, nil
}
