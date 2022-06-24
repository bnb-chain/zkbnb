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
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas-eth-rpc/_rpc"
	"github.com/bnb-chain/zkbas/service/cronjob/blockMonitor/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/blockMonitor/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/blockMonitor/internal/svc"
)

var configFile = flag.String("f",
	"D:\\Projects\\mygo\\src\\Zecrey\\zecrey\\service\\rpc\\blockMonitor\\etc\\avalanche.yaml", "the config file")

func TestBlockMonitor(t *testing.T) {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	ZkbasRollupAddress, err := ctx.SysConfig.GetSysconfigByName(c.ChainConfig.ZkbasContractAddrSysConfigName)

	if err != nil {
		logx.Severef("[blockMonitor] fatal error, cannot fetch ZkbasContractAddr from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.ZkbasContractAddrSysConfigName)
		panic(err)
	}

	NetworkRpc, err := ctx.SysConfig.GetSysconfigByName(c.ChainConfig.NetworkRPCSysConfigName)
	if err != nil {
		logx.Severef("[blockMonitor] fatal error, cannot fetch NetworkRPC from sysConfig, err: %s, SysConfigName: %s",
			err.Error(), c.ChainConfig.NetworkRPCSysConfigName)
		panic(err)
	}

	logx.Infof("[blockMonitor] ChainName: %s, ZkbasRollupAddress: %s, NetworkRpc: %s",
		c.ChainConfig.ZkbasContractAddrSysConfigName,
		ZkbasRollupAddress.Value,
		NetworkRpc.Value)

	// load client
	cli, err := _rpc.NewClient(NetworkRpc.Value)
	if err != nil {
		panic(err)
	}

	logx.Info("========================= start monitor blocks =========================")
	err = logic.MonitorBlocks(
		cli,
		c.ChainConfig.StartL1BlockHeight, c.ChainConfig.PendingBlocksCount, c.ChainConfig.MaxHandledBlocksCount,
		ZkbasRollupAddress.Value,
		ctx.L1BlockMonitor,
	)
	if err != nil {
		logx.Error("[logic.MonitorBlocks main] unable to run:", err)
	}
	logx.Info("========================= end monitor blocks =========================")

}
