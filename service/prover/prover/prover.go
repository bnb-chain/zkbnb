package prover

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bnb-chain/zkbas/service/prover/config"

	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/blockwitness"
	"github.com/bnb-chain/zkbas/common/model/proof"
	"github.com/bnb-chain/zkbas/common/util"
	lockUtil "github.com/bnb-chain/zkbas/common/util/globalmapHandler"
)

type Prover struct {
	Config config.Config

	RedisConn *redis.Redis

	ProofSenderModel  proof.ProofModel
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
func NewProver(c config.Config) *Prover {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %s", err.Error())
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))
	prover := &Prover{
		Config:            c,
		RedisConn:         redisConn,
		BlockWitnessModel: blockwitness.NewBlockWitnessModel(conn, c.CacheRedis, gormPointer),
		ProofSenderModel:  proof.NewProofModel(gormPointer),
	}

	prover.OptionalBlockSizes = c.BlockConfig.OptionalBlockSizes
	prover.ProvingKeys = make([]groth16.ProvingKey, len(prover.OptionalBlockSizes))
	prover.VerifyingKeys = make([]groth16.VerifyingKey, len(prover.OptionalBlockSizes))
	prover.R1cs = make([]frontend.CompiledConstraintSystem, len(prover.OptionalBlockSizes))
	for i := 0; i < len(prover.OptionalBlockSizes); i++ {
		var circuit block.BlockConstraints
		circuit.TxsCount = prover.OptionalBlockSizes[i]
		circuit.Txs = make([]block.TxConstraints, circuit.TxsCount)
		for i := 0; i < circuit.TxsCount; i++ {
			circuit.Txs[i] = block.GetZeroTxConstraint()
		}
		logx.Infof("start compile block size %d circuit", circuit.TxsCount)
		prover.R1cs[i], err = frontend.Compile(ecc.BN254, r1cs.NewBuilder, &circuit, frontend.IgnoreUnconstrainedInputs())
		if err != nil {
			panic("r1cs init error")
		}
		logx.Infof("circuit constraints: %d", prover.R1cs[i].GetNbConstraints())
		logx.Info("finish compile circuit")
		// read proving and verifying keys
		prover.ProvingKeys[i], err = util.LoadProvingKey(c.KeyPath.ProvingKeyPath[i])
		if err != nil {
			panic("provingKey loading error")
		}
		prover.VerifyingKeys[i], err = util.LoadVerifyingKey(c.KeyPath.VerifyingKeyPath[i])
		if err != nil {
			panic("verifyingKey loading error")
		}
	}

	return prover
}

func (p *Prover) ProveBlock() error {
	lock := lockUtil.GetRedisLockByKey(p.RedisConn, RedisLockKey)
	err := lockUtil.TryAcquireLock(lock)
	if err != nil {
		return fmt.Errorf("acquire lock error, err=%s", err.Error())
	}
	defer lock.Release()

	// fetch unproved block
	blockWitness, err := p.BlockWitnessModel.GetBlockWitnessByMode(util.CooMode)
	if err != nil {
		return fmt.Errorf("GetUnprovedBlock Error: err: %v", err)
	}
	// update status of block
	err = p.BlockWitnessModel.UpdateBlockWitnessStatus(blockWitness, blockwitness.StatusReceived)
	if err != nil {
		return fmt.Errorf("update block status error, err=%v", err)
	}

	// parse CryptoBlock
	var cryptoBlock *block.Block
	err = json.Unmarshal([]byte(blockWitness.WitnessData), &cryptoBlock)
	if err != nil {
		return errors.New("json.Unmarshal Error")
	}

	var keyIndex int
	for ; keyIndex < len(p.OptionalBlockSizes); keyIndex++ {
		if len(cryptoBlock.Txs) == p.OptionalBlockSizes[keyIndex] {
			break
		}
	}
	if keyIndex == len(p.OptionalBlockSizes) {
		logx.Errorf("Can't find correct vk/pk")
		return err
	}

	// Generate Proof
	blockProof, err := util.GenerateProof(p.R1cs[keyIndex], p.ProvingKeys[keyIndex], p.VerifyingKeys[keyIndex], cryptoBlock)
	if err != nil {
		return errors.New("GenerateProof Error")
	}

	formattedProof, err := util.FormatProof(blockProof, cryptoBlock.OldStateRoot, cryptoBlock.NewStateRoot, cryptoBlock.BlockCommitment)
	if err != nil {
		logx.Errorf("unable to format blockProof: %v", err)
		return err
	}

	// marshal formattedProof
	proofBytes, err := json.Marshal(formattedProof)
	if err != nil {
		logx.Errorf("formattedProof json.Marshal error: %v", err)
		return err
	}

	// check the existence of blockProof
	_, err = p.ProofSenderModel.GetProofByBlockNumber(blockWitness.Height)
	if err == nil {
		return fmt.Errorf("blockProof of current height exists")
	}

	var row = &proof.Proof{
		ProofInfo:   string(proofBytes),
		BlockNumber: blockWitness.Height,
		Status:      proof.NotSent,
	}
	err = p.ProofSenderModel.CreateProof(row)
	if err != nil {
		_ = p.BlockWitnessModel.UpdateBlockWitnessStatus(blockWitness, blockwitness.StatusPublished)
		return fmt.Errorf("create blockProof error, err=%v", err)
	}
	return nil
}
