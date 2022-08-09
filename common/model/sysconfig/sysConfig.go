/*
 * Copyright © 2021 Zkbas Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package sysconfig

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	SysconfigModel interface {
		CreateSysconfigTable() error
		DropSysconfigTable() error
		GetSysconfigByName(name string) (info *Sysconfig, err error)
		CreateSysconfigInBatches(configs []*Sysconfig) (rowsAffected int64, err error)
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
		table:      TableName,
		DB:         db,
	}
}

func (*Sysconfig) TableName() string {
	return TableName
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
	Description: get sysconfig by config name
*/
func (m *defaultSysconfigModel) GetSysconfigByName(name string) (config *Sysconfig, err error) {
	dbTx := m.DB.Table(m.table).Where("name = ?", name).Find(&config)
	if dbTx.Error != nil {
		logx.Errorf("get sys config by name error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return config, nil
}

func (m *defaultSysconfigModel) CreateSysconfigInBatches(configs []*Sysconfig) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(configs, len(configs))
	if dbTx.Error != nil {
		logx.Errorf("create sys configs error, err: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return 0, errorcode.DbErrFailToCreateSysconfig
	}
	return dbTx.RowsAffected, nil
}
