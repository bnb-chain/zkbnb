package logic

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/bnb-chain/zkbas-crypto/accumulators/merkleTree"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/blockForCommit"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func ConvertMempoolTxToTx(mempoolTx *MempoolTx, txDetails []*tx.TxDetail, accountRoot string, currentBlockHeight int64) (tx *Tx) {
	tx = &Tx{
		TxHash:        mempoolTx.TxHash,
		TxType:        mempoolTx.TxType,
		GasFee:        mempoolTx.GasFee,
		GasFeeAssetId: mempoolTx.GasFeeAssetId,
		TxStatus:      TxStatusPending,
		BlockHeight:   currentBlockHeight,
		StateRoot:     accountRoot,
		NftIndex:      mempoolTx.NftIndex,
		PairIndex:     mempoolTx.PairIndex,
		AssetId:       mempoolTx.AssetId,
		TxAmount:      mempoolTx.TxAmount,
		NativeAddress: mempoolTx.NativeAddress,
		TxInfo:        mempoolTx.TxInfo,
		TxDetails:     txDetails,
		ExtraInfo:     mempoolTx.ExtraInfo,
		Memo:          mempoolTx.Memo,
		AccountIndex:  mempoolTx.AccountIndex,
		Nonce:         mempoolTx.Nonce,
		ExpiredAt:     mempoolTx.ExpiredAt,
	}
	return tx
}

func newStateRoot(accountTree, liquidityTree, nftTree *merkleTree.Tree) string {
	hFunc := mimc.NewMiMC()
	hFunc.Write(accountTree.RootNode.Value)
	hFunc.Write(liquidityTree.RootNode.Value)
	hFunc.Write(nftTree.RootNode.Value)
	stateRoot := common.Bytes2Hex(hFunc.Sum(nil))
	return stateRoot
}

func newAccount(pendingUpdateAccountIndexMap map[int64]bool, accountMap map[int64]*commonAsset.AccountInfo,
	currentBlockHeight int64) ([]*Account, []*AccountHistory, error) {
	var pendingUpdateAccounts []*Account
	var pendingNewAccountHistory []*AccountHistory
	for accountIndex, flag := range pendingUpdateAccountIndexMap {
		if !flag {
			continue
		}
		accountInfo, err := commonAsset.FromFormatAccountInfo(accountMap[accountIndex])
		if err != nil {
			logx.Errorf("[CommitterTask] unable to convert from format account info: %s", err.Error())
			return nil, nil, err
		}
		pendingUpdateAccounts = append(pendingUpdateAccounts, accountInfo)
		pendingNewAccountHistory = append(pendingNewAccountHistory, &AccountHistory{
			AccountIndex:    accountInfo.AccountIndex,
			Nonce:           accountInfo.Nonce,
			CollectionNonce: accountInfo.CollectionNonce,
			AssetInfo:       accountInfo.AssetInfo,
			AssetRoot:       accountInfo.AssetRoot,
			L2BlockHeight:   currentBlockHeight,
		})
	}
	return pendingUpdateAccounts, pendingNewAccountHistory, nil
}

func newLiquidity(pendingUpdateLiquidityIndexMap map[int64]bool, liquidityMap map[int64]*liquidity.Liquidity,
	currentBlockHeight int64) ([]*liquidity.Liquidity, []*liquidity.LiquidityHistory) {
	var pendingUpdateLiquidity []*liquidity.Liquidity
	var pendingNewLiquidityHistory []*liquidity.LiquidityHistory
	for pairIndex, flag := range pendingUpdateLiquidityIndexMap {
		if !flag {
			continue
		}
		pendingUpdateLiquidity = append(pendingUpdateLiquidity, liquidityMap[pairIndex])
		pendingNewLiquidityHistory = append(pendingNewLiquidityHistory, &LiquidityHistory{
			PairIndex:            liquidityMap[pairIndex].PairIndex,
			AssetAId:             liquidityMap[pairIndex].AssetAId,
			AssetA:               liquidityMap[pairIndex].AssetA,
			AssetBId:             liquidityMap[pairIndex].AssetBId,
			AssetB:               liquidityMap[pairIndex].AssetB,
			LpAmount:             liquidityMap[pairIndex].LpAmount,
			KLast:                liquidityMap[pairIndex].KLast,
			FeeRate:              liquidityMap[pairIndex].FeeRate,
			TreasuryAccountIndex: liquidityMap[pairIndex].TreasuryAccountIndex,
			TreasuryRate:         liquidityMap[pairIndex].TreasuryRate,
			L2BlockHeight:        currentBlockHeight,
		})
	}
	return pendingUpdateLiquidity, pendingNewLiquidityHistory
}

func newPendingNewNftHistory(pendingNewNftIndexMap map[int64]bool, nftMap map[int64]*nft.L2Nft, currentBlockHeight int64) []*nft.L2NftHistory {
	var pendingNewNftHistory []*nft.L2NftHistory
	for nftIndex, flag := range pendingNewNftIndexMap {
		if !flag {
			continue
		}
		pendingNewNftHistory = append(pendingNewNftHistory, &L2NftHistory{
			NftIndex:            nftMap[nftIndex].NftIndex,
			CreatorAccountIndex: nftMap[nftIndex].CreatorAccountIndex,
			OwnerAccountIndex:   nftMap[nftIndex].OwnerAccountIndex,
			NftContentHash:      nftMap[nftIndex].NftContentHash,
			NftL1Address:        nftMap[nftIndex].NftL1Address,
			NftL1TokenId:        nftMap[nftIndex].NftL1TokenId,
			CreatorTreasuryRate: nftMap[nftIndex].CreatorTreasuryRate,
			CollectionId:        nftMap[nftIndex].CollectionId,
			L2BlockHeight:       currentBlockHeight,
		})
	}
	return pendingNewNftHistory
}

func newBlock(createAtTime time.Time, commitment, finalStateRoot string, txs []*tx.Tx, pendingOnChainOperationsPubData [][]byte,
	currentBlockHeight, priorityOperations int64, pendingOnChainOperationsHash []byte) (*block.Block, error) {
	block := &Block{
		Model:                        gorm.Model{CreatedAt: createAtTime},
		BlockCommitment:              commitment,
		BlockHeight:                  currentBlockHeight,
		StateRoot:                    finalStateRoot,
		PriorityOperations:           priorityOperations,
		PendingOnChainOperationsHash: common.Bytes2Hex(pendingOnChainOperationsHash),
		Txs:                          txs,
		BlockStatus:                  block.StatusPending,
	}
	if pendingOnChainOperationsPubData != nil {
		onChainOperationsPubDataBytes, err := json.Marshal(pendingOnChainOperationsPubData)
		if err != nil {
			logx.Errorf("[CommitterTask] unable to marshal on chain operations pub data: %s", err.Error())
			return nil, err
		}
		block.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
	}
	return block, nil
}
func newBlockForCommit(currentBlockHeight int64, finalStateRoot string,
	pubData []byte, createdAt int64, pubDataOffset []uint32) (*blockForCommit.BlockForCommit, error) {
	offsetBytes, err := json.Marshal(pubDataOffset)
	if err != nil {
		logx.Errorf("[Marshal] unable to marshal pub data: %s", err.Error())
		return nil, err
	}
	return &BlockForCommit{
		BlockHeight:       currentBlockHeight,
		StateRoot:         finalStateRoot,
		PublicData:        common.Bytes2Hex(pubData),
		Timestamp:         createdAt,
		PublicDataOffsets: string(offsetBytes),
	}, nil
}

/**
handleTxPubData: handle different layer-1 txs
*/
func handleTxPubData(
	mempoolTx *MempoolTx,
	oldPubData []byte,
	oldPendingOnChainOperationsPubData [][]byte,
	oldPendingOnChainOperationsHash []byte,
	oldPubDataOffset []uint32,
) (
	priorityOperation int64,
	newPendingOnChainOperationsPubData [][]byte,
	newPendingOnChainOperationsHash []byte,
	newPubData []byte,
	newPubDataOffset []uint32,
	err error,
) {
	priorityOperation = 0
	newPendingOnChainOperationsHash = oldPendingOnChainOperationsHash
	newPendingOnChainOperationsPubData = oldPendingOnChainOperationsPubData
	newPubDataOffset = oldPubDataOffset
	var pubData []byte
	switch mempoolTx.TxType {
	case TxTypeRegisterZns:
		pubData, err = util.ConvertTxToRegisterZNSPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to registerZNS pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		priorityOperation++
		break
	case TxTypeCreatePair:
		pubData, err = util.ConvertTxToCreatePairPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to create pair pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		priorityOperation++
		break
	case TxTypeUpdatePairRate:
		pubData, err = util.ConvertTxToUpdatePairRatePubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to update pair rate pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		priorityOperation++
		break
	case TxTypeDeposit:
		pubData, err = util.ConvertTxToDepositPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to deposit pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		priorityOperation++
		break
	case TxTypeDepositNft:
		pubData, err = util.ConvertTxToDepositNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to deposit nft pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		priorityOperation++
		break
	case TxTypeTransfer:
		pubData, err = util.ConvertTxToTransferPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to transfer pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeSwap:
		pubData, err = util.ConvertTxToSwapPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to swap pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeAddLiquidity:
		pubData, err = util.ConvertTxToAddLiquidityPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to add liquidity pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeRemoveLiquidity:
		pubData, err = util.ConvertTxToRemoveLiquidityPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to remove liquidity pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeCreateCollection:
		pubData, err = util.ConvertTxToCreateCollectionPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to create collection pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeMintNft:
		pubData, err = util.ConvertTxToMintNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to mint nft pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeTransferNft:
		pubData, err = util.ConvertTxToTransferNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to transfer nft pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeAtomicMatch:
		pubData, err = util.ConvertTxToAtomicMatchPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to atomic match pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeCancelOffer:
		pubData, err = util.ConvertTxToCancelOfferPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to cancel offer pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		break
	case TxTypeWithdraw:
		pubData, err = util.ConvertTxToWithdrawPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to withdraw pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		newPendingOnChainOperationsPubData = append(newPendingOnChainOperationsPubData, pubData)
		newPendingOnChainOperationsHash = util.ConcatKeccakHash(oldPendingOnChainOperationsHash, pubData)
		break
	case TxTypeWithdrawNft:
		pubData, err = util.ConvertTxToWithdrawNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to withdraw nft pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		newPendingOnChainOperationsPubData = append(newPendingOnChainOperationsPubData, pubData)
		newPendingOnChainOperationsHash = util.ConcatKeccakHash(oldPendingOnChainOperationsHash, pubData)
		break
	case TxTypeFullExit:
		pubData, err = util.ConvertTxToFullExitPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to full exit pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		priorityOperation++
		newPendingOnChainOperationsPubData = append(newPendingOnChainOperationsPubData, pubData)
		newPendingOnChainOperationsHash = util.ConcatKeccakHash(oldPendingOnChainOperationsHash, pubData)
		break
	case TxTypeFullExitNft:
		pubData, err = util.ConvertTxToFullExitNftPubData(mempoolTx)
		if err != nil {
			logx.Errorf("[handleTxPubData] unable to convert tx to full exit nft pub data")
			return priorityOperation, nil, nil, nil, nil, err
		}
		newPubDataOffset = append(newPubDataOffset, uint32(len(oldPubData)))
		priorityOperation++
		newPendingOnChainOperationsPubData = append(newPendingOnChainOperationsPubData, pubData)
		newPendingOnChainOperationsHash = util.ConcatKeccakHash(oldPendingOnChainOperationsHash, pubData)
		break
	default:
		logx.Errorf("[handleTxPubData] invalid tx type")
		return priorityOperation, nil, nil, nil, nil, errors.New("[handleTxPubData] invalid tx type")
	}
	newPubData = append(oldPubData, pubData...)
	return priorityOperation, newPendingOnChainOperationsPubData, newPendingOnChainOperationsHash, newPubData, newPubDataOffset, nil
}
