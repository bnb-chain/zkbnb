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
	"github.com/bnb-chain/zkbnb/dao/desertexit"
	"github.com/bnb-chain/zkbnb/dao/rollback"
	"github.com/bnb-chain/zkbnb/tools/desertexit/config"
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
	CommitAddress      string
	VerifyAddress      string
	UpgradeGateKeeper  string
	BUSDToken          string
	LEGToken           string
	REYToken           string
	ERC721             string
	ZnsPriceOracle     string
	DefaultNftFactory  string
}

type dao struct {
	sysConfigModel          sysconfig.SysConfigModel
	accountModel            account.AccountModel
	accountHistoryModel     account.AccountHistoryModel
	assetModel              asset.AssetModel
	txPoolModel             tx.TxPoolModel
	txDetailModel           tx.TxDetailModel
	txModel                 tx.TxModel
	blockModel              block.BlockModel
	compressedBlockModel    compressedblock.CompressedBlockModel
	blockWitnessModel       blockwitness.BlockWitnessModel
	proofModel              proof.ProofModel
	l1SyncedBlockModel      l1syncedblock.L1SyncedBlockModel
	priorityRequestModel    priorityrequest.PriorityRequestModel
	l1RollupTModel          l1rolluptx.L1RollupTxModel
	nftModel                nft.L2NftModel
	nftHistoryModel         nft.L2NftHistoryModel
	rollbackModel           rollback.RollbackModel
	nftMetadataHistoryModel nft.L2NftMetadataHistoryModel
	desertExitBlockModel    desertexit.DesertExitBlockModel
}

func Initialize(
	dsn string,
	contractAddrFile string,
	bscTestNetworkRPC, localTestNetworkRPC, optionalBlockSizes string,
) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}
	var svrConf contractAddr
	conf.MustLoad(contractAddrFile, &svrConf)

	unmarshal, _ := json.Marshal(svrConf)
	logx.Infof("init configs: %s", string(unmarshal))

	dao := &dao{
		sysConfigModel:          sysconfig.NewSysConfigModel(db),
		accountModel:            account.NewAccountModel(db),
		accountHistoryModel:     account.NewAccountHistoryModel(db),
		assetModel:              asset.NewAssetModel(db),
		txPoolModel:             tx.NewTxPoolModel(db),
		txDetailModel:           tx.NewTxDetailModel(db),
		txModel:                 tx.NewTxModel(db),
		blockModel:              block.NewBlockModel(db),
		compressedBlockModel:    compressedblock.NewCompressedBlockModel(db),
		blockWitnessModel:       blockwitness.NewBlockWitnessModel(db),
		proofModel:              proof.NewProofModel(db),
		l1SyncedBlockModel:      l1syncedblock.NewL1SyncedBlockModel(db),
		priorityRequestModel:    priorityrequest.NewPriorityRequestModel(db),
		l1RollupTModel:          l1rolluptx.NewL1RollupTxModel(db),
		nftModel:                nft.NewL2NftModel(db),
		nftHistoryModel:         nft.NewL2NftHistoryModel(db),
		rollbackModel:           rollback.NewRollbackModel(db),
		nftMetadataHistoryModel: nft.NewL2NftMetadataHistoryModel(db),
	}

	dropTables(dao)
	initTable(dao, &svrConf, bscTestNetworkRPC, localTestNetworkRPC, optionalBlockSizes)

	return nil
}

func InitializeDesertExit(
	configFile string,
) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	db, err := gorm.Open(postgres.Open(c.Postgres.MasterDataSource), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}
	dao := &dao{
		accountModel:         account.NewAccountModel(db),
		nftModel:             nft.NewL2NftModel(db),
		l1SyncedBlockModel:   l1syncedblock.NewL1SyncedBlockModel(db),
		desertExitBlockModel: desertexit.NewDesertExitBlockModel(db),
		priorityRequestModel: priorityrequest.NewPriorityRequestModel(db),
	}

	dropTablesDesertExit(dao)
	initTableDesertExit(dao)

	return nil
}

func initSysConfig(svrConf *contractAddr, bscTestNetworkRPC, localTestNetworkRPC, optionalBlockSizes string) []*sysconfig.SysConfig {

	// to config gas for different transaction types, need to be evaluated and tune these values
	bnbGasFee := make(map[int]int64)
	bnbGasFee[types.TxTypeChangePubKey] = 10000000000000
	bnbGasFee[types.TxTypeTransfer] = 10000000000000
	bnbGasFee[types.TxTypeWithdraw] = 20000000000000
	bnbGasFee[types.TxTypeCreateCollection] = 10000000000000
	bnbGasFee[types.TxTypeMintNft] = 10000000000000
	bnbGasFee[types.TxTypeTransferNft] = 12000000000000
	bnbGasFee[types.TxTypeAtomicMatch] = 18000000000000
	bnbGasFee[types.TxTypeCancelOffer] = 12000000000000
	bnbGasFee[types.TxTypeWithdrawNft] = 20000000000000

	busdGasFee := make(map[int]int64)
	busdGasFee[types.TxTypeChangePubKey] = 10000000000000
	busdGasFee[types.TxTypeTransfer] = 10000000000000
	busdGasFee[types.TxTypeWithdraw] = 20000000000000
	busdGasFee[types.TxTypeCreateCollection] = 10000000000000
	busdGasFee[types.TxTypeMintNft] = 10000000000000
	busdGasFee[types.TxTypeTransferNft] = 12000000000000
	busdGasFee[types.TxTypeAtomicMatch] = 18000000000000
	busdGasFee[types.TxTypeCancelOffer] = 12000000000000
	busdGasFee[types.TxTypeWithdrawNft] = 20000000000000

	gasFeeConfig := make(map[uint32]map[int]int64) // asset id -> (tx type -> gas fee value)
	gasFeeConfig[types.BNBAssetId] = bnbGasFee     // bnb asset
	gasFeeConfig[types.BUSDAssetId] = busdGasFee   // busd asset

	gas, err := json.Marshal(gasFeeConfig)
	if err != nil {
		logx.Severe("fail to marshal gas fee config")
		panic("fail to marshal gas fee config")
	}

	return []*sysconfig.SysConfig{
		{
			Name:      types.SysGasFee,
			Value:     string(gas),
			ValueType: "string",
			Comment:   "based on BNB",
		},
		{
			Name:      types.ProtocolRate,
			Value:     "200",
			ValueType: "int",
			Comment:   "protocol rate",
		},
		{
			Name:      types.ProtocolAccountIndex,
			Value:     "0",
			ValueType: "int",
			Comment:   "protocol index",
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
		{
			Name:      types.CommitAddress,
			Value:     svrConf.CommitAddress,
			ValueType: "string",
			Comment:   "ZkBNB commit on BSC",
		},
		{
			Name:      types.VerifyAddress,
			Value:     svrConf.VerifyAddress,
			ValueType: "string",
			Comment:   "ZkBNB verify on BSC",
		},
		// Governance Contract
		{
			Name:      types.GovernanceContract,
			Value:     svrConf.Governance,
			ValueType: "string",
			Comment:   "Governance contract on BSC",
		},
		{
			Name:      types.OptionalBlockSizes,
			Value:     optionalBlockSizes,
			ValueType: "string",
			Comment:   "OptionalBlockSizes config for committer and prover",
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
		{
			Name:      types.DefaultNftFactory,
			Value:     svrConf.DefaultNftFactory,
			ValueType: "string",
			Comment:   "ZkBNB default nft factory contract on BSC",
		},
	}
}

func initAssetsInfo(busdAddress string) []*asset.Asset {
	return []*asset.Asset{
		{
			AssetId:     types.BNBAssetId,
			L1Address:   types.BNBAddress,
			AssetName:   "BNB",
			AssetSymbol: "BNB",
			Decimals:    18,
			Status:      0,
			IsGasAsset:  asset.IsGasAsset,
		},
		{
			AssetId:     types.BUSDAssetId,
			L1Address:   busdAddress,
			AssetName:   "BUSD",
			AssetSymbol: "BUSD",
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
	assert.Nil(nil, dao.nftModel.DropL2NftTable())
	assert.Nil(nil, dao.nftHistoryModel.DropL2NftHistoryTable())
	assert.Nil(nil, dao.rollbackModel.DropRollbackTable())
	assert.Nil(nil, dao.nftMetadataHistoryModel.DropL2NftMetadataHistoryTable())

}

func dropTablesDesertExit(dao *dao) {
	assert.Nil(nil, dao.accountModel.DropAccountTable())
	assert.Nil(nil, dao.nftModel.DropL2NftTable())
	assert.Nil(nil, dao.l1SyncedBlockModel.DropL1SyncedBlockTable())
	assert.Nil(nil, dao.desertExitBlockModel.DropDesertExitBlockTable())
	assert.Nil(nil, dao.priorityRequestModel.DropPriorityRequestTable())

}

func initTable(dao *dao, svrConf *contractAddr, bscTestNetworkRPC, localTestNetworkRPC, optionalBlockSizes string) {
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
	assert.Nil(nil, dao.nftModel.CreateL2NftTable())
	assert.Nil(nil, dao.nftHistoryModel.CreateL2NftHistoryTable())
	assert.Nil(nil, dao.rollbackModel.CreateRollbackTable())
	assert.Nil(nil, dao.nftMetadataHistoryModel.CreateL2NftMetadataHistoryTable())
	rowsAffected, err := dao.assetModel.CreateAssets(initAssetsInfo(svrConf.BUSDToken))
	if err != nil {
		logx.Severef("failed to initialize assets data, %v", err)
		panic("failed to initialize assets data, err:" + err.Error())
	}
	logx.Infof("l2 assets info rows affected: %d", rowsAffected)
	rowsAffected, err = dao.sysConfigModel.CreateSysConfigs(initSysConfig(svrConf, bscTestNetworkRPC, localTestNetworkRPC, optionalBlockSizes))
	if err != nil {
		logx.Severef("failed to initialize system configuration data, %v", err)
		panic("failed to initialize system configuration data, err:" + err.Error())
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
		logx.Severef("failed to create the genesis block data, %v", err)
		panic("failed to create the genesis block data, err:" + err.Error())
	}
}

func initTableDesertExit(dao *dao) {
	assert.Nil(nil, dao.accountModel.CreateAccountTable())
	assert.Nil(nil, dao.nftModel.CreateL2NftTable())
	assert.Nil(nil, dao.l1SyncedBlockModel.CreateL1SyncedBlockTable())
	assert.Nil(nil, dao.desertExitBlockModel.CreateDesertExitBlockTable())
	assert.Nil(nil, dao.priorityRequestModel.CreatePriorityRequestTable())
}
