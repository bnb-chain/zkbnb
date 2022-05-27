/*
 * Copyright © 2021 Zecrey Protocol
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

package init

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/basic"
	"github.com/zecrey-labs/zecrey-legend/common/model/block"
	"github.com/zecrey-labs/zecrey-legend/common/model/l1BlockMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l1TxSender"
	"github.com/zecrey-labs/zecrey-legend/common/model/l1amount"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2BlockEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2asset"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/common/model/proofSender"
	"github.com/zecrey-labs/zecrey-legend/common/model/sysconfig"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"log"
	"testing"
)

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

var (
	redisConn = redis.New(basic.CacheConf[0].Host, WithRedis(basic.CacheConf[0].Type, basic.CacheConf[0].Pass))
	// sys config
	sysconfigModel = sysconfig.NewSysconfigModel(basic.Connection, basic.CacheConf, basic.DB)
	// price
	//priceModel = price.NewPriceModel(basic.Connection, basic.CacheConf, basic.DB)
	// account
	accountModel        = account.NewAccountModel(basic.Connection, basic.CacheConf, basic.DB)
	accountHistoryModel = account.NewAccountHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	// l2 asset
	l2AssetInfoModel = l2asset.NewL2AssetInfoModel(basic.Connection, basic.CacheConf, basic.DB)
	// mempool
	mempoolDetailModel = mempool.NewMempoolDetailModel(basic.Connection, basic.CacheConf, basic.DB)
	mempoolModel       = mempool.NewMempoolModel(basic.Connection, basic.CacheConf, basic.DB)
	// tx
	failTxModel   = tx.NewFailTxModel(basic.Connection, basic.CacheConf, basic.DB)
	txDetailModel = tx.NewTxDetailModel(basic.Connection, basic.CacheConf, basic.DB)
	txModel       = tx.NewTxModel(basic.Connection, basic.CacheConf, basic.DB, redisConn)
	// block
	blockModel = block.NewBlockModel(basic.Connection, basic.CacheConf, basic.DB, redisConn)
	// block for proverUtil
	proofSenderModel = proofSender.NewProofSenderModel(basic.DB)
	// monitor
	l1BlockMonitorModel      = l1BlockMonitor.NewL1BlockMonitorModel(basic.Connection, basic.CacheConf, basic.DB)
	l2TxEventMonitorModel    = l2TxEventMonitor.NewL2TxEventMonitorModel(basic.Connection, basic.CacheConf, basic.DB)
	l2BlockEventMonitorModel = l2BlockEventMonitor.NewL2BlockEventMonitorModel(basic.Connection, basic.CacheConf, basic.DB)
	// sender
	l1TxSenderModel = l1TxSender.NewL1TxSenderModel(basic.Connection, basic.CacheConf, basic.DB)
	// l1 amount
	l1AmountModel = l1amount.NewL1AmountModel(basic.Connection, basic.CacheConf, basic.DB)
	// liquidity
	liquidityModel        = liquidity.NewLiquidityModel(basic.Connection, basic.CacheConf, basic.DB)
	liquidityHistoryModel = liquidity.NewLiquidityHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	// nft
	nftModel                = nft.NewL2NftModel(basic.Connection, basic.CacheConf, basic.DB)
	offerModel              = nft.NewOfferModel(basic.Connection, basic.CacheConf, basic.DB)
	nftHistoryModel         = nft.NewL2NftHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	nftExchangeModel        = nft.NewL2NftExchangeModel(basic.Connection, basic.CacheConf, basic.DB)
	nftExchangeHistoryModel = nft.NewL2NftExchangeHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	nftCollectionModel      = nft.NewL2NftCollectionModel(basic.Connection, basic.CacheConf, basic.DB)
)

func TestDropTables(t *testing.T) {
	sysconfigModel.DropSysconfigTable()
	//priceModel.
	accountModel.DropAccountTable()
	accountHistoryModel.DropAccountHistoryTable()
	l2AssetInfoModel.DropL2AssetInfoTable()
	mempoolDetailModel.DropMempoolDetailTable()
	mempoolModel.DropMempoolTxTable()
	failTxModel.DropFailTxTable()
	txDetailModel.DropTxDetailTable()
	txModel.DropTxTable()
	blockModel.DropBlockTable()
	proofSenderModel.DropProofSenderTable()
	l1BlockMonitorModel.DropL1BlockMonitorTable()
	l2TxEventMonitorModel.DropL2TxEventMonitorTable()
	l2BlockEventMonitorModel.DropL2BlockEventMonitorTable()
	l1TxSenderModel.DropL1TxSenderTable()
	l1AmountModel.DropL1AmountTable()
	liquidityModel.DropLiquidityTable()
	liquidityHistoryModel.DropLiquidityHistoryTable()
	nftModel.DropL2NftTable()
	offerModel.DropOfferTable()
	nftHistoryModel.DropL2NftHistoryTable()
	nftExchangeModel.DropL2NftExchangeTable()
	nftExchangeHistoryModel.DropL2NftExchangeHistoryTable()
	nftCollectionModel.DropL2NftCollectionTable()
}

func TestDataInitialize(t *testing.T) {
	// create tables
	sysconfigModel.CreateSysconfigTable()
	//priceModel.
	accountModel.CreateAccountTable()
	accountHistoryModel.CreateAccountHistoryTable()
	l2AssetInfoModel.CreateL2AssetInfoTable()
	mempoolDetailModel.CreateMempoolDetailTable()
	mempoolModel.CreateMempoolTxTable()
	failTxModel.CreateFailTxTable()
	txDetailModel.CreateTxDetailTable()
	txModel.CreateTxTable()
	blockModel.CreateBlockTable()
	proofSenderModel.CreateProofSenderTable()
	l1BlockMonitorModel.CreateL1BlockMonitorTable()
	l2TxEventMonitorModel.CreateL2TxEventMonitorTable()
	l2BlockEventMonitorModel.CreateL2BlockEventMonitorTable()
	l1TxSenderModel.CreateL1TxSenderTable()
	l1AmountModel.CreateL1AmountTable()
	liquidityModel.CreateLiquidityTable()
	liquidityHistoryModel.CreateLiquidityHistoryTable()
	nftModel.CreateL2NftTable()
	offerModel.CreateOfferTable()
	nftHistoryModel.CreateL2NftHistoryTable()
	nftExchangeModel.CreateL2NftExchangeTable()
	nftExchangeHistoryModel.CreateL2NftExchangeHistoryTable()
	nftCollectionModel.CreateL2NftCollectionTable()

	// init l1 asset info
	rowsAffected, err := l2AssetInfoModel.CreateL2AssetsInfoInBatches(initAssetsInfo())
	if err != nil {
		t.Fatal(err)
	}
	log.Println("l2 assets info rows affected:", rowsAffected)
	// sys config
	rowsAffected, err = sysconfigModel.CreateSysconfigInBatches(initSysConfig())
	if err != nil {
		t.Fatal(err)
	}
	log.Println("sys config rows affected:", rowsAffected)
	// genesis block
	err = blockModel.CreateGenesisBlock(&block.Block{
		BlockCommitment:              "0000000000000000000000000000000000000000000000000000000000000000",
		BlockHeight:                  0,
		AccountRoot:                  common.Bytes2Hex(tree.NilAccountRoot),
		PriorityOperations:           0,
		PendingOnchainOperationsHash: "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
		CommittedTxHash:              "",
		CommittedAt:                  0,
		VerifiedTxHash:               "",
		VerifiedAt:                   0,
		BlockStatus:                  block.StatusVerified,
	})
	if err != nil {
		t.Fatal(err)
	}
}
