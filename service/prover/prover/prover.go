package prover

import (
	"encoding/json"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb-crypto/circuit"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/common/redislock"
	"github.com/bnb-chain/zkbnb/dao/blockwitness"
	"github.com/bnb-chain/zkbnb/dao/proof"
	"github.com/bnb-chain/zkbnb/service/prover/config"
	"github.com/bnb-chain/zkbnb/types"
)

type Prover struct {
	Config config.Config

	RedisConn *redis.Redis

	DB                *gorm.DB
	ProofModel        proof.ProofModel
	BlockWitnessModel blockwitness.BlockWitnessModel

	VerifyingKeys      []groth16.VerifyingKey
	ProvingKeys        []groth16.ProvingKey
	OptionalBlockSizes []int
	R1cs               []frontend.CompiledConstraintSystem
}

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

func NewProver(c config.Config) *Prover {
	db, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	prover := &Prover{
		Config:            c,
		RedisConn:         redisConn,
		DB:                db,
		BlockWitnessModel: blockwitness.NewBlockWitnessModel(db),
		ProofModel:        proof.NewProofModel(db),
	}

	if !IsBlockSizesSorted(c.BlockConfig.OptionalBlockSizes) {
		panic("invalid OptionalBlockSizes")
	}

	prover.OptionalBlockSizes = c.BlockConfig.OptionalBlockSizes
	prover.ProvingKeys = make([]groth16.ProvingKey, len(prover.OptionalBlockSizes))
	prover.VerifyingKeys = make([]groth16.VerifyingKey, len(prover.OptionalBlockSizes))
	prover.R1cs = make([]frontend.CompiledConstraintSystem, len(prover.OptionalBlockSizes))
	for i := 0; i < len(prover.OptionalBlockSizes); i++ {
		var blockConstraints circuit.BlockConstraints
		blockConstraints.TxsCount = prover.OptionalBlockSizes[i]
		blockConstraints.Txs = make([]circuit.TxConstraints, blockConstraints.TxsCount)
		for i := 0; i < blockConstraints.TxsCount; i++ {
			blockConstraints.Txs[i] = circuit.GetZeroTxConstraint()
		}
		logx.Infof("start compile block size %d blockConstraints", blockConstraints.TxsCount)
		prover.R1cs[i], err = frontend.Compile(ecc.BN254, r1cs.NewBuilder, &blockConstraints, frontend.IgnoreUnconstrainedInputs())
		if err != nil {
			panic("r1cs init error")
		}
		logx.Infof("blockConstraints constraints: %d", prover.R1cs[i].GetNbConstraints())
		logx.Info("finish compile blockConstraints")
		// read proving and verifying keys
		prover.ProvingKeys[i], err = prove.LoadProvingKey(c.KeyPath.ProvingKeyPath[i])
		if err != nil {
			panic("provingKey loading error")
		}
		prover.VerifyingKeys[i], err = prove.LoadVerifyingKey(c.KeyPath.VerifyingKeyPath[i])
		if err != nil {
			panic("verifyingKey loading error")
		}
	}

	return prover
}

func (p *Prover) ProveBlock() error {
	blockWitness, err := func() (*blockwitness.BlockWitness, error) {
		lock := redislock.GetRedisLockByKey(p.RedisConn, RedisLockKey)
		err := redislock.TryAcquireLock(lock)
		if err != nil {
			return nil, err
		}
		//nolint:errcheck
		defer lock.Release()

		// Fetch unproved block witness.
		blockWitness, err := p.BlockWitnessModel.GetLatestBlockWitness()
		if err != nil {
			return nil, err
		}
		// Update status of block witness.
		err = p.BlockWitnessModel.UpdateBlockWitnessStatus(blockWitness, blockwitness.StatusReceived)
		if err != nil {
			return nil, err
		}
		return blockWitness, nil
	}()
	if err != nil {
		if err == types.DbErrNotFound {
			return nil
		}
		return err
	}
	defer func() {
		if err == nil {
			return
		}

		// Recover block witness status.
		res := p.BlockWitnessModel.UpdateBlockWitnessStatus(blockWitness, blockwitness.StatusPublished)
		if res != nil {
			logx.Errorf("revert block witness status failed, err %v", res)
		}
	}()

	// Parse crypto block.
	var cryptoBlock *circuit.Block
	err = json.Unmarshal([]byte(blockWitness.WitnessData), &cryptoBlock)
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

	// Generate proof.
	blockProof, err := prove.GenerateProof(p.R1cs[keyIndex], p.ProvingKeys[keyIndex], p.VerifyingKeys[keyIndex], cryptoBlock)
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
	return err
}

func (p *Prover) Shutdown() {
	sqlDB, err := p.DB.DB()
	if err == nil && sqlDB != nil {
		err = sqlDB.Close()
	}
	if err != nil {
		logx.Errorf("close db error: %s", err.Error())
	}
}
