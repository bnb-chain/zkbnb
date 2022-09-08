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
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb-eth-rpc/_rpc"
	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/blockwitness"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/mempool"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/priorityrequest"
	"github.com/bnb-chain/zkbnb/dao/proof"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/monitor/config"
	"github.com/bnb-chain/zkbnb/types"
)

type Monitor struct {
	Config config.Config

	cli *_rpc.ProviderClient

	zkbnbContractAddress      string
	governanceContractAddress string

	db                    *gorm.DB
	BlockModel            block.BlockModel
	TxModel               tx.TxModel
	TxDetailModel         tx.TxDetailModel
	CompressedBlockModel  compressedblock.CompressedBlockModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	L2NftModel            nft.L2NftModel
	L2NftHistoryModel     nft.L2NftHistoryModel
	MempoolModel          mempool.MempoolModel
	SysConfigModel        sysconfig.SysConfigModel
	L1RollupTxModel       l1rolluptx.L1RollupTxModel
	L2AssetModel          asset.AssetModel
	PriorityRequestModel  priorityrequest.PriorityRequestModel
	L1SyncedBlockModel    l1syncedblock.L1SyncedBlockModel
	ProofModel            proof.ProofModel
	BlockWitnessModel     blockwitness.BlockWitnessModel
}

func NewMonitor(c config.Config) *Monitor {
	db, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err: %s", err.Error())
	}
	monitor := &Monitor{
		Config:                c,
		PriorityRequestModel:  priorityrequest.NewPriorityRequestModel(db),
		db:                    db,
		MempoolModel:          mempool.NewMempoolModel(db),
		BlockModel:            block.NewBlockModel(db),
		TxModel:               tx.NewTxModel(db),
		TxDetailModel:         tx.NewTxDetailModel(db),
		CompressedBlockModel:  compressedblock.NewCompressedBlockModel(db),
		AccountModel:          account.NewAccountModel(db),
		AccountHistoryModel:   account.NewAccountHistoryModel(db),
		LiquidityModel:        liquidity.NewLiquidityModel(db),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(db),
		L2NftModel:            nft.NewL2NftModel(db),
		L2NftHistoryModel:     nft.NewL2NftHistoryModel(db),
		L1RollupTxModel:       l1rolluptx.NewL1RollupTxModel(db),
		L1SyncedBlockModel:    l1syncedblock.NewL1SyncedBlockModel(db),
		L2AssetModel:          asset.NewAssetModel(db),
		SysConfigModel:        sysconfig.NewSysConfigModel(db),
		ProofModel:            proof.NewProofModel(db),
		BlockWitnessModel:     blockwitness.NewBlockWitnessModel(db),
	}

	zkbnbAddressConfig, err := monitor.SysConfigModel.GetSysConfigByName(types.ZkBNBContract)
	if err != nil {
		logx.Errorf("GetSysConfigByName err: %s", err.Error())
		panic(err)
	}

	governanceAddressConfig, err := monitor.SysConfigModel.GetSysConfigByName(types.GovernanceContract)
	if err != nil {
		logx.Severef("fatal error, cannot fetch governance contract from sysconfig, err: %s, SysConfigName: %s",
			err.Error(), types.GovernanceContract)
		panic(err)
	}

	networkRpc, err := monitor.SysConfigModel.GetSysConfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("fatal error, cannot fetch NetworkRPC from sysconfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	logx.Infof("ChainName: %s, zkbnbContractAddress: %s, networkRpc: %s",
		c.ChainConfig.NetworkRPCSysConfigName, zkbnbAddressConfig.Value, networkRpc.Value)

	bscRpcCli, err := _rpc.NewClient(networkRpc.Value)
	if err != nil {
		panic(err)
	}

	monitor.zkbnbContractAddress = zkbnbAddressConfig.Value
	monitor.governanceContractAddress = governanceAddressConfig.Value
	monitor.cli = bscRpcCli

	return monitor
}
