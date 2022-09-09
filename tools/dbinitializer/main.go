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
 *
 */

package dbinitializer

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/dao/asset"
	"github.com/bnb-chain/zkbnb/dao/block"
	"github.com/bnb-chain/zkbnb/dao/blockwitness"
	"github.com/bnb-chain/zkbnb/dao/compressedblock"
	"github.com/bnb-chain/zkbnb/dao/l1rolluptx"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/dao/liquidity"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/priorityrequest"
	"github.com/bnb-chain/zkbnb/dao/proof"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

type contractAddr struct {
	Governance         string
	AssetGovernance    string
	VerifierProxy      string
	ZnsControllerProxy string
	ZnsResolverProxy   string
	ZkBNBProxy         string
	UpgradeGateKeeper  string
	LEGToken           string
	REYToken           string
	ERC721             string
	ZnsPriceOracle     string
}

type dao struct {
	sysConfigModel        sysconfig.SysConfigModel
	accountModel          account.AccountModel
	accountHistoryModel   account.AccountHistoryModel
	assetModel            asset.AssetModel
	txPoolModel           tx.TxPoolModel
	txDetailModel         tx.TxDetailModel
	txModel               tx.TxModel
	blockModel            block.BlockModel
	compressedBlockModel  compressedblock.CompressedBlockModel
	blockWitnessModel     blockwitness.BlockWitnessModel
	proofModel            proof.ProofModel
	l1SyncedBlockModel    l1syncedblock.L1SyncedBlockModel
	priorityRequestModel  priorityrequest.PriorityRequestModel
	l1RollupTModel        l1rolluptx.L1RollupTxModel
	liquidityModel        liquidity.LiquidityModel
	liquidityHistoryModel liquidity.LiquidityHistoryModel
	nftModel              nft.L2NftModel
	nftHistoryModel       nft.L2NftHistoryModel
}

func Initialize(
	dsn string,
	contractAddrFile string,
	bscTestNetworkRPC, localTestNetworkRPC string,
) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	var svrConf contractAddr
	conf.MustLoad(contractAddrFile, &svrConf)

	unmarshal, _ := json.Marshal(svrConf)
	logx.Infof("init configs: %s", string(unmarshal))

	dao := &dao{
		sysConfigModel:        sysconfig.NewSysConfigModel(db),
		accountModel:          account.NewAccountModel(db),
		accountHistoryModel:   account.NewAccountHistoryModel(db),
		assetModel:            asset.NewAssetModel(db),
		txPoolModel:           tx.NewTxPoolModel(db),
		txDetailModel:         tx.NewTxDetailModel(db),
		txModel:               tx.NewTxModel(db),
		blockModel:            block.NewBlockModel(db),
		compressedBlockModel:  compressedblock.NewCompressedBlockModel(db),
		blockWitnessModel:     blockwitness.NewBlockWitnessModel(db),
		proofModel:            proof.NewProofModel(db),
		l1SyncedBlockModel:    l1syncedblock.NewL1SyncedBlockModel(db),
		priorityRequestModel:  priorityrequest.NewPriorityRequestModel(db),
		l1RollupTModel:        l1rolluptx.NewL1RollupTxModel(db),
		liquidityModel:        liquidity.NewLiquidityModel(db),
		liquidityHistoryModel: liquidity.NewLiquidityHistoryModel(db),
		nftModel:              nft.NewL2NftModel(db),
		nftHistoryModel:       nft.NewL2NftHistoryModel(db),
	}

	dropTables(dao)
	initTable(dao, &svrConf, bscTestNetworkRPC, localTestNetworkRPC)

	return nil
}

func initSysConfig(svrConf *contractAddr, bscTestNetworkRPC, localTestNetworkRPC string) []*sysconfig.SysConfig {
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
			Name:      types.ZkBNBContract,
			Value:     svrConf.ZkBNBProxy,
			ValueType: "string",
			Comment:   "ZkBNB contract on BSC",
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
			Value:     bscTestNetworkRPC,
			ValueType: "string",
			Comment:   "BSC network rpc",
		},
		{
			Name:      types.LocalTestNetworkRpc,
			Value:     localTestNetworkRPC,
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

func dropTables(dao *dao) {
	assert.Nil(nil, dao.sysConfigModel.DropSysConfigTable())
	assert.Nil(nil, dao.accountModel.DropAccountTable())
	assert.Nil(nil, dao.accountHistoryModel.DropAccountHistoryTable())
	assert.Nil(nil, dao.assetModel.DropAssetTable())
	assert.Nil(nil, dao.txPoolModel.DropPoolTxTable())
	assert.Nil(nil, dao.txDetailModel.DropTxDetailTable())
	assert.Nil(nil, dao.txModel.DropTxTable())
	assert.Nil(nil, dao.blockModel.DropBlockTable())
	assert.Nil(nil, dao.compressedBlockModel.DropCompressedBlockTable())
	assert.Nil(nil, dao.blockWitnessModel.DropBlockWitnessTable())
	assert.Nil(nil, dao.proofModel.DropProofTable())
	assert.Nil(nil, dao.l1SyncedBlockModel.DropL1SyncedBlockTable())
	assert.Nil(nil, dao.priorityRequestModel.DropPriorityRequestTable())
	assert.Nil(nil, dao.l1RollupTModel.DropL1RollupTxTable())
	assert.Nil(nil, dao.liquidityModel.DropLiquidityTable())
	assert.Nil(nil, dao.liquidityHistoryModel.DropLiquidityHistoryTable())
	assert.Nil(nil, dao.nftModel.DropL2NftTable())
	assert.Nil(nil, dao.nftHistoryModel.DropL2NftHistoryTable())
}

func initTable(dao *dao, svrConf *contractAddr, bscTestNetworkRPC, localTestNetworkRPC string) {
	assert.Nil(nil, dao.sysConfigModel.CreateSysConfigTable())
	assert.Nil(nil, dao.accountModel.CreateAccountTable())
	assert.Nil(nil, dao.accountHistoryModel.CreateAccountHistoryTable())
	assert.Nil(nil, dao.assetModel.CreateAssetTable())
	assert.Nil(nil, dao.txPoolModel.CreatePoolTxTable())
	assert.Nil(nil, dao.blockModel.CreateBlockTable())
	assert.Nil(nil, dao.txModel.CreateTxTable())
	assert.Nil(nil, dao.txDetailModel.CreateTxDetailTable())
	assert.Nil(nil, dao.compressedBlockModel.CreateCompressedBlockTable())
	assert.Nil(nil, dao.blockWitnessModel.CreateBlockWitnessTable())
	assert.Nil(nil, dao.proofModel.CreateProofTable())
	assert.Nil(nil, dao.l1SyncedBlockModel.CreateL1SyncedBlockTable())
	assert.Nil(nil, dao.priorityRequestModel.CreatePriorityRequestTable())
	assert.Nil(nil, dao.l1RollupTModel.CreateL1RollupTxTable())
	assert.Nil(nil, dao.liquidityModel.CreateLiquidityTable())
	assert.Nil(nil, dao.liquidityHistoryModel.CreateLiquidityHistoryTable())
	assert.Nil(nil, dao.nftModel.CreateL2NftTable())
	assert.Nil(nil, dao.nftHistoryModel.CreateL2NftHistoryTable())
	rowsAffected, err := dao.assetModel.CreateAssets(initAssetsInfo())
	if err != nil {
		panic(err)
	}
	logx.Infof("l2 assets info rows affected: %d", rowsAffected)
	rowsAffected, err = dao.sysConfigModel.CreateSysConfigs(initSysConfig(svrConf, bscTestNetworkRPC, localTestNetworkRPC))
	if err != nil {
		panic(err)
	}
	logx.Infof("sys config rows affected: %d", rowsAffected)
	err = dao.blockModel.CreateGenesisBlock(&block.Block{
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
