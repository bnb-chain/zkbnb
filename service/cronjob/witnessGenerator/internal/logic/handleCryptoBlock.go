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
 */

package logic

import (
	"errors"
	"time"

	"github.com/bnb-chain/zkbas/errorcode"

	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/blockForProof"
	"github.com/bnb-chain/zkbas/common/proverUtil"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/service/cronjob/witnessGenerator/internal/svc"
)

func GenerateWitness(
	accountTree *tree.Tree,
	assetTrees *[]*tree.Tree,
	liquidityTree *tree.Tree,
	nftTree *tree.Tree,
	ctx *svc.ServiceContext,
	deltaHeight int64,
) {
	err := generateUnprovedBlockWitness(ctx, accountTree, assetTrees, liquidityTree, nftTree, deltaHeight)
	if err != nil {
		logx.Errorf("generate block witness error, err=%s", err.Error())
	}

	updateTimeoutUnprovedBlock(ctx)
}

func generateUnprovedBlockWitness(
	ctx *svc.ServiceContext,
	accountTree *tree.Tree,
	assetTrees *[]*tree.Tree,
	liquidityTree *tree.Tree,
	nftTree *tree.Tree,
	deltaHeight int64,
) error {
	latestUnprovedHeight, err := ctx.BlockForProofModel.GetLatestUnprovedBlockHeight()
	if err != nil {
		if err == errorcode.DbErrNotFound {
			latestUnprovedHeight = 0
		} else {
			return err
		}
	}

	// get last handled block info
	blocks, err := ctx.BlockModel.GetBlocksBetween(latestUnprovedHeight+1, latestUnprovedHeight+deltaHeight)
	if err != nil {
		return err
	}

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
			cryptoTxs = append(cryptoTxs, cryptoTx)
			logx.Info("after state root:", common.Bytes2Hex(newStateRoot))
		}
		emptyTxCount := int(oBlock.BlockSize) - len(oBlock.Txs)
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
			var nCryptoBlockInfo = &CryptoBlockInfo{
				BlockInfo: blockInfo,
				Status:    blockForProof.StatusPublished,
			}
			logx.Info("new root:", common.Bytes2Hex(nCryptoBlockInfo.BlockInfo.NewStateRoot))
			logx.Info("BlockCommitment:", common.Bytes2Hex(nCryptoBlockInfo.BlockInfo.BlockCommitment))

			// insert crypto blocks array
			unprovedCryptoBlockModel, err := CryptoBlockInfoToBlockForProof(nCryptoBlockInfo)
			if err != nil {
				logx.Errorf("[prover] marshal crypto block info error, err=%s", err.Error())
				return err
			}
			err = ctx.BlockForProofModel.CreateConsecutiveUnprovedCryptoBlock(unprovedCryptoBlockModel)
			if err != nil {
				logx.Errorf("[prover] create unproved crypto block error, err=%s", err.Error())
				return err
			}
		}
	}
	return nil
}

func updateTimeoutUnprovedBlock(ctx *svc.ServiceContext) {
	latestConfirmedProof, err := ctx.ProofSenderModel.GetLatestConfirmedProof()
	if err != nil && err != errorcode.DbErrNotFound {
		return
	}

	var nextBlockNumber int64 = 1
	if err != errorcode.DbErrNotFound {
		nextBlockNumber = latestConfirmedProof.BlockNumber + 1
	}

	nextUnprovedBlock, err := ctx.BlockForProofModel.GetUnprovedCryptoBlockByBlockNumber(nextBlockNumber)
	if err != nil {
		return
	}

	// skip if next block is not processed
	if nextUnprovedBlock.Status == blockForProof.StatusPublished {
		return
	}

	// skip if the next block proof exists
	// if the proof is not submitted and verified in L1, there should be another alerts
	_, err = ctx.ProofSenderModel.GetProofByBlockNumber(nextBlockNumber)
	if err == nil {
		return
	}

	// update block status to Published if it's timeout
	if time.Now().After(nextUnprovedBlock.UpdatedAt.Add(UnprovedBlockReceivedTimeout)) {
		err := ctx.BlockForProofModel.UpdateUnprovedCryptoBlockStatus(nextUnprovedBlock, blockForProof.StatusPublished)
		if err != nil {
			logx.Errorf("update unproved block status error, err=%s", err.Error())
			return
		}
	}
}
