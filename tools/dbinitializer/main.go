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

package main

import (
	"encoding/json"
	"flag"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/asset"
	"github.com/bnb-chain/zkbas/dao/block"
	"github.com/bnb-chain/zkbas/dao/blockwitness"
	"github.com/bnb-chain/zkbas/dao/compressedblock"
	"github.com/bnb-chain/zkbas/dao/l1rolluptx"
	"github.com/bnb-chain/zkbas/dao/l1syncedblock"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/mempool"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/priorityrequest"
	"github.com/bnb-chain/zkbas/dao/proof"
	"github.com/bnb-chain/zkbas/dao/sysconfig"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/tree"
	"github.com/bnb-chain/zkbas/types"
)

var (
	dsn   = "host=localhost user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable"
	DB, _ = gorm.Open(postgres.Open(dsn), &gorm.Config{})
)

var configFile = flag.String("f", "./contractaddr.yaml", "the config file")
var svrConf config

const (
	BSCTestNetworkRPC   = "http://tf-dex-preview-validator-nlb-6fd109ac8b9d390a.elb.ap-northeast-1.amazonaws.com:8545"
	LocalTestNetworkRPC = "http://127.0.0.1:8545/"
)

func main() {
	conf.MustLoad(*configFile, &svrConf)

	unmarshal, _ := json.Marshal(svrConf)
	logx.Infof("init configs: %s", string(unmarshal))

	dropTables()
	initTable()
}

func initSysConfig() []*sysconfig.SysConfig {
	return []*sysconfig.SysConfig{
		{
			Name:      types.SysGasFee,
			Value:     "100000000000000",
			ValueType: "string",
			Comment:   "based on BNB",
		},
		{
			Name:      types.TreasuryAccountIndex,
			Value:     "0",
			ValueType: "int",
			Comment:   "treasury index",
		},
		{
			Name:      types.GasAccountIndex,
			Value:     "1",
			ValueType: "int",
			Comment:   "gas index",
		},
		{
			Name:      types.ZkbasContract,
			Value:     svrConf.ZkbasProxy,
			ValueType: "string",
			Comment:   "Zkbas contract on BSC",
		},
		// Governance Contract
		{
			Name:      types.GovernanceContract,
			Value:     svrConf.Governance,
			ValueType: "string",
			Comment:   "Governance contract on BSC",
		},
		// network rpc
		{
			Name:      types.BscTestNetworkRpc,
			Value:     BSCTestNetworkRPC,
			ValueType: "string",
			Comment:   "BSC network rpc",
		},
		{
			Name:      types.LocalTestNetworkRpc,
			Value:     LocalTestNetworkRPC,
			ValueType: "string",
			Comment:   "Local network rpc",
		},
		{
			Name:      types.ZnsPriceOracle,
			Value:     svrConf.ZnsPriceOracle,
			ValueType: "string",
			Comment:   "Zns Price Oracle",
		},
	}
}

func initAssetsInfo() []*asset.Asset {
	return []*asset.Asset{
		{
			AssetId:     0,
			L1Address:   "0x00",
			AssetName:   "BNB",
			AssetSymbol: "BNB",
			Decimals:    18,
			Status:      0,
			IsGasAsset:  asset.IsGasAsset,
		},
	}
}

var (
	sysConfigModel          = sysconfig.NewSysConfigModel(DB)
	accountModel            = account.NewAccountModel(DB)
	accountHistoryModel     = account.NewAccountHistoryModel(DB)
	assetModel              = asset.NewAssetModel(DB)
	mempoolDetailModel      = mempool.NewMempoolDetailModel(DB)
	mempoolModel            = mempool.NewMempoolModel(DB)
	failTxModel             = tx.NewFailTxModel(DB)
	txDetailModel           = tx.NewTxDetailModel(DB)
	txModel                 = tx.NewTxModel(DB)
	blockModel              = block.NewBlockModel(DB)
	compressedBlockModel    = compressedblock.NewCompressedBlockModel(DB)
	blockWitnessModel       = blockwitness.NewBlockWitnessModel(DB)
	proofModel              = proof.NewProofModel(DB)
	l1SyncedBlockModel      = l1syncedblock.NewL1SyncedBlockModel(DB)
	priorityRequestModel    = priorityrequest.NewPriorityRequestModel(DB)
	l1RollupTModel          = l1rolluptx.NewL1RollupTxModel(DB)
	liquidityModel          = liquidity.NewLiquidityModel(DB)
	liquidityHistoryModel   = liquidity.NewLiquidityHistoryModel(DB)
	nftModel                = nft.NewL2NftModel(DB)
	nftHistoryModel         = nft.NewL2NftHistoryModel(DB)
	nftCollectionModel      = nft.NewL2NftCollectionModel(DB)
	nftWithdrawHistoryModel = nft.NewL2NftWithdrawHistoryModel(DB)
)

func dropTables() {
	assert.Nil(nil, sysConfigModel.DropSysConfigTable())
	assert.Nil(nil, accountModel.DropAccountTable())
	assert.Nil(nil, accountHistoryModel.DropAccountHistoryTable())
	assert.Nil(nil, assetModel.DropAssetTable())
	assert.Nil(nil, mempoolDetailModel.DropMempoolDetailTable())
	assert.Nil(nil, mempoolModel.DropMempoolTxTable())
	assert.Nil(nil, failTxModel.DropFailTxTable())
	assert.Nil(nil, txDetailModel.DropTxDetailTable())
	assert.Nil(nil, txModel.DropTxTable())
	assert.Nil(nil, blockModel.DropBlockTable())
	assert.Nil(nil, compressedBlockModel.DropCompressedBlockTable())
	assert.Nil(nil, blockWitnessModel.DropBlockWitnessTable())
	assert.Nil(nil, proofModel.DropProofTable())
	assert.Nil(nil, l1SyncedBlockModel.DropL1SyncedBlockTable())
	assert.Nil(nil, priorityRequestModel.DropPriorityRequestTable())
	assert.Nil(nil, l1RollupTModel.DropL1RollupTxTable())
	assert.Nil(nil, liquidityModel.DropLiquidityTable())
	assert.Nil(nil, liquidityHistoryModel.DropLiquidityHistoryTable())
	assert.Nil(nil, nftModel.DropL2NftTable())
	assert.Nil(nil, nftHistoryModel.DropL2NftHistoryTable())
	assert.Nil(nil, nftCollectionModel.DropL2NftCollectionTable())
	assert.Nil(nil, nftWithdrawHistoryModel.DropL2NftWithdrawHistoryTable())
}

func initTable() {
	assert.Nil(nil, sysConfigModel.CreateSysConfigTable())
	assert.Nil(nil, accountModel.CreateAccountTable())
	assert.Nil(nil, accountHistoryModel.CreateAccountHistoryTable())
	assert.Nil(nil, assetModel.CreateAssetTable())
	assert.Nil(nil, mempoolModel.CreateMempoolTxTable())
	assert.Nil(nil, mempoolDetailModel.CreateMempoolDetailTable())
	assert.Nil(nil, failTxModel.CreateFailTxTable())
	assert.Nil(nil, blockModel.CreateBlockTable())
	assert.Nil(nil, txModel.CreateTxTable())
	assert.Nil(nil, txDetailModel.CreateTxDetailTable())
	assert.Nil(nil, compressedBlockModel.CreateCompressedBlockTable())
	assert.Nil(nil, blockWitnessModel.CreateBlockWitnessTable())
	assert.Nil(nil, proofModel.CreateProofTable())
	assert.Nil(nil, l1SyncedBlockModel.CreateL1SyncedBlockTable())
	assert.Nil(nil, priorityRequestModel.CreatePriorityRequestTable())
	assert.Nil(nil, l1RollupTModel.CreateL1RollupTxTable())
	assert.Nil(nil, liquidityModel.CreateLiquidityTable())
	assert.Nil(nil, liquidityHistoryModel.CreateLiquidityHistoryTable())
	assert.Nil(nil, nftModel.CreateL2NftTable())
	assert.Nil(nil, nftHistoryModel.CreateL2NftHistoryTable())
	assert.Nil(nil, nftCollectionModel.CreateL2NftCollectionTable())
	assert.Nil(nil, nftWithdrawHistoryModel.CreateL2NftWithdrawHistoryTable())
	rowsAffected, err := assetModel.CreateAssetsInBatch(initAssetsInfo())
	if err != nil {
		panic(err)
	}
	logx.Infof("l2 assets info rows affected: %d", rowsAffected)
	rowsAffected, err = sysConfigModel.CreateSysConfigInBatches(initSysConfig())
	if err != nil {
		panic(err)
	}
	logx.Infof("sys config rows affected: %d", rowsAffected)
	err = blockModel.CreateGenesisBlock(&block.Block{
		BlockCommitment:              "0000000000000000000000000000000000000000000000000000000000000000",
		BlockHeight:                  0,
		StateRoot:                    common.Bytes2Hex(tree.NilStateRoot),
		PriorityOperations:           0,
		PendingOnChainOperationsHash: "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
		CommittedTxHash:              "",
		CommittedAt:                  0,
		VerifiedTxHash:               "",
		VerifiedAt:                   0,
		BlockStatus:                  block.StatusVerifiedAndExecuted,
	})
	if err != nil {
		panic(err)
	}
}

type config struct {
	Governance         string
	AssetGovernance    string
	VerifierProxy      string
	ZnsControllerProxy string
	ZnsResolverProxy   string
	ZkbasProxy         string
	UpgradeGateKeeper  string
	LEGToken           string
	REYToken           string
	ERC721             string
	ZnsPriceOracle     string
}
