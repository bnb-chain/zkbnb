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

package proverUtil

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/bnb-chain/bas-smt/database/memory"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/basic"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/pkg/treedb"
)

func TestConstructRegisterZnsCryptoTxFirst(t *testing.T) {
	redisConn := redis.New(basic.CacheConf[0].Host, WithRedis(basic.CacheConf[0].Type, basic.CacheConf[0].Pass))
	txModel := tx.NewTxModel(basic.Connection, basic.CacheConf, basic.DB, redisConn)
	accountModel := account.NewAccountModel(basic.Connection, basic.CacheConf, basic.DB)
	accountHistoryModel := account.NewAccountHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	//liquidityModel := liquidity.NewLiquidityModel(basic.Connection, basic.CacheConf, basic.DB)
	liquidityHistoryModel := liquidity.NewLiquidityHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	//nftModel := nft.NewL2NftModel(basic.Connection, basic.CacheConf, basic.DB)
	nftHistoryModel := nft.NewL2NftHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	ctx := &treedb.Context{
		Driver: treedb.MemoryDB,
		TreeDB: memory.NewMemoryDB(),
	}
	txInfo, err := txModel.GetTxByTxId(3)
	if err != nil {
		t.Fatal(err)
	}
	blockHeight := int64(2)
	accountTree, accountAssetTrees, err := tree.InitAccountTree(accountModel, accountHistoryModel, blockHeight, ctx)
	if err != nil {
		t.Fatal(err)
	}
	liquidityTree, err := tree.InitLiquidityTree(liquidityHistoryModel, blockHeight, ctx)
	if err != nil {
		t.Fatal(err)
	}
	nftTree, err := tree.InitNftTree(nftHistoryModel, blockHeight, ctx)
	if err != nil {
		t.Fatal(err)
	}
	cryptoTx, err := NewWitnessHelper(ctx,
		accountTree,
		liquidityTree,
		nftTree,
		&accountAssetTrees,
		accountModel).constructCryptoTx(txInfo, 0)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(common.Bytes2Hex(tree.NilAccountNodeHash))
	log.Println(common.Bytes2Hex(cryptoTx.MerkleProofsAccountBefore[0][0]))
	txBytes, err := json.Marshal(cryptoTx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(txBytes))
}

func TestConstructRegisterZnsCryptoTxNotFirst(t *testing.T) {
	redisConn := redis.New(basic.CacheConf[0].Host, WithRedis(basic.CacheConf[0].Type, basic.CacheConf[0].Pass))
	txModel := tx.NewTxModel(basic.Connection, basic.CacheConf, basic.DB, redisConn)
	accountModel := account.NewAccountModel(basic.Connection, basic.CacheConf, basic.DB)
	accountHistoryModel := account.NewAccountHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	//liquidityModel := liquidity.NewLiquidityModel(basic.Connection, basic.CacheConf, basic.DB)
	liquidityHistoryModel := liquidity.NewLiquidityHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	//nftModel := nft.NewL2NftModel(basic.Connection, basic.CacheConf, basic.DB)
	nftHistoryModel := nft.NewL2NftHistoryModel(basic.Connection, basic.CacheConf, basic.DB)
	ctx := &treedb.Context{
		Driver: treedb.MemoryDB,
		TreeDB: memory.NewMemoryDB(),
	}
	txInfo, err := txModel.GetTxByTxHash("e5d6dd7c-da46-11ec-8abf-7cb27d9ca483")
	if err != nil {
		t.Fatal(err)
	}
	blockHeight := int64(1)
	accountTree, accountAssetTrees, err := tree.InitAccountTree(accountModel, accountHistoryModel, blockHeight, ctx)
	if err != nil {
		t.Fatal(err)
	}
	liquidityTree, err := tree.InitLiquidityTree(liquidityHistoryModel, blockHeight, ctx)
	if err != nil {
		t.Fatal(err)
	}
	nftTree, err := tree.InitNftTree(nftHistoryModel, blockHeight, ctx)
	if err != nil {
		t.Fatal(err)
	}
	cryptoTx, err := NewWitnessHelper(ctx,
		accountTree,
		liquidityTree,
		nftTree,
		&accountAssetTrees,
		accountModel).constructCryptoTx(txInfo, 0)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := json.Marshal(cryptoTx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(txBytes))
}
