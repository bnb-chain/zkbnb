package desertexit

import (
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb/tools/desertexit/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"io/ioutil"
	"os"

	"github.com/bnb-chain/zkbnb/core"
)

const DefaultProofFolder = "./tools/desertexit/proofdata/"

type GenerateProof struct {
	running bool
	config  *config.Config
	bc      *core.BlockChain
	pool    *ants.Pool
}

func NewGenerateProof(config *config.Config) (*GenerateProof, error) {
	bc, err := core.NewBlockChainForDesertExit(config)
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	pool, err := ants.NewPool(50, ants.WithPanicHandler(func(p interface{}) {
		panic("worker exits from a panic")
	}))

	if config.ProofFolder == "" {
		config.ProofFolder = DefaultProofFolder
	}
	desertExit := &GenerateProof{
		running: true,
		config:  config,
		bc:      bc,
		pool:    pool,
	}
	return desertExit, nil
}

func (c *GenerateProof) GenerateProof(blockHeight int64, txType uint8) error {
	accountInfo, err := c.bc.AccountModel.GetAccountByL1Address(c.config.Address)
	if err != nil {
		logx.Errorf("get account by address error L1Address=%s,%v,", c.config.Address, err)
		return err
	}
	witness, pubData, err := c.generateWitness(blockHeight, accountInfo.AccountIndex, c.config.NftIndex, c.config.Token, txType)
	if err != nil {
		return err
	}

	logx.Info("generate witness successfully")

	proofs, err := c.proveDesert(witness)
	if err != nil {
		return err
	}

	storedBlockInfo, err := c.getStoredBlockInfo()
	if err != nil {
		return err
	}
	performDesertData := PerformDesertAssetData{}
	performDesertData.StoredBlockInfo = storedBlockInfo
	performDesertData.PubData = common.Bytes2Hex(pubData)
	performDesertData.Proofs = proofs

	data, err := json.Marshal(performDesertData)
	if err != nil {
		return err
	}
	mkdir(c.config.ProofFolder)
	err = ioutil.WriteFile(c.config.ProofFolder+"performDesert.json", data, 0777)
	if err != nil {
		return err
	}

	return nil
}

func (c *GenerateProof) getStoredBlockInfo() (*StoredBlockInfo, error) {
	m, err := NewDesertExit(c.config)
	if err != nil {
		return nil, err
	}
	desertExitBlock, err := c.bc.DB().DesertExitBlockModel.GetLatestExecutedBlock()
	if err != nil {
		logx.Errorf("get desert exit block failed: %s", err)
		return nil, err
	}

	lastStoredBlockInfo, err := m.getLastStoredBlockInfo(desertExitBlock.VerifiedTxHash, desertExitBlock.BlockHeight)
	if err != nil {
		logx.Errorf("get last stored block info failed: %s", err)
		return nil, err
	}

	storedBlockInfo := &StoredBlockInfo{
		BlockSize:                    lastStoredBlockInfo.BlockSize,
		BlockNumber:                  lastStoredBlockInfo.BlockNumber,
		PriorityOperations:           lastStoredBlockInfo.PriorityOperations,
		PendingOnchainOperationsHash: common.Bytes2Hex(lastStoredBlockInfo.PendingOnchainOperationsHash[:]),
		Timestamp:                    lastStoredBlockInfo.Timestamp.Int64(),
		StateRoot:                    common.Bytes2Hex(lastStoredBlockInfo.StateRoot[:]),
		Commitment:                   common.Bytes2Hex(lastStoredBlockInfo.Commitment[:]),
	}
	return storedBlockInfo, nil
}

func mkdir(dir string) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		logx.Errorf("make dir error,%s", err)
	}
}

func (c *GenerateProof) Shutdown() {
	c.running = false
	c.bc.Statedb.Close()
	c.bc.ChainDB.Close()
}
