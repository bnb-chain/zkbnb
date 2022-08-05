package logic

import (
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

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
