/*
 * Copyright © 2021 ZkBAS Protocol
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
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/types"
)

const (
	TableName = `sys_config`
)

type (
	SysConfigModel interface {
		CreateSysConfigTable() error
		DropSysConfigTable() error
		GetSysConfigByName(name string) (info *SysConfig, err error)
		CreateSysConfigInBatches(configs []*SysConfig) (rowsAffected int64, err error)
	}

	defaultSysConfigModel struct {
		table string
		DB    *gorm.DB
	}

	SysConfig struct {
		gorm.Model
		Name      string
		Value     string
		ValueType string
		Comment   string
	}
)

func NewSysConfigModel(db *gorm.DB) SysConfigModel {
	return &defaultSysConfigModel{
		table: TableName,
		DB:    db,
	}
}

func (*SysConfig) TableName() string {
	return TableName
}

func (m *defaultSysConfigModel) CreateSysConfigTable() error {
	return m.DB.AutoMigrate(SysConfig{})
}

func (m *defaultSysConfigModel) DropSysConfigTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultSysConfigModel) GetSysConfigByName(name string) (config *SysConfig, err error) {
	dbTx := m.DB.Table(m.table).Where("name = ?", name).Find(&config)
	if dbTx.Error != nil {
		return nil, types.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, types.DbErrNotFound
	}
	return config, nil
}

func (m *defaultSysConfigModel) CreateSysConfigInBatches(configs []*SysConfig) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(configs, len(configs))
	if dbTx.Error != nil {
		return 0, types.DbErrSqlOperation
	}
	if dbTx.RowsAffected == 0 {
		return 0, types.DbErrFailToCreateSysconfig
	}
	return dbTx.RowsAffected, nil
}
