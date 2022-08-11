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
	"errors"

	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/block"
)

func SetFixedAccountArray(proof [][]byte) (res [AccountMerkleLevels][]byte, err error) {
	if len(proof) != AccountMerkleLevels {
		logx.Errorf("[SetFixedAccountArray] invalid size")
		return res, errors.New("[SetFixedAccountArray] invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedAccountAssetArray(proof [][]byte) (res [AssetMerkleLevels][]byte, err error) {
	if len(proof) != AssetMerkleLevels {
		logx.Errorf("[SetFixedAccountAssetArray] invalid size")
		return res, errors.New("[SetFixedAccountAssetArray] invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedLiquidityArray(proof [][]byte) (res [LiquidityMerkleLevels][]byte, err error) {
	if len(proof) != LiquidityMerkleLevels {
		logx.Errorf("[SetFixedLiquidityArray] invalid size")
		return res, errors.New("[SetFixedLiquidityArray] invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedNftArray(proof [][]byte) (res [NftMerkleLevels][]byte, err error) {
	if len(proof) != NftMerkleLevels {
		logx.Errorf("[SetFixedNftArray] invalid size")
		return res, errors.New("[SetFixedNftArray] invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func BlockToCryptoBlock(
	oBlock *block.Block,
	oldStateRoot, newStateRoot []byte,
	cryptoTxs []*cryptoBlock.Tx,
) (cBlock *cryptoBlock.Block, err error) {
	cBlock = &cryptoBlock.Block{
		BlockNumber:     oBlock.BlockHeight,
		CreatedAt:       oBlock.CreatedAt.UnixMilli(),
		OldStateRoot:    oldStateRoot,
		NewStateRoot:    newStateRoot,
		BlockCommitment: common.FromHex(oBlock.BlockCommitment),
	}
	for i := 0; i < len(cryptoTxs); i++ {
		cBlock.Txs = append(cBlock.Txs, cryptoTxs[i])
	}
	return cBlock, nil
}
