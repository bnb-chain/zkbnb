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
	"context"
	"fmt"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	common2 "github.com/bnb-chain/zkbnb/common"
	monitor2 "github.com/bnb-chain/zkbnb/common/monitor"
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
	authCli              *rpc.AuthClient
	zkbnbInstance        *zkbnb.ZkBNB
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
	chainId, err := newPerformExodus.cli.ChainID(context.Background())
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	newPerformExodus.authCli, err = rpc.NewAuthClient(c.ChainConfig.PrivateKey, chainId)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	newPerformExodus.zkbnbInstance, err = zkbnb.LoadZkBNBInstance(newPerformExodus.cli, newPerformExodus.ZkBnbContractAddress)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	return newPerformExodus, nil
}

//func (m *PerformExodus) PerformDesert(nftRoot [32]byte,
//	exitData zkbnb.ExodusVerifierExitData, assetMerkleProof [16][32]byte, accountMerkleProof [32][32]byte) error {
//	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
//	if err != nil {
//		logx.Errorf("failed to fetch gas price: %v", err)
//		return err
//	}
//	txHash, err := zkbnb.PerformDesert(m.cli, m.authCli, m.zkbnbInstance, nftRoot, exitData, assetMerkleProof, accountMerkleProof, gasPrice, m.Config.ChainConfig.GasLimit)
//	if err != nil {
//		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
//	}
//	return nil
//}

//func (m *PerformExodus) PerformDesertNft(ownerAccountIndex *big.Int, accountRoot [32]byte, exitNfts []zkbnb.ExodusVerifierExitNftData, nftMerkleProofs [][40][32]byte) error {
//	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
//	if err != nil {
//		logx.Errorf("failed to fetch gas price: %v", err)
//		return err
//	}
//	txHash, err := zkbnb.PerformDesertNft(m.cli, m.authCli, m.zkbnbInstance, ownerAccountIndex, accountRoot, exitNfts, nftMerkleProofs, gasPrice, m.Config.ChainConfig.GasLimit)
//	if err != nil {
//		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
//	}
//	return nil
//}

func (m *PerformExodus) WithdrawPendingBalance(owner common.Address, token common.Address, amount *big.Int) error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}
	txHash, err := zkbnb.WithdrawPendingBalance(m.cli, m.authCli, m.zkbnbInstance, owner, token, amount, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}
	return nil
}

func (m *PerformExodus) WithdrawPendingNFTBalance(nftIndex *big.Int) error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}
	txHash, err := zkbnb.WithdrawPendingNFTBalance(m.cli, m.authCli, m.zkbnbInstance, nftIndex, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}
	return nil
}

func (m *PerformExodus) CancelOutstandingDeposit() error {
	priorityRequests, err := m.getOutstandingDeposits()
	if err != nil {
		return err
	}
	maxRequestId := int64(0)
	depositsPubData := make([][]byte, 0)

	for i, request := range priorityRequests {
		logx.Infof("process pending priority request, requestId=%d", request.RequestId)
		depositsPubData[i] = []byte(request.Pubdata)
		maxRequestId = common2.MaxInt64(request.RequestId, maxRequestId)
	}
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}
	txHash, err := zkbnb.CancelOutstandingDepositsForExodusMode(m.cli, m.authCli, m.zkbnbInstance, uint64(maxRequestId), depositsPubData, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
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
			case monitor2.ZkbnbLogNewPriorityRequestSigHash.Hex():
				l2TxEventMonitorInfo, err := monitor.ConvertLogToNewPriorityRequestEvent(vlog)
				if err != nil {
					return nil, fmt.Errorf("failed to convert NewPriorityRequest log, err: %v", err)
				}
				if l2TxEventMonitorInfo.TxType == monitor2.TxTypeDeposit || l2TxEventMonitorInfo.TxType == monitor2.TxTypeDepositNft {
					priorityRequests = append(priorityRequests, l2TxEventMonitorInfo)
				}
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
