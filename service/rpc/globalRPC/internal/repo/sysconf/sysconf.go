package sysconf

import (
	"context"
	"fmt"
	"github.com/bnb-chain/zkbas/errorcode"
	"gorm.io/gorm"

	table "github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/pkg/multcache"
)

type model struct {
	table string
	db    *gorm.DB
	cache multcache.MultCache
}

func (m *model) GetSysconfigByName(ctx context.Context, name string) (*table.Sysconfig, error) {
	f := func() (interface{}, error) {
		var config table.Sysconfig
		dbTx := m.db.Table(m.table).Where("name = ?", name).Find(&config)
		if dbTx.Error != nil {
			return nil, errorcode.RepoErrSqlOperation.RefineError(fmt.Sprintf("GetSysconfigByName:%v", dbTx.Error))
		} else if dbTx.RowsAffected == 0 {
			return nil, errorcode.RepoErrNotFound
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
