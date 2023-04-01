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
	"fmt"
	zkbnb "github.com/bnb-chain/zkbnb-eth-rpc/core"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	common2 "github.com/bnb-chain/zkbnb/common"
	monitor2 "github.com/bnb-chain/zkbnb/common/monitor"
	"github.com/bnb-chain/zkbnb/tools/desertexit/config"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"sort"
)

type PerformDesert struct {
	Config               config.Config
	cli                  *rpc.ProviderClient
	authCli              *rpc.AuthClient
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
	return newPerformDesert, nil
}

func (m *PerformDesert) PerformDesert(performDesertAsset PerformDesertAssetData) error {
	nftRoot := new(big.Int).SetBytes(common.FromHex(performDesertAsset.NftRoot))
	accountExitData, accountMerkleProof := getVerifierExitData(performDesertAsset.AccountExitData, performDesertAsset.AccountMerkleProof)

	storedBlockInfo := getStoredBlockInfo(performDesertAsset.StoredBlockInfo)

	var assetMerkleProof [16]*big.Int
	for i, _ := range performDesertAsset.AssetMerkleProof {
		assetMerkleProof[i] = new(big.Int).SetBytes(common.FromHex(performDesertAsset.AssetMerkleProof[i]))
	}

	assetExitData := zkbnb.DesertVerifierAssetExitData{}
	assetExitData.OfferCanceledOrFinalized = new(big.Int).SetInt64(performDesertAsset.AssetExitData.OfferCanceledOrFinalized)
	assetExitData.Amount = new(big.Int).SetInt64(performDesertAsset.AssetExitData.Amount)
	assetExitData.AssetId = performDesertAsset.AssetExitData.AssetId

	return m.doPerformDesert(storedBlockInfo, nftRoot, assetExitData, accountExitData, assetMerkleProof, accountMerkleProof)
}

func (m *PerformDesert) doPerformDesert(storedBlockInfo zkbnb.StorageStoredBlockInfo, nftRoot *big.Int, assetExitData zkbnb.DesertVerifierAssetExitData, accountExitData zkbnb.DesertVerifierAccountExitData,
	assetMerkleProof [16]*big.Int, accountMerkleProof [32]*big.Int) error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := zkbnb.PerformDesert(m.cli, m.authCli, m.zkbnbInstance, storedBlockInfo, nftRoot, assetExitData, accountExitData, assetMerkleProof, accountMerkleProof, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}
	logx.Infof("performDesert success,txHash=%s", txHash)
	return nil
}

func (m *PerformDesert) PerformDesertNft(performDesertNftData PerformDesertNftData) error {
	accountExitData, accountMerkleProof := getVerifierExitData(performDesertNftData.AccountExitData, performDesertNftData.AccountMerkleProof)
	storedBlockInfo := getStoredBlockInfo(performDesertNftData.StoredBlockInfo)
	assetRoot := new(big.Int).SetBytes(common.FromHex(performDesertNftData.AssetRoot))

	nftMerkleProofs := make([][40]*big.Int, len(performDesertNftData.NftMerkleProofs))

	for i, _ := range performDesertNftData.NftMerkleProofs {
		for j, nftMerkleProof := range performDesertNftData.NftMerkleProofs[i] {
			nftMerkleProofs[i][j] = new(big.Int).SetBytes(common.FromHex(nftMerkleProof))
		}
	}
	exitNfts := make([]zkbnb.DesertVerifierNftExitData, 0)

	for _, nftExitData := range performDesertNftData.ExitNfts {
		var nftContentHash1 [16]byte
		var nftContentHash2 [16]byte
		copy(nftContentHash1[:], common.FromHex(nftExitData.NftContentHash1[:]))
		copy(nftContentHash2[:], common.FromHex(nftExitData.NftContentHash2[:]))
		exitNfts = append(exitNfts, zkbnb.DesertVerifierNftExitData{
			NftIndex:            new(big.Int).SetInt64(int64(nftExitData.NftIndex)),
			OwnerAccountIndex:   new(big.Int).SetInt64(nftExitData.OwnerAccountIndex),
			CreatorAccountIndex: new(big.Int).SetInt64(nftExitData.CreatorAccountIndex),
			NftContentHash1:     nftContentHash1,
			NftContentHash2:     nftContentHash2,
			NftContentType:      nftExitData.NftContentType,
			CreatorTreasuryRate: new(big.Int).SetInt64(nftExitData.CreatorTreasuryRate),
			CollectionId:        new(big.Int).SetInt64(nftExitData.CollectionId),
		})
	}

	return m.doPerformDesertNft(storedBlockInfo, assetRoot, accountExitData, exitNfts, accountMerkleProof, nftMerkleProofs)
}

func (m *PerformDesert) doPerformDesertNft(storedBlockInfo zkbnb.StorageStoredBlockInfo, assetRoot *big.Int, accountExitData zkbnb.DesertVerifierAccountExitData, exitNfts []zkbnb.DesertVerifierNftExitData, accountMerkleProof [32]*big.Int, nftMerkleProofs [][40]*big.Int) error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := zkbnb.PerformDesertNft(m.cli, m.authCli, m.zkbnbInstance, storedBlockInfo, assetRoot, accountExitData, exitNfts, accountMerkleProof, nftMerkleProofs, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}
	logx.Infof("performDesertNft success,txHash=%s", txHash)
	return nil
}

func getVerifierExitData(accountExitData DesertVerifierAccountExitData, accountMerkleProofStr []string) (exitData zkbnb.DesertVerifierAccountExitData, accountMerkleProof [32]*big.Int) {
	var pubKeyX [32]byte
	var pubKeyY [32]byte
	var l1Address [20]byte

	copy(pubKeyX[:], common.FromHex(accountExitData.PubKeyX)[:])
	copy(pubKeyY[:], common.FromHex(accountExitData.PubKeyY)[:])
	copy(l1Address[:], common.FromHex(accountExitData.L1Address)[:])

	exitData.PubKeyX = pubKeyX
	exitData.PubKeyY = pubKeyY
	exitData.CollectionNonce = new(big.Int).SetInt64(accountExitData.CollectionNonce)
	exitData.L1Address = l1Address
	exitData.AccountId = accountExitData.AccountId
	exitData.Nonce = new(big.Int).SetInt64(accountExitData.Nonce)

	for i, _ := range accountMerkleProofStr {
		accountMerkleProof[i] = new(big.Int).SetBytes(common.FromHex(accountMerkleProofStr[i]))
	}
	return exitData, accountMerkleProof
}

func getStoredBlockInfo(storedBlockInfo StoredBlockInfo) zkbnb.StorageStoredBlockInfo {
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

	txHash, err := zkbnb.WithdrawPendingBalance(m.cli, m.authCli, m.zkbnbInstance, owner, token, amount, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
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

func (m *PerformDesert) WithdrawPendingNFTBalance(nftIndex *big.Int) error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := zkbnb.WithdrawPendingNFTBalance(m.cli, m.authCli, m.zkbnbInstance, nftIndex, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}
	logx.Infof("withdrawPendingNFTBalance success,txHash=%s", txHash)
	return nil
}

func (m *PerformDesert) CancelOutstandingDeposit(address string) error {
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
		priorityRequests, err := newDesertExit.PriorityRequestModel.GetPriorityRequestsByTxTypes(address, int64(requestId), []int64{monitor2.TxTypeDeposit, monitor2.TxTypeDepositNft})
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
			depositsPubData[index] = common.FromHex(request.Pubdata)
			maxRequestId = common2.MaxInt64(request.RequestId, maxRequestId)
			if int64(len(depositsPubData[index])) == m.Config.ChainConfig.MaxCancelOutstandingDepositCount {
				m.doCancelOutstandingDeposit(uint64(maxRequestId), depositsPubData)
				maxRequestId = int64(0)
				depositsPubData = make([][]byte, 0)
				index = 0
				continue
			}
			index++
		}

		m.doCancelOutstandingDeposit(uint64(maxRequestId), depositsPubData)
		requestId = uint64(maxRequestId + 1)
	}
	return nil
}

func (m *PerformDesert) doCancelOutstandingDeposit(maxRequestId uint64, depositsPubData [][]byte) {
	if len(depositsPubData) == 0 {
		return
	}
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return
	}

	txHash, err := zkbnb.CancelOutstandingDepositsForDesertMode(m.cli, m.authCli, m.zkbnbInstance, maxRequestId, depositsPubData, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		logx.Errorf("failed to send tx: %v:%s", err, txHash)
		return
	}
	logx.Infof("cancelOutstandingDepositsForDesertMode success,txHash=%s", txHash)
}

func (m *PerformDesert) ActivateDesertMode() error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}

	txHash, err := zkbnb.ActivateDesertMode(m.cli, m.authCli, m.zkbnbInstance, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
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
