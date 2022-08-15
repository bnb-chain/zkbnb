package witness

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	smt "github.com/bnb-chain/bas-smt"
	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockwitness"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/proof"
	utils "github.com/bnb-chain/zkbas/common/proverUtil"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/common/treedb"
	"github.com/bnb-chain/zkbas/service/cronjob/witness/config"
)

const (
	UnprovedBlockWitnessTimeout = 10 * time.Minute

	BlockProcessDelta = 10
)

type Witness struct {
	// config
	config config.Config
	helper *utils.WitnessHelper

	// Trees
	treeCtx       *treedb.Context
	accountTree   smt.SparseMerkleTree
	assetTrees    []smt.SparseMerkleTree
	liquidityTree smt.SparseMerkleTree
	nftTree       smt.SparseMerkleTree

	// The data access object
	blockModel            block.BlockModel
	accountModel          account.AccountModel
	accountHistoryModel   account.AccountHistoryModel
	liquidityHistoryModel liquidity.LiquidityHistoryModel
	nftHistoryModel       nft.L2NftHistoryModel
	proofModel            proof.ProofModel
	blockWitnessModel     blockwitness.BlockWitnessModel
}

func NewWitness(c config.Config) *Witness {
	datasource := c.Postgres.DataSource
	dbInstance, err := gorm.Open(postgres.Open(datasource))
	if err != nil {
		logx.Errorf("gorm connect db error, err: %v", err)
	}
	conn := sqlx.NewSqlConn("postgres", datasource)

	w := &Witness{
		config:                c,
		blockModel:            block.NewBlockModel(conn, c.CacheRedis, dbInstance),
		blockWitnessModel:    blockwitness.NewBlockWitnessModel(conn, c.CacheRedis, dbInstance),
		accountModel:          account.NewAccountModel(conn, c.CacheRedis, dbInstance),
		accountHistoryModel:   account.NewAccountHistoryModel(conn, c.CacheRedis, dbInstance),
		liquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, c.CacheRedis, dbInstance),
		nftHistoryModel:       nft.NewL2NftHistoryModel(conn, c.CacheRedis, dbInstance),
		proofModel:            proof.NewProofModel(dbInstance),
	}
	w.initState()
	return w
}

func (w *Witness) initState() {
	p, err := w.proofModel.GetLatestConfirmedProof()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			logx.Error("=> GetLatestConfirmedProof error:", err)
			return
		} else {
			p = &proof.Proof{
				BlockNumber: 0,
			}
		}
	}
	// init tree database
	treeCtx := &treedb.Context{
		Name:          "witness",
		Driver:        w.config.TreeDB.Driver,
		LevelDBOption: &w.config.TreeDB.LevelDBOption,
		RedisDBOption: &w.config.TreeDB.RedisDBOption,
	}
	err = treedb.SetupTreeDB(treeCtx)
	if err != nil {
		panic(fmt.Sprintf("Init tree database failed %v", err))
	}
	// init accountTree and accountStateTrees
	// the init block number use the latest sent block
	w.accountTree, w.assetTrees, err = tree.InitAccountTree(
		w.accountModel,
		w.accountHistoryModel,
		p.BlockNumber,
		treeCtx,
	)
	// the blockHeight depends on the proof start position
	if err != nil {
		logx.Errorf("InitMerkleTree error: %v", err)
		return
	}

	w.liquidityTree, err = tree.InitLiquidityTree(w.liquidityHistoryModel, p.BlockNumber,
		treeCtx)
	if err != nil {
		logx.Errorf("InitLiquidityTree error: %v", err)
		return
	}
	w.nftTree, err = tree.InitNftTree(w.nftHistoryModel, p.BlockNumber,
		treeCtx)
	if err != nil {
		logx.Errorf("InitNftTree error: %v", err)
		return
	}
	w.helper = utils.NewWitnessHelper(w.treeCtx, w.accountTree, w.liquidityTree, w.nftTree, &w.assetTrees, w.accountModel)
}

func (w *Witness) GenerateBlockWitness() (err error) {
	var latestWitnessHeight int64
	latestWitnessHeight, err = w.blockWitnessModel.GetLatestBlockWitnessHeight()
	if err != nil && err != errorcode.DbErrNotFound {
		return err
	}
	// get next batch of blocks
	blocks, err := w.blockModel.GetBlocksBetween(latestWitnessHeight+1, latestWitnessHeight+BlockProcessDelta)
	if err != nil {
		return err
	}
	// get latestVerifiedBlockNr
	latestVerifiedBlockNr, err := w.blockModel.GetLatestVerifiedBlockHeight()
	if err != nil {
		return err
	}

	// scan each block
	for _, block := range blocks {
		// Step1: construct witness
		blockWitness, err := w.constructBlockWitness(block, latestVerifiedBlockNr)
		if err != nil {
			logx.Errorf("failed to construct block witness, err: %v", err)
			return err
		}
		// Step2: commit trees for witness
		err = tree.CommitTrees(uint64(latestVerifiedBlockNr), w.accountTree, &w.assetTrees, w.liquidityTree, w.nftTree)
		if err != nil {
			logx.Errorf("unable to commit trees after txs is executed, error: %v", err)
			return err
		}
		// Step3: insert witness into database
		err = w.blockWitnessModel.CreateBlockWitness(blockWitness)
		if err != nil {
			// rollback trees
			rollBackErr := tree.RollBackTrees(uint64(block.BlockHeight)-1, w.accountTree, &w.assetTrees, w.liquidityTree, w.nftTree)
			if rollBackErr != nil {
				logx.Errorf("unable to rollback trees %v", rollBackErr)
			}
			logx.Errorf("create unproved crypto block error, err: %v", err)
			return err
		}
	}
	return nil
}

func (w *Witness) RescheduleBlockWitness() error {
	latestConfirmedProof, err := w.proofModel.GetLatestConfirmedProof()
	if err != nil && err != errorcode.DbErrNotFound {
		return err
	}

	var nextBlockNumber int64 = 1
	if err != errorcode.DbErrNotFound {
		nextBlockNumber = latestConfirmedProof.BlockNumber + 1
	}

	nextBlockWitness, err := w.blockWitnessModel.GetBlockWitnessByNumber(nextBlockNumber)
	if err != nil {
		return err
	}

	// skip if next block is not processed
	if nextBlockWitness.Status == blockwitness.StatusPublished {
		return nil
	}

	// skip if the next block proof exists
	// if the proof is not submitted and verified in L1, there should be another alerts
	_, err = w.proofModel.GetProofByBlockNumber(nextBlockNumber)
	if err == nil {
		return nil
	}

	// update block status to Published if it's timeout
	if time.Now().After(nextBlockWitness.UpdatedAt.Add(UnprovedBlockWitnessTimeout)) {
		err := w.blockWitnessModel.UpdateBlockWitnessStatus(nextBlockWitness, blockwitness.StatusPublished)
		if err != nil {
			logx.Errorf("update unproved block status error, err: %v", err)
			return err
		}
	}
	return nil
}

func (w *Witness) constructBlockWitness(block *block.Block, latestVerifiedBlockNr int64) (*blockwitness.BlockWitness, error) {
	var oldStateRoot, newStateRoot []byte
	cryptoTxs := make([]*cryptoBlock.Tx, 0, block.BlockSize)
	// scan each transaction
	for idx, tx := range block.Txs {
		cryptoTx, err := w.helper.ConstructCryptoTx(tx, uint64(latestVerifiedBlockNr))
		if err != nil {
			return nil, err
		}
		cryptoTxs = append(cryptoTxs, cryptoTx)
		// if it is the first tx of the block
		if idx == 0 {
			oldStateRoot = cryptoTx.StateRootBefore
		}
		// if it is the last tx of the block
		if idx == len(block.Txs)-1 {
			newStateRoot = cryptoTx.StateRootAfter
		}
	}

	emptyTxCount := int(block.BlockSize) - len(block.Txs)
	for i := 0; i < emptyTxCount; i++ {
		cryptoTxs = append(cryptoTxs, cryptoBlock.EmptyTx())
	}
	if common.Bytes2Hex(newStateRoot) != block.StateRoot {
		logx.Errorf("block %d state root mismatch, expect %s, get %s", block.BlockHeight,
			common.Bytes2Hex(newStateRoot), block.StateRoot)
		return nil, errors.New("state root doesn't match")
	}

	b := &cryptoBlock.Block{
		BlockNumber:     block.BlockHeight,
		CreatedAt:       block.CreatedAt.UnixMilli(),
		OldStateRoot:    oldStateRoot,
		NewStateRoot:    newStateRoot,
		BlockCommitment: common.FromHex(block.BlockCommitment),
		Txs:             cryptoTxs,
	}
	bz, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	blockWitness := blockwitness.BlockWitness{
		Height:      block.BlockHeight,
		WitnessData: string(bz),
		Status:      blockwitness.StatusPublished,
	}
	return &blockWitness, nil
}
