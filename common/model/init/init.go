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

package main

import (
	"encoding/json"
	"flag"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/asset"
	"github.com/bnb-chain/zkbas/common/model/basic"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/blockForProof"
	"github.com/bnb-chain/zkbas/common/model/l1RollupTx"
	"github.com/bnb-chain/zkbas/common/model/l1SyncedBlock"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/priorityRequest"
	"github.com/bnb-chain/zkbas/common/model/proof"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/sysConfigName"
	"github.com/bnb-chain/zkbas/common/tree"
)

var configFile = flag.String("f", "./contractaddr.yaml", "the config file")
var svrConf config

const (
	BSC_Test_Network_RPC   = "http://tf-dex-preview-validator-nlb-6fd109ac8b9d390a.elb.ap-northeast-1.amazonaws.com:8545"
	Local_Test_Network_RPC = "http://127.0.0.1:8545/"
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
			Name:      sysConfigName.SysGasFee,
			Value:     "100000000000000",
			ValueType: "string",
			Comment:   "based on BNB",
		},
		{
			Name:      sysConfigName.TreasuryAccountIndex,
			Value:     "0",
			ValueType: "int",
			Comment:   "treasury index",
		},
		{
			Name:      sysConfigName.GasAccountIndex,
			Value:     "1",
			ValueType: "int",
			Comment:   "gas index",
		},
		{
			Name:      sysConfigName.ZkbasContract,
			Value:     svrConf.ZkbasProxy,
			ValueType: "string",
			Comment:   "Zkbas contract on BSC",
		},
		// Governance Contract
		{
			Name:      sysConfigName.GovernanceContract,
			Value:     svrConf.Governance,
			ValueType: "string",
			Comment:   "Governance contract on BSC",
		},
		// network rpc
		{
			Name:      sysConfigName.BscTestNetworkRpc,
			Value:     BSC_Test_Network_RPC,
			ValueType: "string",
			Comment:   "BSC network rpc",
		},
		{
			Name:      sysConfigName.LocalTestNetworkRpc,
			Value:     Local_Test_Network_RPC,
			ValueType: "string",
			Comment:   "Local network rpc",
		},
		{
			Name:      sysConfigName.ZnsPriceOracle,
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
	sysConfigModel          = sysconfig.NewSysConfigModel(basic.Connection, basic.CacheConf, basic.DB)
	accountModel            = account.NewAccountModel(basic.Connection, basic.CacheConf, basic.DB)
	accountHistoryModel     = account.NewAccountHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	assetModel              = asset.NewAssetModel(basic.Connection, basic.CacheConf, basic.DB)
	mempoolDetailModel      = mempool.NewMempoolDetailModel(basic.Connection, basic.CacheConf, basic.DB)
	mempoolModel            = mempool.NewMempoolModel(basic.Connection, basic.CacheConf, basic.DB)
	failTxModel             = tx.NewFailTxModel(basic.Connection, basic.CacheConf, basic.DB)
	txDetailModel           = tx.NewTxDetailModel(basic.Connection, basic.CacheConf, basic.DB)
	txModel                 = tx.NewTxModel(basic.Connection, basic.CacheConf, basic.DB)
	blockModel              = block.NewBlockModel(basic.Connection, basic.CacheConf, basic.DB)
	blockForCommitModel     = blockForCommit.NewBlockForCommitModel(basic.Connection, basic.CacheConf, basic.DB)
	blockForProofModel      = blockForProof.NewBlockForProofModel(basic.Connection, basic.CacheConf, basic.DB)
	proofModel              = proof.NewProofModel(basic.DB)
	l1SyncedBlockModel      = l1SyncedBlock.NewL1SyncedBlockModel(basic.Connection, basic.CacheConf, basic.DB)
	priorityRequestModel    = priorityRequest.NewPriorityRequestModel(basic.Connection, basic.CacheConf, basic.DB)
	l1RollupTModel          = l1RollupTx.NewL1RollupTxModel(basic.Connection, basic.CacheConf, basic.DB)
	liquidityModel          = liquidity.NewLiquidityModel(basic.Connection, basic.CacheConf, basic.DB)
	liquidityHistoryModel   = liquidity.NewLiquidityHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	nftModel                = nft.NewL2NftModel(basic.Connection, basic.CacheConf, basic.DB)
	offerModel              = nft.NewOfferModel(basic.Connection, basic.CacheConf, basic.DB)
	nftHistoryModel         = nft.NewL2NftHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	nftExchangeModel        = nft.NewL2NftExchangeModel(basic.Connection, basic.CacheConf, basic.DB)
	nftCollectionModel      = nft.NewL2NftCollectionModel(basic.Connection, basic.CacheConf, basic.DB)
	nftWithdrawHistoryModel = nft.NewL2NftWithdrawHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
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
	assert.Nil(nil, blockForCommitModel.DropBlockForCommitTable())
	assert.Nil(nil, blockForProofModel.DropBlockForProofTable())
	assert.Nil(nil, proofModel.DropProofTable())
	assert.Nil(nil, l1SyncedBlockModel.DropL1SyncedBlockTable())
	assert.Nil(nil, priorityRequestModel.DropPriorityRequestTable())
	assert.Nil(nil, l1RollupTModel.DropL1RollupTxTable())
	assert.Nil(nil, liquidityModel.DropLiquidityTable())
	assert.Nil(nil, liquidityHistoryModel.DropLiquidityHistoryTable())
	assert.Nil(nil, nftModel.DropL2NftTable())
	assert.Nil(nil, offerModel.DropOfferTable())
	assert.Nil(nil, nftHistoryModel.DropL2NftHistoryTable())
	assert.Nil(nil, nftExchangeModel.DropL2NftExchangeTable())
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
	assert.Nil(nil, blockForCommitModel.CreateBlockForCommitTable())
	assert.Nil(nil, blockForProofModel.CreateBlockForProofTable())
	assert.Nil(nil, proofModel.CreateProofTable())
	assert.Nil(nil, l1SyncedBlockModel.CreateL1SyncedBlockTable())
	assert.Nil(nil, priorityRequestModel.CreatePriorityRequestTable())
	assert.Nil(nil, l1RollupTModel.CreateL1RollupTxTable())
	assert.Nil(nil, liquidityModel.CreateLiquidityTable())
	assert.Nil(nil, liquidityHistoryModel.CreateLiquidityHistoryTable())
	assert.Nil(nil, nftModel.CreateL2NftTable())
	assert.Nil(nil, offerModel.CreateOfferTable())
	assert.Nil(nil, nftHistoryModel.CreateL2NftHistoryTable())
	assert.Nil(nil, nftExchangeModel.CreateL2NftExchangeTable())
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
