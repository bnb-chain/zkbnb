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

package tree

import (
	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/zkbas-crypto/hash/bn254/zmimc"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/treedb"
)

func InitLiquidityTree(
	liquidityHistoryModel LiquidityHistoryModel,
	blockHeight int64,
	ctx *treedb.Context,
) (
	liquidityTree bsmt.SparseMerkleTree, err error,
) {

	liquidityTree, err = bsmt.NewBASSparseMerkleTree(bsmt.NewHasher(zmimc.Hmimc),
		treedb.SetNamespace(ctx, LiquidityPrefix), LiquidityTreeHeight, NilLiquidityNodeHash,
		ctx.Options(blockHeight)...)
	if err != nil {
		logx.Errorf("unable to create tree from db: %s", err.Error())
		return nil, err
	}

	if ctx.IsLoad() {
		nums, err := liquidityHistoryModel.GetLatestLiquidityCountByBlockHeight(blockHeight)
		if err != nil {
			logx.Errorf("unable to get latest liquidity assets: %s", err.Error())
			return nil, err
		}
		for i := 0; i < int(nums); i += ctx.BatchReloadSize() {
			err := loadLiquidityTreeFromRDB(
				liquidityHistoryModel, blockHeight,
				i, i+ctx.BatchReloadSize(), liquidityTree)
			if err != nil {
				return nil, err
			}
			_, err = liquidityTree.Commit(nil)
			if err != nil {
				logx.Errorf("unable to commit liquidity tree: %s", err.Error())
				return nil, err
			}
		}

		return liquidityTree, nil
	}

	// It's not loading from RDB, need to check tree version
	if liquidityTree.LatestVersion() > bsmt.Version(blockHeight) && !liquidityTree.IsEmpty() {
		logx.Infof("liquidity tree version [%d] is higher than block, rollback to %d", liquidityTree.LatestVersion(), blockHeight)
		err := liquidityTree.Rollback(bsmt.Version(blockHeight))
		if err != nil {
			logx.Errorf("unable to rollback liquidity tree: %s, version: %d", err.Error(), blockHeight)
			return nil, err
		}
	}

	return liquidityTree, nil
}

func loadLiquidityTreeFromRDB(
	liquidityHistoryModel LiquidityHistoryModel,
	blockHeight int64,
	offset, limit int,
	liquidityTree bsmt.SparseMerkleTree,
) error {
	liquidityAssets, err := liquidityHistoryModel.GetLatestLiquidityByBlockHeight(blockHeight,
		limit, offset)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			logx.Errorf("unable to get latest liquidity assets: %s", err.Error())
			return err
		}
	}
	for _, liquidityAsset := range liquidityAssets {
		pairIndex := liquidityAsset.PairIndex
		hashVal, err := LiquidityAssetToNode(
			liquidityAsset.AssetAId, liquidityAsset.AssetA,
			liquidityAsset.AssetBId, liquidityAsset.AssetB,
			liquidityAsset.LpAmount, liquidityAsset.KLast,
			liquidityAsset.FeeRate, liquidityAsset.TreasuryAccountIndex, liquidityAsset.TreasuryRate)
		if err != nil {
			logx.Errorf("unable to convert liquidity asset to node: %s", err.Error())
			return err
		}
		err = liquidityTree.Set(uint64(pairIndex), hashVal)
		if err != nil {
			logx.Errorf("unable to write liquidity asset to tree: %s", err.Error())
			return err
		}
	}
	return nil
}

func LiquidityAssetToNode(
	assetAId int64,
	assetA string,
	assetBId int64,
	assetB string,
	lpAmount string,
	kLast string,
	feeRate int64,
	treasuryAccountIndex int64,
	treasuryFeeRate int64,
) (hashVal []byte, err error) {
	hashVal, err = ComputeLiquidityAssetLeafHash(
		assetAId, assetA,
		assetBId, assetB,
		lpAmount,
		kLast,
		feeRate,
		treasuryAccountIndex,
		treasuryFeeRate,
	)
	if err != nil {
		logx.Errorf("unable to compute liquidity asset leaf hash: %s", err.Error())
		return nil, err
	}
	return hashVal, nil
}
