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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/proverClient/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/proverHubProto"
	"github.com/zeromicro/go-zero/core/logx"
)

func ProveBlock(ctx *svc.ServiceContext) error {
	// fetch unproved block
	resp, err := ctx.ProverHubRPC.GetUnprovedBlock(context.Background(), &proverHubProto.ReqGetUnprovedBlock{Mode: 1})
	if err != nil || resp == nil || resp.Status == util.FailStatus {
		return errors.New(fmt.Sprintf("[ProveBlock] GetUnprovedBlock Error: err / resp : %v/%v", err, resp))
	}

	// parse CryptoBlock
	var cryptoBlock *CryptoBlock
	err = json.Unmarshal([]byte(resp.Result.BlockInfo), &cryptoBlock)
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
		log.Println("[ProveBlock] Can't find correct vk/pk")
		return err
	}

	// Generate Proof
	proof, err := util.GenerateProof(R1cs[keyIndex], ProvingKeys[keyIndex], VerifyingKeys[keyIndex], cryptoBlock)
	if err != nil {
		return errors.New("[ProveBlock] GenerateProof Error")
	}

	formattedProof, err := util.FormatProof(proof, cryptoBlock.OldStateRoot, cryptoBlock.NewStateRoot, cryptoBlock.BlockCommitment)
	if err != nil {
		log.Println("[ProveBlock] Unable to Format Proof:", err)
		return err
	}

	// marshal formattedProof
	proofBytes, err := json.Marshal(formattedProof)
	if err != nil {
		log.Println("[ProveBlock] formattedProof json.Marshal error:", err)
		return err
	}

	// marshal cryptoBlock
	BlockInfoBytes, err := json.Marshal(cryptoBlock)
	if err != nil {
		log.Println("[ProveBlock] cryptoBlock json.Marshal error:", err)
		return err
	}

	// submit proof
	submitProofRPCResp, err := ctx.ProverHubRPC.SubmitProof(context.Background(), &proverHubProto.ReqSubmitProof{
		Proof:     string(proofBytes),
		BlockInfo: string(BlockInfoBytes),
	})
	if err != nil || submitProofRPCResp == nil {
		logx.Error("ProverHubRPC.SubmitProof Error: ", err)
		return errors.New("ProverHubRPC.SubmitProof Error")
	}
	// TODO proof store locally
	return nil
}
