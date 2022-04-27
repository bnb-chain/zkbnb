/*
 * Copyright © 2021 Zecrey Protocol
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

package test

import (
	"flag"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/tree"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/committer/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/committer/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/committer/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f",
	"/Users/gavin/Desktop/zecrey-v2/service/rpc/committer/etc/committer.yaml", "the config file")

func TestCommitter(t *testing.T) {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)

	var (
		accountTree       *tree.Tree
		accountStateTrees []*tree.AccountStateTree
	)
	// get latest account
	latestHeight, err := ctx.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		panic(err)
	}
	// init accountTree and accountStateTrees
	accountTree, accountStateTrees, err = tree.InitAccountTree(
		ctx.AccountHistoryModel,
		ctx.AccountAssetHistoryModel,
		ctx.LiquidityAssetHistoryModel,
		latestHeight,
	)
	if err != nil {
		logx.Error("[committer] => InitMerkleTree error:", err)
		return
	}

	var lastCommitTimeStamp = time.Now()

	logx.Info("========================= start committer task =========================")
	err = logic.CommitterTask(
		ctx,
		lastCommitTimeStamp,
		accountTree,
		// TODO nft tree
		nil,
		accountStateTrees,
	)
	if err != nil {
		logx.Info("[committer.CommitterTask main] unable to run:", err)
	}
	logx.Info("========================= end committer task =========================")
}
