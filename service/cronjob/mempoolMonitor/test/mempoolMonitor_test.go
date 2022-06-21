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
	"github.com/bnb-chain/zkbas/common/model/l2TxEventMonitor"
	"github.com/bnb-chain/zkbas/service/cronjob/mempoolMonitor/internal/config"
	"github.com/bnb-chain/zkbas/service/cronjob/mempoolMonitor/internal/logic"
	"github.com/bnb-chain/zkbas/service/cronjob/mempoolMonitor/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"testing"
)

var configFile = flag.String("f",
	"/Users/gavin/Desktop/zecrey-v2/service/rpc/mempoolMonitor/etc/mempoolMonitor.yaml", "the config file")

func TestMempoolMonitor(t *testing.T) {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)

	logx.Info("===== start monitor mempool txs")
	err := logic.MonitorMempool(
		ctx,
	)
	if err != nil {
		if err == l2TxEventMonitor.ErrNotFound {
			logx.Info("[mempoolMonitor.MonitorMempool main] no l2 tx event need to monitor")
		} else {
			logx.Info("[mempoolMonitor.MonitorMempool main] unable to run:", err)
		}
	}
	logx.Info("===== end monitor mempool txs")
}
