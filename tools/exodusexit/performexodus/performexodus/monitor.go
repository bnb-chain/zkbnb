/*
 * Copyright © 2021 ZkBNB Protocol
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

package performexodus

import (
	"fmt"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/performexodus/config"

	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/zeromicro/go-zero/core/logx"
)

type Monitor struct {
	Config               config.Config
	cli                  *rpc.ProviderClient
	ZkBnbContractAddress string
}

func NewMonitor(c config.Config) (*Monitor, error) {
	monitor := &Monitor{
		Config: c,
	}

	bscRpcCli, err := rpc.NewClient(c.ChainConfig.BscTestNetRpc)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}

	monitor.ZkBnbContractAddress = c.ChainConfig.ZkBnbContractAddress
	monitor.cli = bscRpcCli
	return monitor, nil
}

func (m *Monitor) Shutdown() {
}

func (m *Monitor) getBlockRangeToSync() (int64, int64, error) {
	handledHeight := m.Config.ChainConfig.StartL1BlockHeight

	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := m.cli.GetHeight()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get l1 height, err: %v", err)
	}
	safeHeight := latestHeight - m.Config.ChainConfig.ConfirmBlocksCount
	safeHeight = uint64(common2.MinInt64(int64(safeHeight), handledHeight+m.Config.ChainConfig.MaxHandledBlocksCount))
	return handledHeight + 1, int64(safeHeight), nil
}
