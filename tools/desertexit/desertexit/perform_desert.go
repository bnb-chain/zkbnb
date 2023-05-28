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

package desertexit

import (
	"context"
	"encoding/json"
	"fmt"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	common2 "github.com/bnb-chain/zkbnb/common"
	monitor2 "github.com/bnb-chain/zkbnb/common/monitor"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/tools/desertexit/config"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"sort"
	"time"
)

type PerformDesert struct {
	Config               config.Config
	cli                  *rpc.ProviderClient
	authCli              *rpc.AuthClient
	zkBNBCli             *zkbnb.ZkBNBClient
	zkbnbInstance        *zkbnb.ZkBNB
	ZkBnbContractAddress string
}

func NewPerformDesert(c config.Config) (*PerformDesert, error) {
	if c.ChainConfig.MaxCancelOutstandingDepositCount == 0 {
		c.ChainConfig.MaxCancelOutstandingDepositCount = 100
	}
	newPerformDesert := &PerformDesert{
		Config: c,
	}
	bscRpcCli, err := rpc.NewClient(c.ChainConfig.BscTestNetRpc)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	newPerformDesert.ZkBnbContractAddress = c.ChainConfig.ZkBnbContractAddress
	newPerformDesert.cli = bscRpcCli
	chainId, err := newPerformDesert.cli.ChainID(context.Background())
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	newPerformDesert.authCli, err = rpc.NewAuthClient(c.ChainConfig.PrivateKey, chainId)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	newPerformDesert.zkbnbInstance, err = zkbnb.LoadZkBNBInstance(newPerformDesert.cli, newPerformDesert.ZkBnbContractAddress)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}

	newPerformDesert.zkBNBCli, err = zkbnb.NewZkBNBClient(newPerformDesert.cli, newPerformDesert.ZkBnbContractAddress)
	if err != nil {
		logx.Severe(err)
		return nil, err
	}
	newPerformDesert.zkBNBCli.ActivateConstructor = newPerformDesert.authCli
	newPerformDesert.zkBNBCli.RevertConstructor = newPerformDesert.authCli
	newPerformDesert.zkBNBCli.PerformConstructor = newPerformDesert.authCli
	newPerformDesert.zkBNBCli.WithdrawConstructor = newPerformDesert.authCli
	newPerformDesert.zkBNBCli.CancelDepositConstructor = newPerformDesert.authCli

	return newPerformDesert, nil
}

func (m *PerformDesert) PerformDesert(performDesertAsset PerformDesertAssetData) error {
	storedBlockInfo := getStoredBlockInfo(performDesertAsset.StoredBlockInfo)
	var proofs []*big.Int

	var proofInfo *prove.FormattedProof
	if err := json.Unmarshal([]byte(performDesertAsset.Proofs), &proofInfo); err != nil {
		return err
	}

	proofs = append(proofs, proofInfo.A[:]...)
	proofs = append(proofs, proofInfo.B[0][0], proofInfo.B[0][1])
	proofs = append(proofs, proofInfo.B[1][0], proofInfo.B[1][1])
	proofs = append(proofs, proofInfo.C[:]...)

	return m.doPerformDesert(storedBlockInfo, common.FromHex(performDesertAsset.PubData), proofs)
}

func (m *PerformDesert) doPerformDesert(storedBlockInfo zkbnb.StorageStoredBlockInfo, pubData []byte, proofs []*big.Int) error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := m.zkBNBCli.PerformDesert(storedBlockInfo, pubData, proofs, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}
	err = m.checkTxSuccess(txHash)
	if err != nil {
		return err
	}
	logx.Infof("performDesert success,txHash=%s", txHash)
	return nil
}

func getStoredBlockInfo(storedBlockInfo *StoredBlockInfo) zkbnb.StorageStoredBlockInfo {
	var pendingOnchainOperationsHash [32]byte
	var stateRoot [32]byte
	var commitment [32]byte

	copy(pendingOnchainOperationsHash[:], common.FromHex(storedBlockInfo.PendingOnchainOperationsHash)[:])
	copy(stateRoot[:], common.FromHex(storedBlockInfo.StateRoot)[:])
	copy(commitment[:], common.FromHex(storedBlockInfo.Commitment)[:])

	return zkbnb.StorageStoredBlockInfo{
		BlockSize:                    storedBlockInfo.BlockSize,
		BlockNumber:                  storedBlockInfo.BlockNumber,
		PriorityOperations:           storedBlockInfo.PriorityOperations,
		PendingOnchainOperationsHash: pendingOnchainOperationsHash,
		Timestamp:                    new(big.Int).SetInt64(storedBlockInfo.Timestamp),
		StateRoot:                    stateRoot,
		Commitment:                   commitment,
	}
}

func (m *PerformDesert) WithdrawPendingBalance(owner common.Address, token common.Address, amount *big.Int) error {
	pendingBalanceBefore, err := m.GetPendingBalance(owner, token)
	if err != nil {
		logx.Errorf("failed to get pending balance: %v", err)
		return err
	}
	logx.Infof("get pending balance,pendingBalanceBefore=%d", pendingBalanceBefore.Int64())

	balanceBefore, err := m.GetBalance(owner, token)
	if err != nil {
		logx.Errorf("failed to get balance: %v", err)
		return err
	}
	logx.Infof("get balance,balanceBefore=%d", balanceBefore.Int64())

	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := m.zkBNBCli.WithdrawPendingBalance(owner, token, amount, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}

	err = m.checkTxSuccess(txHash)
	if err != nil {
		return err
	}

	logx.Infof("withdrawPendingBalance success,txHash=%s", txHash)

	pendingBalanceAfter, err := m.GetPendingBalance(owner, token)
	if err != nil {
		logx.Errorf("failed to get pending balance: %v", err)
		return err
	}
	logx.Infof("get pending balance,pendingBalanceAfter=%d", pendingBalanceAfter.Int64())

	//time.Sleep(30 * time.Second)
	balanceAfter, err := m.GetBalance(owner, token)
	if err != nil {
		logx.Errorf("failed to get balance: %v", err)
		return err
	}
	logx.Infof("get balance,balanceAfter=%d", balanceAfter.Int64())
	return nil
}

func (m *PerformDesert) WithdrawPendingNFTBalance(nftIndex int64) error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := m.zkBNBCli.WithdrawPendingNFTBalance(big.NewInt(nftIndex), gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}

	err = m.checkTxSuccess(txHash)
	if err != nil {
		return err
	}

	logx.Infof("withdrawPendingNFTBalance success,nftIndex=%d,txHash=%s", nftIndex, txHash)
	return nil
}

func (m *PerformDesert) CancelOutstandingDeposit() error {
	newDesertExit, err := NewDesertExit(&m.Config)
	if err != nil {
		return err
	}
	total, err := zkbnb.TotalOpenPriorityRequests(m.zkbnbInstance)
	if err != nil {
		return err
	}
	if total == 0 {
		logx.Infof("There are no outstanding deposits")
		return nil
	}
	requestId, err := zkbnb.FirstPriorityRequestId(m.zkbnbInstance)
	if err != nil {
		return err
	}

	for true {
		priorityRequests, err := newDesertExit.PriorityRequestModel.GetPriorityRequestsByTxTypes(int64(requestId), []int64{monitor2.TxTypeDeposit, monitor2.TxTypeDepositNft})
		if err != nil && err != types.DbErrNotFound {
			return err
		}
		if priorityRequests == nil {
			return nil
		}

		maxRequestId := int64(0)
		depositsPubData := make([][]byte, 0)
		index := int64(0)

		sort.Slice(priorityRequests, func(i, j int) bool {
			return priorityRequests[i].RequestId < priorityRequests[j].RequestId
		})

		for _, request := range priorityRequests {
			logx.Infof("process pending priority request, requestId=%d", request.RequestId)
			depositsPubData = append(depositsPubData, common.FromHex(request.Pubdata))
			maxRequestId = common2.MaxInt64(request.RequestId, maxRequestId)
			if int64(len(depositsPubData[index])) == m.Config.ChainConfig.MaxCancelOutstandingDepositCount {
				err := m.doCancelOutstandingDeposit(uint64(maxRequestId), depositsPubData)
				if err != nil {
					return err
				}
				maxRequestId = int64(0)
				depositsPubData = make([][]byte, 0)
				index = 0
				continue
			}
			index++
		}

		err = m.doCancelOutstandingDeposit(uint64(maxRequestId), depositsPubData)
		if err != nil {
			return err
		}
		requestId = uint64(maxRequestId + 1)
	}
	return nil
}

func (m *PerformDesert) doCancelOutstandingDeposit(maxRequestId uint64, depositsPubData [][]byte) error {
	if len(depositsPubData) == 0 {
		return nil
	}
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := m.zkBNBCli.CancelOutstandingDepositsForExodusMode(maxRequestId, depositsPubData, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		logx.Errorf("failed to send tx: %v:%s", err, txHash)
		return err
	}
	err = m.checkTxSuccess(txHash)
	if err != nil {
		return err
	}

	logx.Infof("cancelOutstandingDepositsForDesertMode success,txHash=%s", txHash)
	return nil
}

func (m *PerformDesert) ActivateDesertMode() error {
	desertMode, err := zkbnb.DesertMode(m.zkbnbInstance)
	if err != nil {
		logx.Errorf("failed to fetch desert mode: %v", err)
		return err
	}
	if desertMode {
		logx.Infof("desert mode has been activated")
		return nil
	}
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := m.zkBNBCli.ActivateDesertMode(gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}

	err = m.checkTxSuccess(txHash)
	if err != nil {
		return err
	}

	logx.Infof("activateDesertMode success,txHash=%s", txHash)
	return nil
}

func (m *PerformDesert) GetBalance(address common.Address, assetAddr common.Address) (*big.Int, error) {
	if assetAddr == common.HexToAddress(types.BNBAddress) {
		amount, err := m.cli.GetBalance(address.Hex())
		if err != nil {
			return nil, err
		}
		logx.Infof("get balance,balance=%d", amount.Int64())
		return amount, nil
	}

	instance, err := zkbnb.LoadERC20(m.cli, assetAddr.Hex())
	if err != nil {
		logx.Severe(err)
		return nil, err
	}

	amount, err := zkbnb.BalanceOf(instance, address, assetAddr)
	if err != nil {
		return nil, err
	}
	logx.Infof("get balance,balance=%d", amount.Int64())
	return amount, nil
}

func (m *PerformDesert) GetPendingBalance(address common.Address, token common.Address) (*big.Int, error) {
	amount, err := zkbnb.GetPendingBalance(m.zkbnbInstance, address, token)
	if err != nil {
		logx.Errorf("failed to get pending balance: %v", err)
		return nil, err
	}
	logx.Infof("get pending balance,pendingBalance=%d", amount.Int64())
	return amount, nil
}

func (m *PerformDesert) checkTxSuccess(txHash string) error {
	startDate := time.Now()
	for {
		receipt, err := m.cli.GetTransactionReceipt(txHash)
		if err != nil {
			logx.Errorf("query transaction receipt %s failed, err: %v", txHash, err)
			if time.Now().After(startDate.Add(time.Duration(m.Config.ChainConfig.MaxWaitingTime) * time.Second)) {
				return fmt.Errorf("failed to sent tx, tx_hash=%s,error=%s", txHash, err)
			}
			continue
		}

		latestL1Height, err := m.cli.GetHeight()
		if err != nil {
			return fmt.Errorf("failed to get l1 block height, err: %v", err)
		}
		if latestL1Height < receipt.BlockNumber.Uint64()+m.Config.ChainConfig.ConfirmBlocksCount {
			continue
		} else {
			if receipt.Status == 0 {
				return fmt.Errorf("failed to sent tx, tx_hash=%s,receipt.Status=0", txHash)
			}
			return nil
		}
	}
}
