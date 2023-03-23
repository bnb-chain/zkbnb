/*
 * Copyright © 2021 ZkBNB Protocol
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
 */

package generateproof

import (
	"fmt"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb/dao/exodusexit"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/config"
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/plugin/dbresolver"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/types"
)

type Monitor struct {
	Config                    *config.Config
	cli                       *rpc.ProviderClient
	ZkBnbContractAddress      string
	GovernanceContractAddress string
	db                        *gorm.DB
	L1SyncedBlockModel        l1syncedblock.L1SyncedBlockModel
	ExodusExitBlockModel      exodusexit.ExodusExitBlockModel
}

func NewMonitor(c *config.Config) (*Monitor, error) {
	masterDataSource := c.Postgres.MasterDataSource
	db, err := gorm.Open(postgres.Open(masterDataSource))
	if err != nil {
		logx.Severef("gorm connect db error, err: %s", err.Error())
		return nil, err
	}

	db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{postgres.Open(masterDataSource)},
	}))

	monitor := &Monitor{
		Config:               c,
		db:                   db,
		L1SyncedBlockModel:   l1syncedblock.NewL1SyncedBlockModel(db),
		ExodusExitBlockModel: exodusexit.NewExodusExitBlockModel(db),
	}

	bscRpcCli, err := rpc.NewClient(c.ChainConfig.BscTestNetRpc)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}

	monitor.ZkBnbContractAddress = c.ChainConfig.ZkBnbContractAddress
	monitor.GovernanceContractAddress = c.ChainConfig.GovernanceContractAddress
	monitor.cli = bscRpcCli
	return monitor, nil
}

func (m *Monitor) Shutdown() {
	sqlDB, err := m.db.DB()
	if err == nil && sqlDB != nil {
		err = sqlDB.Close()
	}
	if err != nil {
		logx.Errorf("close db error: %s", err.Error())
	}
}

func (m *Monitor) getBlockRangeToSync(monitorType int) (int64, int64, error) {
	latestHandledBlock, err := m.L1SyncedBlockModel.GetLatestL1SyncedBlockByType(monitorType)
	var handledHeight int64
	if err != nil {
		if err == types.DbErrNotFound {
			handledHeight = m.Config.ChainConfig.StartL1BlockHeight
		} else {
			return 0, 0, fmt.Errorf("failed to get latest l1 monitor block, err: %v", err)
		}
	} else {
		handledHeight = latestHandledBlock.L1BlockHeight
	}

	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := m.cli.GetHeight()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get l1 height, err: %v", err)
	}
	safeHeight := latestHeight - m.Config.ChainConfig.ConfirmBlocksCount
	safeHeight = uint64(common2.MinInt64(int64(safeHeight), handledHeight+m.Config.ChainConfig.MaxHandledBlocksCount))
	return handledHeight + 1, int64(safeHeight), nil
}

func (m *Monitor) ValidateAssetAddress(assetAddr common.Address) (uint16, error) {
	instance, err := zkbnb.LoadGovernanceInstance(m.cli, m.GovernanceContractAddress)
	if err != nil {
		logx.Severe(err)
		return 0, err
	}
	return zkbnb.ValidateAssetAddress(instance, assetAddr)
}
