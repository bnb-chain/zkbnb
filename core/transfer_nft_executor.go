package core

import (
	"bytes"
	"encoding/json"
	"math/big"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"
)

type TransferNftExecutor struct {
	BaseExecutor

	txInfo *legendTxTypes.TransferNftTxInfo
}

func NewTransferNftExecutor(bc *BlockChain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := commonTx.ParseTransferNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, errors.New("invalid tx info")
	}

	return &TransferNftExecutor{
		BaseExecutor: BaseExecutor{
			bc:      bc,
			tx:      tx,
			iTxInfo: txInfo,
		},
		txInfo: txInfo,
	}, nil
}

func (e *TransferNftExecutor) Prepare() error {
	txInfo := e.txInfo

	accounts := []int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.GasFeeAssetId}
	err := e.bc.prepareAccountsAndAssets(accounts, assets)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}

	err = e.bc.prepareNft(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("prepare nft failed")
		return errors.New("internal error")
	}

	return nil
}

func (e *TransferNftExecutor) VerifyInputs() error {
	txInfo := e.txInfo

	err := e.BaseExecutor.VerifyInputs()
	if err != nil {
		return err
	}

	fromAccount := e.bc.accountMap[txInfo.FromAccountIndex]
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return errors.New("balance is not enough")
	}

	toAccount := e.bc.accountMap[txInfo.ToAccountIndex]
	if txInfo.ToAccountNameHash != toAccount.AccountNameHash {
		return errors.New("invalid ToAccountNameHash")
	}

	nft := e.bc.nftMap[txInfo.NftIndex]
	if nft.OwnerAccountIndex != txInfo.FromAccountIndex {
		return errors.New("account is not owner of the nft")
	}

	return nil
}

func (e *TransferNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.txInfo

	fromAccount := bc.accountMap[txInfo.FromAccountIndex]
	gasAccount := bc.accountMap[txInfo.GasAccountIndex]
	nft := bc.nftMap[txInfo.NftIndex]

	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Add(gasAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	fromAccount.Nonce++
	nft.OwnerAccountIndex = txInfo.ToAccountIndex

	stateCache := e.bc.stateCache
	stateCache.pendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = StateCachePending
	stateCache.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = StateCachePending
	stateCache.pendingUpdateNftIndexMap[txInfo.NftIndex] = StateCachePending
	return nil
}

func (e *TransferNftExecutor) GeneratePubData() error {
	txInfo := e.txInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(commonTx.TxTypeTransferNft))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(util.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(util.Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(util.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := util.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		return err
	}
	buf.Write(packedFeeBytes)
	chunk := util.SuffixPaddingBufToChunkSize(buf.Bytes())
	buf.Reset()
	buf.Write(chunk)
	buf.Write(util.PrefixPaddingBufToChunkSize(txInfo.CallDataHash))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	buf.Write(util.PrefixPaddingBufToChunkSize([]byte{}))
	pubData := buf.Bytes()

	stateCache := e.bc.stateCache
	stateCache.pubData = append(stateCache.pubData, pubData...)
	return nil
}

func (e *TransferNftExecutor) UpdateTrees() error {
	txInfo := e.txInfo

	accounts := []int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}
	assets := []int64{txInfo.GasFeeAssetId}

	err := e.bc.updateAccountTree(accounts, assets)
	if err != nil {
		logx.Errorf("update account tree error, err: %s", err.Error())
		return err
	}

	err = e.bc.updateNftTree(txInfo.NftIndex)
	if err != nil {
		logx.Errorf("update nft tree error, err: %s", err.Error())
		return err
	}
	return nil
}

func (e *TransferNftExecutor) GetExecutedTx() (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, errors.New("unmarshal tx failed")
	}

	e.tx.TxInfo = string(txInfoBytes)
	return e.BaseExecutor.GetExecutedTx()
}

func (e *TransferNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.txInfo
	nftModel := e.bc.nftMap[txInfo.NftIndex]

	copiedAccounts, err := e.bc.deepCopyAccounts([]int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}
	fromAccount := copiedAccounts[txInfo.FromAccountIndex]
	toAccount := copiedAccounts[txInfo.ToAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)

	// from account gas asset
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.FromAccountIndex,
		AccountName:  fromAccount.AccountName,
		Balance:      fromAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           fromAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: fromAccount.CollectionNonce,
	})
	fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.New("insufficient gas fee balance")
	}

	// to account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		AccountName:  toAccount.AccountName,
		Balance:      toAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ZeroBigInt,
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           toAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: toAccount.CollectionNonce,
	})

	// to account nft delta
	oldNftInfo := &commonAsset.NftInfo{
		NftIndex:            nftModel.NftIndex,
		CreatorAccountIndex: nftModel.CreatorAccountIndex,
		OwnerAccountIndex:   nftModel.OwnerAccountIndex,
		NftContentHash:      nftModel.NftContentHash,
		NftL1TokenId:        nftModel.NftL1TokenId,
		NftL1Address:        nftModel.NftL1Address,
		CreatorTreasuryRate: nftModel.CreatorTreasuryRate,
		CollectionId:        nftModel.CollectionId,
	}
	newNftInfo := &commonAsset.NftInfo{
		NftIndex:            nftModel.NftIndex,
		CreatorAccountIndex: nftModel.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      nftModel.NftContentHash,
		NftL1TokenId:        nftModel.NftL1TokenId,
		NftL1Address:        nftModel.NftL1Address,
		CreatorTreasuryRate: nftModel.CreatorTreasuryRate,
		CollectionId:        nftModel.CollectionId,
	}
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       commonAsset.NftAssetType,
		AccountIndex:    txInfo.ToAccountIndex,
		AccountName:     toAccount.AccountName,
		Balance:         oldNftInfo.String(),
		BalanceDelta:    newNftInfo.String(),
		Order:           order,
		Nonce:           toAccount.Nonce,
		AccountOrder:    commonConstant.NilAccountOrder,
		CollectionNonce: toAccount.CollectionNonce,
	})

	// gas account gas asset
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		AccountName:  gasAccount.AccountName,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: commonAsset.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			ZeroBigInt,
			ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           gasAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: gasAccount.CollectionNonce,
	})
	return txDetails, nil
}

func (e *TransferNftExecutor) GenerateMempoolTx() (*mempool.MempoolTx, error) {
	hash, err := legendTxTypes.ComputeTransferNftMsgHash(e.txInfo, mimc.NewMiMC())
	if err != nil {
		return nil, err
	}
	txHash := common.Bytes2Hex(hash)

	mempoolTx := &mempool.MempoolTx{
		TxHash:        txHash,
		TxType:        e.tx.TxType,
		GasFeeAssetId: e.txInfo.GasFeeAssetId,
		GasFee:        e.txInfo.GasFeeAssetAmount.String(),
		NftIndex:      e.txInfo.NftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		Memo:          "",
		AccountIndex:  e.txInfo.FromAccountIndex,
		Nonce:         e.txInfo.Nonce,
		ExpiredAt:     e.txInfo.ExpiredAt,
		L2BlockHeight: commonConstant.NilBlockHeight,
		Status:        mempool.PendingTxStatus,
		TxInfo:        e.tx.TxInfo,
	}
	return mempoolTx, nil
}
