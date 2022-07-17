package logic

import (
	"encoding/json"
	"errors"

	"github.com/zecrey-labs/zecrey-legend/common/model/blockForProof"
)

func CryptoBlockInfoToBlockForProof(cryptoBlock *CryptoBlockInfo) (*BlockForProof, error) {
	if cryptoBlock == nil {
		return nil, errors.New("crypto block is nil")
	}

	blockInfo, err := json.Marshal(cryptoBlock.BlockInfo)
	if err != nil {
		return nil, err
	}

	blockModel := blockForProof.BlockForProof{
		BlockHeight: cryptoBlock.BlockInfo.BlockNumber,
		BlockData:   string(blockInfo),
		Status:      cryptoBlock.Status,
	}

	return &blockModel, nil
}
