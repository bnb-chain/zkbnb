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
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/zecrey-labs/zecrey-eth-rpc/_rpc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/logic"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/server"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/l2BlockMonitor/l2BlockMonitor"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"testing"
)

var configFile = flag.String("f",
	"D:\\Projects\\mygo\\src\\Zecrey\\SherLzp\\zecrey\\service\\rpc\\l2BlockMonitor\\etc\\l2blockmonitor.yaml", "the config file")

func TestL2BlockMonitor(t *testing.T) {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)
	srv := server.NewL2BlockMonitorServer(ctx)

	// new cron
	cronjob := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	BSCNetworkRpc, err := ctx.SysConfig.GetSysconfigByName(c.ChainConfig.BSCNetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[blockMonitor] fatal error, cannot fetch BSC NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.BSCNetworkRPCSysConfigName)
		panic(err)
	}

	bscCli, err := _rpc.NewClient(BSCNetworkRpc.Value)
	if err != nil {
		panic(err)
	}
	_, err = cronjob.AddFunc("@every 10s", func() {
		logx.Info("========================= start monitor blocks =========================")
		err := logic.MonitorL2BlockEvents(
			bscCli,
			c.ChainConfig.BSCPendingBlocksCount,
			ctx.Mempool,
			ctx.AccountAssetModel, ctx.AccountLiquidityModel, nil,
			ctx.AccountAssetHistoryModel, ctx.AccountLiquidityHistoryModel, nil,
			ctx.Block,
			ctx.L1TxSender,
		)
		if err != nil {
			logx.Error("[logic.MonitorBlocks main] unable to run:", err)
		}
		logx.Info("========================= end monitor blocks =========================")
	})
	if err != nil {
		panic(err)
	}
	cronjob.Start()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		l2BlockMonitor.RegisterL2BlockMonitorServer(grpcServer, srv)
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
