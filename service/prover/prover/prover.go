package prover

import (
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/constraint"
	"github.com/prometheus/client_golang/prometheus"
	"runtime"
	"time"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/std"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"

	"github.com/bnb-chain/zkbnb-crypto/circuit"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/common/redislock"
	"github.com/bnb-chain/zkbnb/dao/blockwitness"
	"github.com/bnb-chain/zkbnb/dao/proof"
	"github.com/bnb-chain/zkbnb/service/prover/config"
	"github.com/bnb-chain/zkbnb/types"
)

var (
	l2ProofHeightMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2_proof_height",
		Help:      "l2_proof_height metrics.",
	})
	l2ExceptionProofHeightMetrics = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2_exception_proof_height",
		Help:      "l2_exception_proof_height metrics.",
	})
)

type Prover struct {
	running bool
	Config  config.Config

	RedisConn *redis.Redis

	DB                *gorm.DB
	ProofModel        proof.ProofModel
	BlockWitnessModel blockwitness.BlockWitnessModel

	VerifyingKeys      []groth16.VerifyingKey
	ProvingKeys        [][]groth16.ProvingKey
	OptionalBlockSizes []int
	SessionNames       []string
	R1cs               []constraint.ConstraintSystem
}

var (
	l2BlockProverGenerateHeightMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "l2Block_prover_generate_height",
		Help:      "l2Block_prover_generate metrics.",
	})

	proofGenerateTimeMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "zkbnb",
		Name:      "proof_generate_time",
		Help:      "proof_generate_time metrics.",
	})
)

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func IsBlockSizesSorted(blockSizes []int) bool {
	for i := 1; i < len(blockSizes); i++ {
		if blockSizes[i] <= blockSizes[i-1] {
			return false
		}
	}
	return true
}

func NewProver(c config.Config) (*Prover, error) {
	if err := prometheus.Register(l2ProofHeightMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2ProofHeightMetrics error: %v", err)
	}
	if err := prometheus.Register(l2ExceptionProofHeightMetrics); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2ExceptionProofHeightMetrics error: %v", err)
	}
	if err := prometheus.Register(l2BlockProverGenerateHeightMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register l2BlockProverGenerateHeightMetric error: %v", err)
	}
	if err := prometheus.Register(proofGenerateTimeMetric); err != nil {
		return nil, fmt.Errorf("prometheus.Register proofGenerateTimeMetric error: %v", err)
	}

	masterDataSource := c.Postgres.MasterDataSource
	slaveDataSource := c.Postgres.SlaveDataSource
	db, err := gorm.Open(postgres.Open(masterDataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}

	db.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{postgres.Open(masterDataSource)},
		Replicas: []gorm.Dialector{postgres.Open(slaveDataSource)},
	}))
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	prover := &Prover{
		running:           true,
		Config:            c,
		RedisConn:         redisConn,
		DB:                db,
		BlockWitnessModel: blockwitness.NewBlockWitnessModel(db),
		ProofModel:        proof.NewProofModel(db),
	}

	if !IsBlockSizesSorted(c.BlockConfig.OptionalBlockSizes) {
		logx.Severe("invalid OptionalBlockSizes")
		panic("invalid OptionalBlockSizes")
	}

	prover.OptionalBlockSizes = c.BlockConfig.OptionalBlockSizes
	prover.ProvingKeys = make([][]groth16.ProvingKey, len(prover.OptionalBlockSizes))
	prover.VerifyingKeys = make([]groth16.VerifyingKey, len(prover.OptionalBlockSizes))
	prover.R1cs = make([]constraint.ConstraintSystem, len(prover.OptionalBlockSizes))
	prover.SessionNames = make([]string, len(prover.OptionalBlockSizes))
	for i := 0; i < len(prover.OptionalBlockSizes); i++ {
		var blockConstraints circuit.BlockConstraints
		blockConstraints.TxsCount = prover.OptionalBlockSizes[i]
		blockConstraints.Txs = make([]circuit.TxConstraints, blockConstraints.TxsCount)
		for i := 0; i < blockConstraints.TxsCount; i++ {
			blockConstraints.Txs[i] = circuit.GetZeroTxConstraint()
		}
		blockConstraints.GasAssetIds = types.GasAssets[:]
		blockConstraints.GasAccountIndex = types.GasAccount
		blockConstraints.Gas = circuit.GetZeroGasConstraints(types.GasAssets[:])

		logx.Infof("start compile block size %d blockConstraints", blockConstraints.TxsCount)
		// prover.R1cs[i], err = frontend.Compile(ecc.BN254, r1cs.NewBuilder, &blockConstraints, frontend.IgnoreUnconstrainedInputs())
		// groth16.LazifyR1cs(prover.R1cs[i])
		std.RegisterHints()

		nbConstraints, err := prove.LoadR1CSLen(c.KeyPath[i] + ".r1cslen")
		if err != nil {
			logx.Severe("r1cs nb constraints read error")
			panic("r1cs nb constraints read error")
		}

		r1cs := groth16.NewCS(ecc.BN254)
		r1cs.LoadFromSplitBinaryConcurrent(c.KeyPath[i], nbConstraints, c.BlockConfig.R1CSBatchSize, runtime.NumCPU())
		prover.R1cs[i] = r1cs
		if err != nil {
			logx.Severe("r1cs init error")
			panic("r1cs init error")
		}
		logx.Infof("blockConstraints constraints: %d", prover.R1cs[i].GetNbConstraints())
		logx.Info("finish compile blockConstraints")
		// read proving and verifying keys
		prover.ProvingKeys[i], err = prove.LoadProvingKey(c.KeyPath[i])
		if err != nil {
			logx.Severe("provingKey loading error")
			panic("provingKey loading error")
		}
		prover.VerifyingKeys[i], err = prove.LoadVerifyingKey(c.KeyPath[i])
		if err != nil {
			logx.Severe("verifyingKey loading error")
			panic("verifyingKey loading error")
		}
		prover.SessionNames[i] = c.KeyPath[i]
	}

	w, err := prover.BlockWitnessModel.GetLatestReceivedBlockWitness(prover.OptionalBlockSizes)
	var wHeight int64
	if err != nil {
		if err == types.DbErrNotFound {
			wHeight = 0
		} else {
			logx.Severe("get latest receive block witness error")
			panic("get latest receive block witness error")
		}
	} else {
		wHeight = w.Height
	}
	var pHeight int64
	p, err := prover.ProofModel.GetLatestProof()
	if err != nil {
		if err == types.DbErrNotFound {
			pHeight = 0
		} else {
			logx.Severe("get latest proof error")
			panic("get latest proof error")
		}
	} else {
		pHeight = p.BlockNumber
	}
	if wHeight > pHeight {
		for i := pHeight + 1; i <= wHeight; i++ {
			err := prover.BlockWitnessModel.UpdateBlockWitnessStatusByHeight(i)
			if err != nil {
				logx.Severe("init witness status error")
				panic("init witness status error")
			}
		}
	}

	return prover, nil
}

func (p *Prover) ProveBlock() error {
	for {
		if !p.running {
			break
		}
		blockWitness, err := p.getWitness()
		if err != nil {
			if err == types.DbErrNotFound {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return err
		}

		logx.Infof("doProveBlock start, height=%d", blockWitness.Height)
		err = p.doProveBlock(blockWitness)
		if err != nil {
			logx.Severef("doProveBlock failed, err %v,height=%d", err, blockWitness.Height)
			// Recover block witness status.
			res := p.BlockWitnessModel.UpdateBlockWitnessStatus(blockWitness, blockwitness.StatusPublished)
			if res != nil {
				logx.Severef("revert block witness status failed, err %v,height=%d", res, blockWitness.Height)
				panic("recover block witness status error " + res.Error())
			}
			l2ExceptionProofHeightMetrics.Set(float64(blockWitness.Height))
			return err
		}
	}
	return nil
}

func (p *Prover) doProveBlock(blockWitness *blockwitness.BlockWitness) error {
	// Parse crypto block.
	var cryptoBlock *circuit.Block
	err := json.Unmarshal([]byte(blockWitness.WitnessData), &cryptoBlock)
	if err != nil {
		return err
	}

	var keyIndex int
	for ; keyIndex < len(p.OptionalBlockSizes); keyIndex++ {
		if len(cryptoBlock.Txs) == p.OptionalBlockSizes[keyIndex] {
			break
		}
	}
	if keyIndex == len(p.OptionalBlockSizes) {
		return fmt.Errorf("can't find correct vk/pk")
	}

	start := time.Now()
	// Generate proof.
	blockProof, err := prove.GenerateProof(p.R1cs[keyIndex], p.ProvingKeys[keyIndex], p.VerifyingKeys[keyIndex], cryptoBlock, p.SessionNames[keyIndex])
	if err != nil {
		return fmt.Errorf("failed to generateProof, err: %v", err)
	}

	formattedProof, err := prove.FormatProof(blockProof, cryptoBlock.OldStateRoot, cryptoBlock.NewStateRoot, cryptoBlock.BlockCommitment)
	if err != nil {
		return fmt.Errorf("unable to format blockProof: %v", err)
	}

	// Marshal formatted proof.
	proofBytes, err := json.Marshal(formattedProof)
	if err != nil {
		return err
	}

	// Check the existence of block proof.
	_, err = p.ProofModel.GetProofByBlockHeight(blockWitness.Height)
	if err == nil {
		logx.Errorf("blockProof of height %d exists", blockWitness.Height)
		return nil
	}

	var row = &proof.Proof{
		ProofInfo:   string(proofBytes),
		BlockNumber: blockWitness.Height,
		Status:      proof.NotSent,
	}
	err = p.ProofModel.CreateProof(row)
	proofGenerateTimeMetric.Set(float64(time.Since(start).Milliseconds()))
	l2BlockProverGenerateHeightMetric.Set(float64(blockWitness.Height))
	l2ExceptionProofHeightMetrics.Set(float64(0))
	l2ProofHeightMetrics.Set(float64(row.BlockNumber))
	return err
}

func (p *Prover) getWitness() (*blockwitness.BlockWitness, error) {
	lock := redislock.GetRedisLockByKey(p.RedisConn, RedisLockKey)
	err := redislock.TryAcquireLock(lock)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer lock.Release()

	// Fetch unproved block witness.
	blockWitness, err := p.BlockWitnessModel.GetLatestBlockWitness(p.OptionalBlockSizes)
	if err != nil {
		return nil, err
	}
	// Update status of block witness.
	err = p.BlockWitnessModel.UpdateBlockWitnessStatus(blockWitness, blockwitness.StatusReceived)
	if err != nil {
		return nil, err
	}
	return blockWitness, nil
}

func (p *Prover) Shutdown() {
	p.running = false
	sqlDB, err := p.DB.DB()
	if err == nil && sqlDB != nil {
		err = sqlDB.Close()
	}
	if err != nil {
		logx.Errorf("close db error: %s", err.Error())
	}
}
