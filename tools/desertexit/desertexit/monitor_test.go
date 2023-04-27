package desertexit

import (
	"context"
	"encoding/json"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb/common/abicoder"
	monitor2 "github.com/bnb-chain/zkbnb/common/monitor"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"testing"
)

const networkRpc = "https://bsc-testnet.nodereal.io/v1/a1cee760ac744f449416a711f20d99dd"
const hash = "0xa4d20fae8559728f347d2202363a01051060ad56b03516b7fff88e877efcb30f"

func TestDesertExit_getCommitBlocksCallData(t *testing.T) {
	client, err := rpc.NewClient(networkRpc)
	if err != nil {
		logx.Severef("failed to create rpc client, %v", err)
		return
	}
	newABIDecoder := abicoder.NewABIDecoder(monitor2.ZkBNBContractAbi)
	transaction, _, err := client.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		logx.Severe(err)
		return
	}

	receipt, err := client.GetTransactionReceipt(hash)
	if err != nil {
		logx.Errorf("query transaction receipt %s failed, err: %v", hash, err)
	} else {
		json, _ := receipt.MarshalJSON()
		logx.Infof(string(json))
	}

	storageStoredBlockInfo := StorageStoredBlockInfo{}
	newBlocksData := make([]ZkBNBCommitBlockInfo, 0)
	callData := CommitBlocksCallData{LastCommittedBlockData: &storageStoredBlockInfo, NewBlocksData: newBlocksData}
	if err := newABIDecoder.UnpackIntoInterface(&callData, "commitBlocks", transaction.Data()[4:]); err != nil {
		logx.Severe(err)
		return
	}
	jsonBytes, err := json.Marshal(callData)
	logx.Infof("callData=%s", string(jsonBytes))
}

func TestDesertExit_getRevertBlocksCallData(t *testing.T) {
	client, err := rpc.NewClient(networkRpc)
	if err != nil {
		logx.Severef("failed to create rpc client, %v", err)
		return
	}
	newABIDecoder := abicoder.NewABIDecoder(monitor2.ZkBNBContractAbi)
	transaction, _, err := client.TransactionByHash(context.Background(), common.HexToHash(hash))
	if err != nil {
		logx.Severe(err)
		return
	}

	receipt, err := client.GetTransactionReceipt(hash)
	if err != nil {
		logx.Errorf("query transaction receipt %s failed, err: %v", hash, err)
	} else {
		json, _ := receipt.MarshalJSON()
		logx.Infof(string(json))
	}

	blocksToRevertData := make([]StorageStoredBlockInfo, 0)
	callData := RevertBlocksCallData{BlocksToRevert: blocksToRevertData}
	if err := newABIDecoder.UnpackIntoInterface(&callData, "revertBlocks", transaction.Data()[4:]); err != nil {
		logx.Severe(err)
		return
	}
	jsonBytes, err := json.Marshal(callData)
	logx.Infof("callData=%s", string(jsonBytes))
}

type RevertBlocksCallData struct {
	BlocksToRevert []StorageStoredBlockInfo `abi:"_blocksToRevert"`
}
