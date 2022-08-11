package prover

import (
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

	"github.com/bnb-chain/zkbas/common/model/blockForProof"
	"github.com/bnb-chain/zkbas/common/model/proof"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/cronjob/prover/config"
)

type Prover struct {
	Config config.Config

	RedisConn *redis.Redis

	ProofSenderModel   proof.ProofModel
	BlockForProofModel blockForProof.BlockForProofModel

	VerifyingKeys []groth16.VerifyingKey
	ProvingKeys   []groth16.ProvingKey
	KeyTxCounts   []int
	R1cs          []frontend.CompiledConstraintSystem
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
		Config:             c,
		RedisConn:          redisConn,
		BlockForProofModel: blockForProof.NewBlockForProofModel(conn, c.CacheRedis, gormPointer),
		ProofSenderModel:   proof.NewProofModel(gormPointer),
	}

	prover.KeyTxCounts = c.KeyPath.KeyTxCounts
	prover.ProvingKeys = make([]groth16.ProvingKey, len(prover.KeyTxCounts))
	prover.VerifyingKeys = make([]groth16.VerifyingKey, len(prover.KeyTxCounts))
	prover.R1cs = make([]frontend.CompiledConstraintSystem, len(prover.KeyTxCounts))
	for i := 0; i < len(prover.KeyTxCounts); i++ {
		var circuit block.BlockConstraints
		circuit.TxsCount = prover.KeyTxCounts[i]
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
