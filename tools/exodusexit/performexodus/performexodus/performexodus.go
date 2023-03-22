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
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/generateproof"
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

func (m *PerformExodus) PerformDesert(performDesertAsset generateproof.PerformDesertAssetData) error {
	nftRoot := new(big.Int).SetBytes(common.FromHex(performDesertAsset.NftRoot))
	accountExitData, accountMerkleProof := getVerifierExitData(performDesertAsset.AccountExitData, performDesertAsset.AccountMerkleProof)
	storedBlockInfo := getStoredBlockInfo(performDesertAsset.StoredBlockInfo)
	var assetMerkleProof [16]*big.Int
	for i, _ := range performDesertAsset.AssetMerkleProof {
		assetMerkleProof[i] = new(big.Int).SetBytes(common.FromHex(performDesertAsset.AssetMerkleProof[i]))
	}
	assetExitData := zkbnb.ExodusVerifierAssetExitData{}
	assetExitData.OfferCanceledOrFinalized = new(big.Int).SetInt64(performDesertAsset.AssetExitData.OfferCanceledOrFinalized)
	assetExitData.Amount = new(big.Int).SetInt64(performDesertAsset.AssetExitData.Amount)
	assetExitData.AssetId = performDesertAsset.AssetExitData.AssetId
	return m.doPerformDesert(storedBlockInfo, nftRoot, assetExitData, accountExitData, assetMerkleProof, accountMerkleProof)
}

func (m *PerformExodus) doPerformDesert(storedBlockInfo zkbnb.StorageStoredBlockInfo, nftRoot *big.Int, assetExitData zkbnb.ExodusVerifierAssetExitData, accountExitData zkbnb.ExodusVerifierAccountExitData,
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
	return nil
}

func (m *PerformExodus) PerformDesertNft(performDesertNftData generateproof.PerformDesertNftData) error {
	accountExitData, accountMerkleProof := getVerifierExitData(performDesertNftData.AccountExitData, performDesertNftData.AccountMerkleProof)
	storedBlockInfo := getStoredBlockInfo(performDesertNftData.StoredBlockInfo)
	assetRoot := new(big.Int).SetBytes(common.FromHex(performDesertNftData.AssetRoot))

	nftMerkleProofs := make([][40]*big.Int, len(performDesertNftData.NftMerkleProofs))

	for i, _ := range performDesertNftData.NftMerkleProofs {
		for j, nftMerkleProof := range performDesertNftData.NftMerkleProofs[i] {
			nftMerkleProofs[i][j] = new(big.Int).SetBytes(common.FromHex(nftMerkleProof))
		}
	}
	exitNfts := make([]zkbnb.ExodusVerifierNftExitData, 0)

	for _, nftExitData := range performDesertNftData.ExitNfts {
		var nftContentHash [32]byte
		copy(nftContentHash[:], common.FromHex(nftExitData.NftContentHash)[:])
		exitNfts = append(exitNfts, zkbnb.ExodusVerifierNftExitData{
			NftIndex:            nftExitData.NftIndex,
			OwnerAccountIndex:   new(big.Int).SetInt64(nftExitData.OwnerAccountIndex),
			CreatorAccountIndex: new(big.Int).SetInt64(nftExitData.CreatorAccountIndex),
			NftContentHash:      nftContentHash,
			NftContentType:      nftExitData.NftContentType,
			CreatorTreasuryRate: new(big.Int).SetInt64(nftExitData.CreatorTreasuryRate),
			CollectionId:        new(big.Int).SetInt64(nftExitData.CollectionId),
		})
	}
	return m.doPerformDesertNft(storedBlockInfo, assetRoot, accountExitData, exitNfts, accountMerkleProof, nftMerkleProofs)
}

func (m *PerformExodus) doPerformDesertNft(storedBlockInfo zkbnb.StorageStoredBlockInfo, assetRoot *big.Int, accountExitData zkbnb.ExodusVerifierAccountExitData, exitNfts []zkbnb.ExodusVerifierNftExitData, accountMerkleProof [32]*big.Int, nftMerkleProofs [][40]*big.Int) error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}
	txHash, err := zkbnb.PerformDesertNft(m.cli, m.authCli, m.zkbnbInstance, storedBlockInfo, assetRoot, accountExitData, exitNfts, accountMerkleProof, nftMerkleProofs, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}
	return nil
}

func getVerifierExitData(accountExitData generateproof.ExodusVerifierAccountExitData, accountMerkleProofStr []string) (exitData zkbnb.ExodusVerifierAccountExitData, accountMerkleProof [32]*big.Int) {
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

func getStoredBlockInfo(storedBlockInfo generateproof.StoredBlockInfo) zkbnb.StorageStoredBlockInfo {
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
	depositsPubData := make([][]byte, len(priorityRequests))

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

func (m *PerformExodus) ActivateDesertMode() error {
	gasPrice, err := m.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logx.Errorf("failed to fetch gas price: %v", err)
		return err
	}
	txHash, err := zkbnb.ActivateDesertMode(m.cli, m.authCli, m.zkbnbInstance, gasPrice, m.Config.ChainConfig.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to send tx: %v:%s", err, txHash)
	}
	return nil
}

func (m *PerformExodus) getOutstandingDeposits() (priorityRequests []*priorityrequest.PriorityRequest, err error) {
	priorityRequests = make([]*priorityrequest.PriorityRequest, 0)
	handledHeight := m.Config.ChainConfig.StartL1BlockHeight
	for {
		startHeight, endHeight, err := m.getBlockRangeToSync(handledHeight)
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
		handledHeight = endHeight
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

func (m *PerformExodus) getBlockRangeToSync(handledHeight int64) (int64, int64, error) {
	// get latest l1 block height(latest height - pendingBlocksCount)
	latestHeight, err := m.cli.GetHeight()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get l1 height, err: %v", err)
	}
	safeHeight := latestHeight - m.Config.ChainConfig.ConfirmBlocksCount
	safeHeight = uint64(common2.MinInt64(int64(safeHeight), handledHeight+m.Config.ChainConfig.MaxHandledBlocksCount))
	return handledHeight + 1, int64(safeHeight), nil
}
