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
 */

package logic

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mathx"

	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/common/proverUtil"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/internal/svc"
)

func InitUnprovedList(
	accountTree *tree.Tree,
	assetTrees *[]*tree.Tree,
	liquidityTree *tree.Tree,
	nftTree *tree.Tree,
	ctx *svc.ServiceContext,
	initHeight int64,
) (err error) {
	proofEndHeight, err := ctx.ProofSenderModel.GetProofStartBlockNumber()
	if err != nil {
		if err == proofSender.ErrNotFound {
			if initHeight == 0 {
				return nil
			} else {
				return errors.New("[InitUnprovedList] proof not found but initHeight is not zero")
			}
		} else {
			return err
		}
	}

	// handle the proof between initHeight and proofEnd

	// get last handled block info in range
	blocks, err := ctx.BlockModel.GetBlocksBetween(initHeight+1, proofEndHeight)
	if err != nil {
		if err == block.ErrNotFound {
			return nil
		}
		return err
	}
	// lock UnprovedBlockMap
	M.Lock()
	defer M.Unlock()

	// scan each block
	for _, oBlock := range blocks {
		var (
			oldStateRoot []byte
			newStateRoot []byte
			isFirst      bool
		)
		var (
			cryptoTxs []*CryptoTx
		)
		// scan each transaction
		for _, oTx := range oBlock.Txs {
			var (
				cryptoTx *CryptoTx
			)
			cryptoTx, err = proverUtil.ConstructCryptoTx(oTx, accountTree, assetTrees, liquidityTree, nftTree, ctx.AccountModel)
			if err != nil {
				logx.Errorf("[prover] unable to construct crypto tx: %s", err.Error())
				return err
			}
			if !isFirst {
				oldStateRoot = cryptoTx.StateRootBefore
				isFirst = true
			}
			newStateRoot = cryptoTx.StateRootAfter
			cryptoTxs = append(cryptoTxs, cryptoTx)
		}
		emptyTxCount := cryptoBlock.NbTxsPerBlock - len(oBlock.Txs)
		for i := 0; i < emptyTxCount; i++ {
			cryptoTxs = append(cryptoTxs, cryptoBlock.EmptyTx())
		}
		// Check if the block is already in proofSender Table
		_, err := ctx.ProofSenderModel.GetProofByBlockNumber(oBlock.BlockHeight)
		if err != nil {
			if err == proofSender.ErrNotFound { // no proof in table, keep inserting the new block
				logx.Info(oBlock.BlockCommitment)
				if common.Bytes2Hex(newStateRoot) != oBlock.StateRoot {
					return errors.New("state root doesn't match")
				} else {
					blockInfo, err := proverUtil.BlockToCryptoBlock(oBlock, oldStateRoot, newStateRoot, cryptoTxs)
					if err != nil {
						logx.Errorf("[prover] unable to convert block to crypto block")
						return err
					}
					var nCryptoBlockInfo = &CryptoBlockInfo{
						BlockInfo: blockInfo,
						Status:    PUBLISHED,
					}
					logx.Info("new root:", common.Bytes2Hex(nCryptoBlockInfo.BlockInfo.NewStateRoot))
					logx.Info("BlockCommitment:", common.Bytes2Hex(nCryptoBlockInfo.BlockInfo.BlockCommitment))
					// insert crypto blocks array
					UnProvedCryptoBlocks = append(UnProvedCryptoBlocks, nCryptoBlockInfo)
				}
			} else {
				logx.Errorf("[InitUnprovedList] GetProofByBlockNumber error: %s", err.Error())
			}
		}

	}
	// after the init the tree status will be updated to proofEndHeight,
	// the UnprovedList will be updated to proofEndHeight too.
	return nil
}

func HandleCryptoBlock(
	accountTree *tree.Tree,
	assetTrees *[]*tree.Tree,
	liquidityTree *tree.Tree,
	nftTree *tree.Tree,
	ctx *svc.ServiceContext,
	deltaHeight int64,
) error {
	var blocks []*block.Block

	proofStart, err := ctx.ProofSenderModel.GetProofStartBlockNumber()
	if err != nil {
		if err == proofSender.ErrNotFound {
			proofStart = 0
		} else {
			return err
		}
	}
	var start = mathx.MaxInt(int(GetLatestUnprovedBlockHeight()), int(proofStart))
	// get last handled block info
	blocks, err = ctx.BlockModel.GetBlocksBetween(int64(start+1), int64(start)+deltaHeight)
	if err != nil {
		return err
	}

	// lock UnprovedBlockMap
	M.Lock()
	defer M.Unlock()
	// scan each block
	for _, oBlock := range blocks {
		var (
			oldStateRoot    []byte
			newStateRoot    []byte
			blockCommitment []byte
			isFirst         bool
		)
		var (
			cryptoTxs []*CryptoTx
		)
		// scan each transaction
		for _, oTx := range oBlock.Txs {
			var (
				cryptoTx *CryptoTx
			)
			cryptoTx, err = proverUtil.ConstructCryptoTx(oTx, accountTree, assetTrees, liquidityTree, nftTree, ctx.AccountModel)
			if err != nil {
				logx.Errorf("[prover] unable to construct crypto tx: %s", err.Error())
				return err
			}
			if !isFirst {
				oldStateRoot = cryptoTx.StateRootBefore
				isFirst = true
			}
			newStateRoot = cryptoTx.StateRootAfter
			//cryptoTxTypeBytes, err := json.Marshal(cryptoTx)
			//if err != nil {
			//	return errors.New("json.Marshal(cryptoTx) error")
			//}
			//logx.Info(string(cryptoTxTypeBytes))
			cryptoTxs = append(cryptoTxs, cryptoTx)
			logx.Info("after state root:", common.Bytes2Hex(newStateRoot))
		}
		emptyTxCount := cryptoBlock.NbTxsPerBlock - len(oBlock.Txs)
		for i := 0; i < emptyTxCount; i++ {
			cryptoTxs = append(cryptoTxs, cryptoBlock.EmptyTx())
		}
		blockCommitment = common.FromHex(oBlock.BlockCommitment)
		if common.Bytes2Hex(newStateRoot) != oBlock.StateRoot {
			logx.Info("error: new root:", common.Bytes2Hex(newStateRoot))
			logx.Info("error: BlockCommitment:", common.Bytes2Hex(blockCommitment))
			return errors.New("state root doesn't match")
		} else {
			blockInfo, err := proverUtil.BlockToCryptoBlock(oBlock, oldStateRoot, newStateRoot, cryptoTxs)
			if err != nil {
				logx.Errorf("[prover] unable to convert block to crypto block")
				return err
			}
			//if blockInfo.BlockNumber == 14 {
			//	infoBytes, _ := json.Marshal(blockInfo)
			//	fmt.Println(string(infoBytes))
			//}
			var nCryptoBlockInfo = &CryptoBlockInfo{
				BlockInfo: blockInfo,
				Status:    PUBLISHED,
			}
			logx.Info("new root:", common.Bytes2Hex(nCryptoBlockInfo.BlockInfo.NewStateRoot))
			logx.Info("BlockCommitment:", common.Bytes2Hex(nCryptoBlockInfo.BlockInfo.BlockCommitment))
			// insert crypto blocks array
			UnProvedCryptoBlocks = append(UnProvedCryptoBlocks, nCryptoBlockInfo)
		}
	}

	return nil
}
