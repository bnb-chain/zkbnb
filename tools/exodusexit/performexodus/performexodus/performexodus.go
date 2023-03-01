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
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb-go-sdk/txutils"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/dao/priorityrequest"
	"github.com/bnb-chain/zkbnb/service/monitor/monitor"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/performexodus/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"time"
)

type PerformExodus struct {
	Config               config.Config
	cli                  *rpc.ProviderClient
	ZkBnbContractAddress string
}

func NewPerformExodus(c config.Config) (*PerformExodus, error) {
	newPerformExodus := &PerformExodus{
		Config: c,
	}
	bscRpcCli, err := rpc.NewClient(c.ChainConfig.BscTestNetRpc)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	newPerformExodus.ZkBnbContractAddress = c.ChainConfig.ZkBnbContractAddress
	newPerformExodus.cli = bscRpcCli
	return newPerformExodus, nil
}

func (m *PerformExodus) SubmitProof() error {
	return nil
}

func (m *PerformExodus) WithdrawPendingBalance() error {
	return nil
}

func (m *PerformExodus) CancelOutstandingDeposit() error {
	if m.Config.AccountName == "" {
		return nil
	}
	accountNameHash, err := txutils.AccountNameHash(m.Config.AccountName)
	if err != nil {
		return err
	}
	priorityRequests, err := m.getOutstandingDeposits()
	if err != nil {
		return err
	}
	for _, request := range priorityRequests {
		logx.Infof("process pending priority request, requestId=%d", request.RequestId)
		// handle request based on request type
		switch request.TxType {
		case monitor.TxTypeDeposit:
			txInfo, err := chain.ParseDepositPubData(common.FromHex(request.Pubdata))
			if err != nil {
				return fmt.Errorf("unable to parse deposit pub data: %v", err)
			}
			if accountNameHash == string(txInfo.AccountNameHash) {

			}
		case monitor.TxTypeDepositNft:
			//txInfo, err := chain.ParseDepositNftPubData(common.FromHex(request.Pubdata))
			//if err != nil {
			//	return fmt.Errorf("unable to parse deposit nft pub data: %v", err)
			//}
		default:
			return fmt.Errorf("invalid request type")
		}
	}
	return nil
}

func (m *PerformExodus) getOutstandingDeposits() (priorityRequests []*priorityrequest.PriorityRequest, err error) {
	priorityRequests = make([]*priorityrequest.PriorityRequest, 0)
	for {
		startHeight, endHeight, err := m.getBlockRangeToSync()
		if err != nil {
			logx.Errorf("get block range to sync error, err: %s", err.Error())
			return nil, err
		}
		if startHeight > m.Config.ChainConfig.EndL1BlockHeight {
			return priorityRequests, nil
		}
		if endHeight < startHeight {
			logx.Infof("no blocks to sync, startHeight: %d, endHeight: %d", startHeight, endHeight)
			time.Sleep(30 * time.Second)
			continue
		}

		logx.Infof("syncing generic l1 blocks from %d to %d", big.NewInt(startHeight), big.NewInt(endHeight))

		logs, err := monitor.GetZkBNBContractLogs(m.cli, m.ZkBnbContractAddress, uint64(startHeight), uint64(endHeight))
		if err != nil {
			return nil, fmt.Errorf("failed to get contract logs, err: %v", err)
		}

		logx.Infof("type is typeGeneric blocks from %d to %d and vlog len: %v", startHeight, endHeight, len(logs))
		for _, vlog := range logs {
			logx.Infof("type is typeGeneric blocks from %d to %d and vlog: %v", startHeight, endHeight, vlog)
		}
		for _, vlog := range logs {
			if vlog.BlockNumber > uint64(m.Config.ChainConfig.EndL1BlockHeight) {
				return priorityRequests, nil
			}
			if vlog.Removed {
				logx.Errorf("Removed to get vlog,TxHash:%v,Index:%v", vlog.TxHash, vlog.Index)
				continue
			}
			switch vlog.Topics[0].Hex() {
			case monitor.ZkbnbLogNewPriorityRequestSigHash.Hex():
				l2TxEventMonitorInfo, err := monitor.ConvertLogToNewPriorityRequestEvent(vlog)
				if err != nil {
					return nil, fmt.Errorf("failed to convert NewPriorityRequest log, err: %v", err)
				}
				priorityRequests = append(priorityRequests, l2TxEventMonitorInfo)
			default:
			}
		}
	}
}

func (m *PerformExodus) getBlockRangeToSync() (int64, int64, error) {
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
