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

package prove

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bnb-chain/bas-smt/database/memory"
	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/tree"
)

func TestConstructSwapCryptoTxFirst(t *testing.T) {
	txModel := tx.NewTxModel(connection, cacheConf, db)
	accountModel := account.NewAccountModel(connection, cacheConf, db)
	accountHistoryModel := account.NewAccountHistoryModel(connection, cacheConf, db)
	//liquidityModel := liquidity.NewLiquidityModel(basic.connection, basic.cacheConf, basic.db)
	liquidityHistoryModel := liquidity.NewLiquidityHistoryModel(connection, cacheConf, db)
	//nftModel := nft.NewL2NftModel(basic.connection, basic.cacheConf, basic.db)
	nftHistoryModel := nft.NewL2NftHistoryModel(connection, cacheConf, db)
	ctx := &tree.Context{
		Driver: tree.MemoryDB,
		TreeDB: memory.NewMemoryDB(),
	}
	txInfo, err := txModel.GetTxById(19)
	if err != nil {
		t.Fatal(err)
	}
	blockHeight := int64(18)
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
