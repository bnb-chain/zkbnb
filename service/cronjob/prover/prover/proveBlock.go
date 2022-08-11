package prover

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/blockForProof"
	"github.com/bnb-chain/zkbas/common/model/proof"
	"github.com/bnb-chain/zkbas/common/util"
	lockUtil "github.com/bnb-chain/zkbas/common/util/globalmapHandler"
)

func (p *Prover) ProveBlock() error {
	lock := lockUtil.GetRedisLockByKey(p.RedisConn, RedisLockKey)
	err := lockUtil.TryAcquireLock(lock)
	if err != nil {
		return fmt.Errorf("acquire lock error, err=%s", err.Error())
	}
	defer lock.Release()

	// fetch unproved block
	unprovedBlock, err := p.BlockForProofModel.GetUnprovedCryptoBlockByMode(util.COO_MODE)
	if err != nil {
		return fmt.Errorf("[ProveBlock] GetUnprovedBlock Error: err: %v", err)
	}
	// update status of block
	err = p.BlockForProofModel.UpdateUnprovedCryptoBlockStatus(unprovedBlock, blockForProof.StatusReceived)
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
	for ; keyIndex < len(p.KeyTxCounts); keyIndex++ {
		if len(cryptoBlock.Txs) == p.KeyTxCounts[keyIndex] {
			break
		}
	}
	if keyIndex == len(p.KeyTxCounts) {
		logx.Errorf("[ProveBlock] Can't find correct vk/pk")
		return err
	}

	// Generate Proof
	blockProof, err := util.GenerateProof(p.R1cs[keyIndex], p.ProvingKeys[keyIndex], p.VerifyingKeys[keyIndex], cryptoBlock)
	if err != nil {
		return errors.New("[ProveBlock] GenerateProof Error")
	}

	formattedProof, err := util.FormatProof(blockProof, cryptoBlock.OldStateRoot, cryptoBlock.NewStateRoot, cryptoBlock.BlockCommitment)
	if err != nil {
		logx.Errorf("[ProveBlock] unable to format blockProof: %s", err.Error())
		return err
	}

	// marshal formattedProof
	proofBytes, err := json.Marshal(formattedProof)
	if err != nil {
		logx.Errorf("[ProveBlock] formattedProof json.Marshal error: %s", err.Error())
		return err
	}

	// check the existence of blockProof
	_, err = p.ProofSenderModel.GetProofByBlockNumber(unprovedBlock.BlockHeight)
	if err == nil {
		return fmt.Errorf("[ProveBlock] blockProof of current height exists")
	}

	var row = &proof.Proof{
		ProofInfo:   string(proofBytes),
		BlockNumber: unprovedBlock.BlockHeight,
		Status:      proof.NotSent,
	}
	err = p.ProofSenderModel.CreateProof(row)
	if err != nil {
		_ = p.BlockForProofModel.UpdateUnprovedCryptoBlockStatus(unprovedBlock, blockForProof.StatusPublished)
		return fmt.Errorf("[ProveBlock] create blockProof error, err=%s", err.Error())
	}
	return nil
}
