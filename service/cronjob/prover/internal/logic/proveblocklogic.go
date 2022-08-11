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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/blockForProof"
	"github.com/bnb-chain/zkbas/common/util"
	lockUtil "github.com/bnb-chain/zkbas/common/util/globalmapHandler"
	"github.com/bnb-chain/zkbas/service/cronjob/prover/internal/svc"
)

func ProveBlock(ctx *svc.ServiceContext) error {
	lock := lockUtil.GetRedisLockByKey(ctx.RedisConn, RedisLockKey)
	err := lockUtil.TryAcquireLock(lock)
	if err != nil {
		return fmt.Errorf("acquire lock error, err=%s", err.Error())
	}
	defer lock.Release()

	// fetch unproved block
	unprovedBlock, err := ctx.BlockForProofModel.GetUnprovedCryptoBlockByMode(util.COO_MODE)
	if err != nil {
		return fmt.Errorf("[ProveBlock] GetUnprovedBlock Error: err: %v", err)
	}
	// update status of block
	err = ctx.BlockForProofModel.UpdateUnprovedCryptoBlockStatus(unprovedBlock, blockForProof.StatusReceived)
	if err != nil {
		return fmt.Errorf("[ProveBlock] update block status error, err=%s", err.Error())
	}

	// parse CryptoBlock
	var cryptoBlock *block.Block
	err = json.Unmarshal([]byte(unprovedBlock.BlockData), &cryptoBlock)
	if err != nil {
		return errors.New("[ProveBlock] json.Unmarshal Error")
	}

	var keyIndex int
	for ; keyIndex < len(KeyTxCounts); keyIndex++ {
		if len(cryptoBlock.Txs) == KeyTxCounts[keyIndex] {
			break
		}
	}
	if keyIndex == len(KeyTxCounts) {
		logx.Errorf("[ProveBlock] Can't find correct vk/pk")
		return err
	}

	// Generate Proof
	proof, err := util.GenerateProof(R1cs[keyIndex], ProvingKeys[keyIndex], VerifyingKeys[keyIndex], cryptoBlock)
	if err != nil {
		return errors.New("[ProveBlock] GenerateProof Error")
	}

	formattedProof, err := util.FormatProof(proof, cryptoBlock.OldStateRoot, cryptoBlock.NewStateRoot, cryptoBlock.BlockCommitment)
	if err != nil {
		logx.Errorf("[ProveBlock] unable to format proof: %s", err.Error())
		return err
	}

	// marshal formattedProof
	proofBytes, err := json.Marshal(formattedProof)
	if err != nil {
		logx.Errorf("[ProveBlock] formattedProof json.Marshal error: %s", err.Error())
		return err
	}

	// check the existence of proof
	_, err = ctx.ProofSenderModel.GetProofByBlockNumber(unprovedBlock.BlockHeight)
	if err == nil {
		return fmt.Errorf("[ProveBlock] proof of current height exists")
	}

	var row = &proof.Proof{
		ProofInfo:   string(proofBytes),
		BlockNumber: unprovedBlock.BlockHeight,
		Status:      proof.NotSent,
	}
	err = ctx.ProofSenderModel.CreateProof(row)
	if err != nil {
		_ = ctx.BlockForProofModel.UpdateUnprovedCryptoBlockStatus(unprovedBlock, blockForProof.StatusPublished)
		return fmt.Errorf("[ProveBlock] create proof error, err=%s", err.Error())
	}
	return nil
}
