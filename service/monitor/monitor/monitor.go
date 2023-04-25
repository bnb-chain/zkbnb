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

package monitor

import (
	"fmt"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb/core/rpc_client"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	lru "github.com/hashicorp/golang-lru"
	"gorm.io/plugin/dbresolver"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/dao/priorityrequest"
	"github.com/bnb-chain/zkbnb/dao/proof"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/monitor/config"
	"github.com/bnb-chain/zkbnb/types"
)

var (
	priorityOperationMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "priority_operation_update",
		Help:      "Priority operation requestID metrics.",
	})

	priorityOperationHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "priority_operation_update_height",
		Help:      "Priority operation height metrics.",
	})

	priorityOperationCreateMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "priority_operation_create",
		Help:      "Priority operation create requestID metrics.",
	})

	priorityOperationHeightCreateMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "priority_operation_create_height",
		Help:      "Priority operation create height metrics.",
	})

	l1SyncedBlockHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "synced_block_insert_height",
		Help:      "Synced block insert height metrics.",
	})

	l1GenericStartHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1_generic_start_height",
		Help:      "l1_generic_start_height metrics.",
	})
	l1GenericEndHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1_generic_end_height",
		Help:      "l1_generic_end_height metrics.",
	})
	l1GenericLenHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1_generic_len_height",
		Help:      "l1_generic_len_height metrics.",
	})

	l1GovernanceStartHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1_governance_start_height",
		Help:      "l1_governance_start_height metrics.",
	})
	l1GovernanceEndHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1_governance_end_height",
		Help:      "l1_governance_end_height metrics.",
	})
	l1GovernanceLenHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1_governance_len_height",
		Help:      "l1_governance_len_height metrics.",
	})
	l1MonitorHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l1_monitor_height",
		Help:      "l1_monitor_height metrics.",
	})
)

type Monitor struct {
	Config config.Config

	zkbnbContractAddress      string
	governanceContractAddress string

	db                   *gorm.DB
	L1AddressCache       *lru.Cache
	RedisCache           dbcache.Cache
	AccountModel         account.AccountModel
	BlockModel           block.BlockModel
	TxModel              tx.TxModel
	TxPoolModel          tx.TxPoolModel
	SysConfigModel       sysconfig.SysConfigModel
	L1RollupTxModel      l1rolluptx.L1RollupTxModel
	ProofModel           proof.ProofModel
	L2AssetModel         asset.AssetModel
	PriorityRequestModel priorityrequest.PriorityRequestModel
	L1SyncedBlockModel   l1syncedblock.L1SyncedBlockModel
}

func NewMonitor(c config.Config) (monitor *Monitor, err error) {

	masterDataSource := c.Postgres.MasterDataSource
	slaveDataSource := c.Postgres.SlaveDataSource
	db, err := gorm.Open(postgres.Open(masterDataSource))
	if err != nil {
		logx.Severef("gorm connect db error, err: %s", err.Error())
		return nil, err
	}

	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDataSource)},
		Replicas: []gorm.Dialector{postgres.Open(slaveDataSource)},
	}))

	l1AddressCache, err := lru.New(c.AccountCacheSize)
	if err != nil {
		logx.Severef("init account cache failed:%v", err)
		return nil, err
	}
	redisCache := dbcache.NewRedisCache(c.CacheRedis[0].Host, c.CacheRedis[0].Pass, 15*time.Minute)
	monitor = &Monitor{
		Config:               c,
		db:                   db,
		L1AddressCache:       l1AddressCache,
		RedisCache:           redisCache,
		AccountModel:         account.NewAccountModel(db),
		PriorityRequestModel: priorityrequest.NewPriorityRequestModel(db),
		TxModel:              tx.NewTxModel(db),
		TxPoolModel:          tx.NewTxPoolModel(db),
		BlockModel:           block.NewBlockModel(db),
		L1RollupTxModel:      l1rolluptx.NewL1RollupTxModel(db),
		ProofModel:           proof.NewProofModel(db),
		L1SyncedBlockModel:   l1syncedblock.NewL1SyncedBlockModel(db),
		L2AssetModel:         asset.NewAssetModel(db),
		SysConfigModel:       sysconfig.NewSysConfigModel(db),
	}

	zkbnbAddressConfig, err := monitor.SysConfigModel.GetSysConfigByName(types.ZkBNBContract)
	if err != nil {
		logx.Severef("failed to get ZkBNB contract configuration: %v", err)
		return nil, err
	}

	governanceAddressConfig, err := monitor.SysConfigModel.GetSysConfigByName(types.GovernanceContract)
	if err != nil {
		logx.Severef("failed to get governance contract configuration, %v", err)
		return nil, err
	}

	err = rpc_client.InitRpcClients(monitor.SysConfigModel, c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("failed to create rpc client instance, %v", err)
		panic("failed to create rpc client instance, err:" + err.Error())
	}

	monitor.zkbnbContractAddress = zkbnbAddressConfig.Value
	monitor.governanceContractAddress = governanceAddressConfig.Value

	if err := prometheus.Register(priorityOperationMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}
	if err := prometheus.Register(priorityOperationHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}
	if err := prometheus.Register(priorityOperationCreateMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}
	if err := prometheus.Register(priorityOperationHeightCreateMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}
	if err := prometheus.Register(l1SyncedBlockHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}

	if err := prometheus.Register(l1GenericStartHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}

	if err := prometheus.Register(l1GenericEndHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}

	if err := prometheus.Register(l1GenericLenHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}

	if err := prometheus.Register(l1GovernanceStartHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}

	if err := prometheus.Register(l1GovernanceEndHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}

	if err := prometheus.Register(l1GovernanceLenHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}

	if err := prometheus.Register(l1MonitorHeightMetric); err != nil {
		logx.Severef("fatal error, cannot register prometheus, err: %v", err)
		return nil, err
	}

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

func (m *Monitor) getBlockRangeToSync(monitorType int, cli *rpc.ProviderClient) (int64, int64, error) {
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
	latestHeight, err := cli.GetHeight()
	if err != nil {
		l1MonitorHeightMetric.Set(float64(0))
		return 0, 0, fmt.Errorf("failed to get l1 height, err: %v", err)
	}
	l1MonitorHeightMetric.Set(float64(latestHeight))
	safeHeight := latestHeight - m.Config.ChainConfig.ConfirmBlocksCount
	safeHeight = uint64(common2.MinInt64(int64(safeHeight), handledHeight+m.Config.ChainConfig.MaxHandledBlocksCount))

	return handledHeight + 1, int64(safeHeight), nil
}

func (m *Monitor) GetProviderClient() *rpc.ProviderClient {
	return rpc_client.GetRpcClient()
}
