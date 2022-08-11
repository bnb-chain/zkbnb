/*
 * Copyright © 2021 Zkbas Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sender

import (
	"context"
	"math/big"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/l1RollupTx"
	"github.com/bnb-chain/zkbas/common/model/proof"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/service/cronjob/sender/config"
)

const (
	PendingStatus          = l1RollupTx.StatusPending
	CommitTxType           = l1RollupTx.TxTypeCommit
	VerifyAndExecuteTxType = l1RollupTx.TxTypeVerifyAndExecute
)

type Sender struct {
	Config config.Config

	// Client
	cli           *ProviderClient
	authCli       *AuthClient
	zkbasInstance *Zkbas

	// Data access objects
	blockModel          block.BlockModel
	blockForCommitModel blockForCommit.BlockForCommitModel
	l1RollupTxModel     l1RollupTx.L1RollupTxModel
	sysConfigModel      sysconfig.SysConfigModel
	proofModel          proof.ProofModel
}

func WithRedis(redisType string, redisPass string) redis.Option {
	return func(p *redis.Redis) {
		p.Type = redisType
		p.Pass = redisPass
	}
}

func NewSender(c config.Config) *Sender {
	gormPointer, err := gorm.Open(postgres.Open(c.Postgres.DataSource))
	if err != nil {
		logx.Errorf("gorm connect db error, err = %v", err)
	}
	conn := sqlx.NewSqlConn("postgres", c.Postgres.DataSource)
	redisConn := redis.New(c.CacheRedis[0].Host, WithRedis(c.CacheRedis[0].Type, c.CacheRedis[0].Pass))

	s := &Sender{
		Config:              c,
		blockModel:          block.NewBlockModel(conn, c.CacheRedis, gormPointer, redisConn),
		blockForCommitModel: blockForCommit.NewBlockForCommitModel(conn, c.CacheRedis, gormPointer),
		l1RollupTxModel:     l1RollupTx.NewL1RollupTxModel(conn, c.CacheRedis, gormPointer),
		sysConfigModel:      sysconfig.NewSysConfigModel(conn, c.CacheRedis, gormPointer),
		proofModel:          proof.NewProofModel(gormPointer),
	}

	l1RPCEndpoint, err := s.sysConfigModel.GetSysConfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch l1RPCEndpoint from sysConfig, err: %v, SysConfigName: %s",
			err, c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}
	rollupAddress, err := s.sysConfigModel.GetSysConfigByName(sysConfigName.ZkbasContract)
	if err != nil {
		logx.Severef("[sender] fatal error, cannot fetch rollupAddress from sysConfig, err: %v, SysConfigName: %s",
			err, sysConfigName.ZkbasContract)
		panic(err)
	}

	s.cli, err = _rpc.NewClient(l1RPCEndpoint.Value)
	if err != nil {
		panic(err)
	}
	var chainId *big.Int
	if c.ChainConfig.L1ChainId == "" {
		chainId, err = s.cli.ChainID(context.Background())
		if err != nil {
			panic(err)
		}
	} else {
		var (
			isValid bool
		)
		chainId, isValid = new(big.Int).SetString(c.ChainConfig.L1ChainId, 10)
		if !isValid {
			panic("invalid l1 chain id")
		}
	}

	s.authCli, err = _rpc.NewAuthClient(s.cli, c.ChainConfig.Sk, chainId)
	if err != nil {
		panic(err)
	}
	s.zkbasInstance, err = zkbas.LoadZkbasInstance(s.cli, rollupAddress.Value)
	if err != nil {
		panic(err)
	}
	return s
}
