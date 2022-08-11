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
 */

package monitor

import (
	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/asset"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/l1RollupTx"
	"github.com/bnb-chain/zkbas/common/model/l1SyncedBlock"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/priorityRequest"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/sysConfigName"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/config"
)

type Monitor struct {
	Config config.Config

	cli *_rpc.ProviderClient

	zkbasContractAddress      string
	governanceContractAddress string

	BlockModel           block.BlockModel
	MempoolModel         mempool.MemPoolModel
	SysConfigModel       sysconfig.SysConfigModel
	L1RollupTxModel      l1RollupTx.L1RollupTxModel
	L2AssetModel         asset.AssetModel
	PriorityRequestModel priorityRequest.PriorityRequestModel
	L1SyncedBlockModel   l1SyncedBlock.L1SyncedBlockModel
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func NewMonitor(c config.Config) *Monitor {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err: %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	monitor := &Monitor{
		Config:               c,
		PriorityRequestModel: priorityRequest.NewPriorityRequestModel(conn, c.CacheRedis, gormPointer),
		MempoolModel:         mempool.NewMempoolModel(conn, c.CacheRedis, gormPointer),
		BlockModel:           block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		L1RollupTxModel:      l1RollupTx.NewL1RollupTxModel(conn, c.CacheRedis, gormPointer),
		L1SyncedBlockModel:   l1SyncedBlock.NewL1SyncedBlockModel(conn, c.CacheRedis, gormPointer),
		L2AssetModel:         asset.NewAssetModel(conn, c.CacheRedis, gormPointer),
		SysConfigModel:       sysconfig.NewSysConfigModel(conn, c.CacheRedis, gormPointer),
	}

	zkbasAddressConfig, err := monitor.SysConfigModel.GetSysConfigByName(sysConfigName.ZkbasContract)
	if err != nil {
		logx.Errorf("GetSysConfigByName err: %s", err.Error())
		panic(err)
	}

	governanceAddressConfig, err := monitor.SysConfigModel.GetSysConfigByName(sysConfigName.GovernanceContract)
	if err != nil {
		logx.Severef("fatal error, cannot fetch governance contract from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), sysConfigName.GovernanceContract)
		panic(err)
	}

	networkRpc, err := monitor.SysConfigModel.GetSysConfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("fatal error, cannot fetch NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	logx.Infof("ChainName: %s, zkbasContractAddress: %s, networkRpc: %s",
		c.ChainConfig.NetworkRPCSysConfigName, zkbasAddressConfig.Value, networkRpc.Value)

	bscRpcCli, err := _rpc.NewClient(networkRpc.Value)
	if err != nil {
		panic(err)
	}

	monitor.zkbasContractAddress = zkbasAddressConfig.Value
	monitor.governanceContractAddress = governanceAddressConfig.Value
	monitor.cli = bscRpcCli

	return monitor
}
