package witness

import (
	"errors"
	"fmt"
	"time"

	smt "github.com/bnb-chain/bas-smt"
	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForProof"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/proof"
	"github.com/bnb-chain/zkbas/common/proverUtil"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/pkg/treedb"
	"github.com/bnb-chain/zkbas/service/cronjob/witness/config"
)

const (
	UnprovedBlockReceivedTimeout = 10 * time.Minute

	BlockProcessDelta = 10
)

type Witness struct {
	// Config
	config config.Config
	helper *proverUtil.WitnessHelper

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
	blockForProofModel    blockForProof.BlockForProofModel
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func NewWitness(c config.Config) *Witness {
	datasource := c.Postgres.DataSource
	dbInstance, err := gorm.Open(postgres.Open(datasource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", datasource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))

	w := &Witness{
		config:                c,
		blockModel:            block.NewBlockModel(conn, c.CacheRedis, dbInstance, redisConn),
		blockForProofModel:    blockForProof.NewBlockForProofModel(conn, c.CacheRedis, dbInstance),
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
			logx.Error("[prover] => GetLatestConfirmedProof error:", err)
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
		logx.Errorf("[prover] => InitMerkleTree error: %v", err)
		return
	}

	w.liquidityTree, err = tree.InitLiquidityTree(w.liquidityHistoryModel, p.BlockNumber,
		treeCtx)
	if err != nil {
		logx.Errorf("[prover] InitLiquidityTree error: %v", err)
		return
	}
	w.nftTree, err = tree.InitNftTree(w.nftHistoryModel, p.BlockNumber,
		treeCtx)
	if err != nil {
		logx.Errorf("[prover] InitNftTree error: %v", err)
		return
	}
	w.helper = proverUtil.NewWitnessHelper(w.treeCtx, w.accountTree, w.liquidityTree, w.nftTree, &w.assetTrees, w.accountModel)
}

func (w *Witness) GenerateBlockWitness() {
	err := w.generateUnprovedBlockWitness(BlockProcessDelta)
	if err != nil {
		logx.Errorf("generate block witness error: %v", err)
	}

	w.updateTimeoutUnprovedBlock()
}

func (w *Witness) generateUnprovedBlockWitness(deltaHeight int64) error {
	latestUnprovedHeight, err := w.blockForProofModel.GetLatestUnprovedBlockHeight()
	if err != nil {
		if err == errorcode.DbErrNotFound {
			latestUnprovedHeight = 0
		} else {
			return err
		}
	}

	// get last handled block info
	blocks, err := w.blockModel.GetBlocksBetween(latestUnprovedHeight+1, latestUnprovedHeight+deltaHeight)
	if err != nil {
		return err
	}
	// get latestVerifiedBlockNr
	latestVerifiedBlockNr, err := w.blockModel.GetLatestVerifiedBlockHeight()
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
			cryptoTx, err = w.helper.ConstructCryptoTx(oTx, uint64(latestVerifiedBlockNr))
			if err != nil {
				return err
			}
			if !isFirst {
				oldStateRoot = cryptoTx.StateRootBefore
				isFirst = true
			}
			newStateRoot = cryptoTx.StateRootAfter
			cryptoTxs = append(cryptoTxs, cryptoTx)
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
		}

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

		// commit trees
		err = tree.CommitTrees(uint64(latestVerifiedBlockNr), w.accountTree, &w.assetTrees, w.liquidityTree, w.nftTree)
		if err != nil {
			logx.Errorf("[prover] unable to commit trees after txs is executed", err.Error())
			return err
		}

		err = w.blockForProofModel.CreateConsecutiveUnprovedCryptoBlock(unprovedCryptoBlockModel)
		if err != nil {
			// rollback trees
			err = tree.RollBackTrees(uint64(oBlock.BlockHeight)-1, w.accountTree, &w.assetTrees, w.liquidityTree, w.nftTree)
			if err != nil {
				logx.Errorf("[prover] unable to rollback trees", err)
			}
			logx.Errorf("[prover] create unproved crypto block error, err=%s", err.Error())
			return err
		}

	}
	return nil
}

func (w *Witness) updateTimeoutUnprovedBlock() {
	latestConfirmedProof, err := w.proofModel.GetLatestConfirmedProof()
	if err != nil && err != errorcode.DbErrNotFound {
		return
	}

	var nextBlockNumber int64 = 1
	if err != errorcode.DbErrNotFound {
		nextBlockNumber = latestConfirmedProof.BlockNumber + 1
	}

	nextUnprovedBlock, err := w.blockForProofModel.GetUnprovedCryptoBlockByBlockNumber(nextBlockNumber)
	if err != nil {
		return
	}

	// skip if next block is not processed
	if nextUnprovedBlock.Status == blockForProof.StatusPublished {
		return
	}

	// skip if the next block proof exists
	// if the proof is not submitted and verified in L1, there should be another alerts
	_, err = w.proofModel.GetProofByBlockNumber(nextBlockNumber)
	if err == nil {
		return
	}

	// update block status to Published if it's timeout
	if time.Now().After(nextUnprovedBlock.UpdatedAt.Add(UnprovedBlockReceivedTimeout)) {
		err := w.blockForProofModel.UpdateUnprovedCryptoBlockStatus(nextUnprovedBlock, blockForProof.StatusPublished)
		if err != nil {
			logx.Errorf("update unproved block status error, err=%s", err.Error())
			return
		}
	}
}
