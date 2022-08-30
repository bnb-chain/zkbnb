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
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas-smt/database/memory"
	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/dao/liquidity"
	"github.com/bnb-chain/zkbas/dao/nft"
	"github.com/bnb-chain/zkbas/dao/tx"
	"github.com/bnb-chain/zkbas/tree"
)

var (
	dsn                   = "host=localhost user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable"
	txModel               tx.TxModel
	accountModel          account.AccountModel
	accountHistoryModel   account.AccountHistoryModel
	liquidityHistoryModel liquidity.LiquidityHistoryModel
	nftHistoryModel       nft.L2NftHistoryModel
)

func TestConstructTxWitness(t *testing.T) {
	testDbSetup()

	allExampleTxCases := []struct {
		txId            int64
		expectedWitness string
	}{
		// TODO add more tx example once we get the full data
		{1, ""},
		{2, ""},
	}

	for _, example := range allExampleTxCases {
		witness, err := constructTxWitness(example.txId)
		assert.NoError(t, err)
		txBytes, err := json.Marshal(witness)
		assert.NoError(t, err)
		assert.Equal(t, string(txBytes), example.expectedWitness)
	}
}

func constructTxWitness(txId int64) (*TxWitness, error) {
	ctx := &tree.Context{
		Driver: tree.MemoryDB,
		TreeDB: memory.NewMemoryDB(),
	}
	txInfo, err := txModel.GetTxById(txId)
	if err != nil {
		return nil, err
	}
	blockHeight := txInfo.BlockHeight
	accountTree, accountAssetTrees, err := tree.InitAccountTree(accountModel, accountHistoryModel, blockHeight, ctx)
	if err != nil {
		return nil, err
	}
	liquidityTree, err := tree.InitLiquidityTree(liquidityHistoryModel, blockHeight, ctx)
	if err != nil {
		return nil, err
	}
	nftTree, err := tree.InitNftTree(nftHistoryModel, blockHeight, ctx)
	if err != nil {
		return nil, err
	}
	return NewWitnessHelper(ctx,
		accountTree,
		liquidityTree,
		nftTree,
		&accountAssetTrees,
		accountModel).constructTxWitness(txInfo, 0)
}

func testDbSetup() {
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	txModel = tx.NewTxModel(db)
	accountModel = account.NewAccountModel(db)
	accountHistoryModel = account.NewAccountHistoryModel(db)
	liquidityHistoryModel = liquidity.NewLiquidityHistoryModel(db)
	nftHistoryModel = nft.NewL2NftHistoryModel(db)
}
