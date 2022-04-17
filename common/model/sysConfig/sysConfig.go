package sysConfig

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheZecreySysconfigIdPrefix   = "cache:zecrey:sysConfig:id:"
	cacheZecreySysconfigNamePrefix = "cache:zecrey:sysConfig:name:"
)

type (
	SysconfigModel interface {
		CreateSysconfigTable() error
		DropSysconfigTable() error
		GetSysconfigByName(name string) (info *Sysconfig, err error)
		CreateSysconfig(config *Sysconfig) error
		CreateSysconfigInBatches(configs []*Sysconfig) (rowsAffected int64, err error)
		UpdateSysconfig(config *Sysconfig) error
	}

	defaultSysconfigModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	Sysconfig struct {
		gorm.Model
		Name      string
		Value     string
		ValueType string
		Comment   string
	}
)

func NewSysconfigModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) SysconfigModel {
	return &defaultSysconfigModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      `sys_config`,
		DB:         db,
	}
}

func (*Sysconfig) TableName() string {
	return `sys_config`
}

/*
	Func: CreateSysconfigTable
	Params:
	Return: err error
	Description: create Sysconfig table
*/
func (m *defaultSysconfigModel) CreateSysconfigTable() error {
	return m.DB.AutoMigrate(Sysconfig{})
}

/*
	Func: DropSysconfigTable
	Params:
	Return: err error
	Description: drop Sysconfig table
*/
func (m *defaultSysconfigModel) DropSysconfigTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetSysconfigByName
	Params: name string
	Return: info *Sysconfig, err error
	Description: get sysConfig by config name
*/
func (m *defaultSysconfigModel) GetSysconfigByName(name string) (config *Sysconfig, err error) {
	dbTx := m.DB.Table(m.table).Where("name = ?", name).Find(&config)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[sysConfig.GetSysconfigByName] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[sysConfig.GetSysconfigByName] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return config, nil
}

/*
	Func: GetSysconfigByName
	Params: name string
	Return: info *Sysconfig, err error
	Description: get sysConfig by config name
*/
/*
func (m *defaultSysconfigModel) GetMaxChainId() (maxChainId int64, err error) {
	var (
		config *Sysconfig
	)

	dbTx := m.DB.Table(m.table).Where("name = ?", ).Find(&config)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[sysConfig.GetSysconfigByName] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[sysConfig.GetSysconfigByName] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return config, nil
}

*/

/*
	Func: CreateSysconfig
	Params: config *Sysconfig
	Return: error
	Description: Insert New Sysconfig
*/

func (m *defaultSysconfigModel) CreateSysconfig(config *Sysconfig) error {
	dbTx := m.DB.Table(m.table).Create(config)
	if dbTx.Error != nil {
		logx.Error("[sysConfig.sysConfig] %s", dbTx.Error)
		return dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[sysConfig.sysConfig] Create Invalid Sysconfig")
		return ErrInvalidSysconfig
	}
	return nil
}

func (m *defaultSysconfigModel) CreateSysconfigInBatches(configs []*Sysconfig) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(configs, len(configs))
	if dbTx.Error != nil {
		logx.Error("[sysConfig.CreateSysconfigInBatches] %s", dbTx.Error)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		logx.Error("[sysConfig.CreateSysconfigInBatches] Create Invalid Sysconfig Batches")
		return 0, ErrInvalidSysconfig
	}
	return dbTx.RowsAffected, nil
}

/*
	Func: UpdateSysconfigByName
	Params: config *Sysconfig
	Return: err error
	Description: update sysConfig by config name
*/
func (m *defaultSysconfigModel) UpdateSysconfig(config *Sysconfig) error {
	dbTx := m.DB.Table(m.table).Where("name = ?", config.Name).Select(NameColumn, ValueColumn, ValueTypeColumn, CommentColumn).
		Updates(config)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[sysConfig.UpdateSysconfig] %s", dbTx.Error)
		logx.Error(err)
		return dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[sysConfig.UpdateSysconfig] %s", ErrNotFound)
		logx.Error(err)
		return ErrNotFound
	}
	return nil
}
