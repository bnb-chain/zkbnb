package executor

import (
	"bytes"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

type MintNftExecutor struct {
	BaseExecutor

	TxInfo *txtypes.MintNftTxInfo
}

func NewMintNftExecutor(bc IBlockchain, tx *tx.Tx) (TxExecutor, error) {
	txInfo, err := types.ParseMintNftTxInfo(tx.TxInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return nil, types.AppErrInvalidTxInfo
	}

	return &MintNftExecutor{
		BaseExecutor: NewBaseExecutor(bc, tx, txInfo, false),
		TxInfo:       txInfo,
	}, nil
}

func (e *MintNftExecutor) Prepare() error {
	txInfo := e.TxInfo
	if !e.bc.StateDB().DryRun {
		if !e.isDesertExit {
			// Set the right nft index for tx info.
			if e.tx.Rollback == false {
				nextNftIndex := e.bc.StateDB().GetNextNftIndex()
				txInfo.NftIndex = nextNftIndex
			} else {
				//for rollback
				nextNftIndex := e.tx.NftIndex
				txInfo.NftIndex = nextNftIndex
			}
		}
		// Mark the tree states that would be affected in this executor.
		e.MarkNftDirty(txInfo.NftIndex)
	}

	e.MarkAccountAssetsDirty(txInfo.CreatorAccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.GasAccountIndex, []int64{txInfo.GasFeeAssetId})
	e.MarkAccountAssetsDirty(txInfo.ToAccountIndex, []int64{})
	return e.BaseExecutor.Prepare()
}

func (e *MintNftExecutor) VerifyInputs(skipGasAmtChk, skipSigChk bool) error {
	txInfo := e.TxInfo
	if err := e.Validate(); err != nil {
		return err
	}
	//if txInfo.CreatorAccountIndex != txInfo.ToAccountIndex {
	//	return types.AppErrInvalidToAccount
	//}
	err := e.BaseExecutor.VerifyInputs(skipGasAmtChk, skipSigChk)
	if err != nil {
		return err
	}

	creatorAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.CreatorAccountIndex)
	if err != nil {
		return err
	}
	if creatorAccount.CollectionNonce <= txInfo.NftCollectionId {
		return types.AppErrInvalidCollectionId
	}
	if creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(txInfo.GasFeeAssetAmount) < 0 {
		return types.AppErrBalanceNotEnough
	}

	toAccount, err := e.bc.StateDB().GetFormatAccount(txInfo.ToAccountIndex)
	if err != nil {
		return err
	}
	if txInfo.ToL1Address != toAccount.L1Address {
		return types.AppErrInvalidToAddress
	}

	return nil
}

func (e *MintNftExecutor) ApplyTransaction() error {
	bc := e.bc
	txInfo := e.TxInfo

	// apply changes
	creatorAccount, err := bc.StateDB().GetFormatAccount(txInfo.CreatorAccountIndex)
	if err != nil {
		return err
	}

	creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	creatorAccount.Nonce++
	stateCache := e.bc.StateDB()
	stateCache.SetPendingAccount(txInfo.CreatorAccountIndex, creatorAccount)
	stateCache.SetPendingNft(txInfo.NftIndex, &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		RoyaltyRate:         txInfo.RoyaltyRate,
		CollectionId:        txInfo.NftCollectionId,
		NftContentType:      txInfo.NftContentType,
	})
	stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	if !e.isDesertExit {
		if e.tx.Rollback == false {
			e.bc.StateDB().UpdateNftIndex(txInfo.NftIndex)
		}
	}
	return e.BaseExecutor.ApplyTransaction()
}

func (e *MintNftExecutor) GeneratePubData() error {
	txInfo := e.TxInfo

	var buf bytes.Buffer
	buf.WriteByte(uint8(types.TxTypeMintNft))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(common2.Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := common2.FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return err
	}
	buf.Write(packedFeeBytes)
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.RoyaltyRate)))
	buf.Write(common2.Uint16ToBytes(uint16(txInfo.NftCollectionId)))
	buf.Write(common2.PrefixPaddingBufToChunkSize(common.FromHex(txInfo.NftContentHash)))
	buf.WriteByte(uint8(txInfo.NftContentType))

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())

	stateCache := e.bc.StateDB()
	stateCache.PubData = append(stateCache.PubData, pubData...)
	return nil
}

func (e *MintNftExecutor) GetExecutedTx(fromApi bool) (*tx.Tx, error) {
	txInfoBytes, err := json.Marshal(e.TxInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return nil, types.AppErrMarshalTxFailed
	}

	e.tx.TxInfo = string(txInfoBytes)
	e.tx.GasFeeAssetId = e.TxInfo.GasFeeAssetId
	e.tx.GasFee = e.TxInfo.GasFeeAssetAmount.String()
	e.tx.NftIndex = e.TxInfo.NftIndex
	e.tx.IsPartialUpdate = true
	return e.BaseExecutor.GetExecutedTx(fromApi)
}

func (e *MintNftExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	txInfo := e.TxInfo

	copiedAccounts, err := e.bc.StateDB().DeepCopyAccounts([]int64{txInfo.CreatorAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex})
	if err != nil {
		return nil, err
	}

	creatorAccount := copiedAccounts[txInfo.CreatorAccountIndex]
	toAccount := copiedAccounts[txInfo.ToAccountIndex]
	gasAccount := copiedAccounts[txInfo.GasAccountIndex]

	txDetails := make([]*tx.TxDetail, 0, 4)

	// from account gas asset
	order := int64(0)
	accountOrder := int64(0)
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.CreatorAccountIndex,
		L1Address:    creatorAccount.L1Address,
		Balance:      creatorAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			ffmath.Neg(txInfo.GasFeeAssetAmount),
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           creatorAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: creatorAccount.CollectionNonce,
		PublicKey:       creatorAccount.PublicKey,
	})
	creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	if creatorAccount.AssetInfo[txInfo.GasFeeAssetId].Balance.Cmp(types.ZeroBigInt) < 0 {
		return nil, types.AppErrInsufficientGasFeeBalance
	}

	// to account empty delta
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.ToAccountIndex,
		L1Address:    toAccount.L1Address,
		Balance:      toAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			types.ZeroBigInt,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           toAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: toAccount.CollectionNonce,
		PublicKey:       toAccount.PublicKey,
	})

	// to account nft delta
	oldNftInfo := types.EmptyNftInfo(txInfo.NftIndex)
	newNftInfo := &types.NftInfo{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		RoyaltyRate:         txInfo.RoyaltyRate,
		CollectionId:        txInfo.NftCollectionId,
		NftContentType:      txInfo.NftContentType,
	}
	order++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:         txInfo.NftIndex,
		AssetType:       types.NftAssetType,
		AccountIndex:    txInfo.ToAccountIndex,
		L1Address:       toAccount.L1Address,
		Balance:         oldNftInfo.String(),
		BalanceDelta:    newNftInfo.String(),
		Order:           order,
		Nonce:           toAccount.Nonce,
		AccountOrder:    types.NilAccountOrder,
		CollectionNonce: toAccount.CollectionNonce,
		PublicKey:       toAccount.PublicKey,
	})

	// gas account gas asset
	order++
	accountOrder++
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.GasFeeAssetId,
		AssetType:    types.FungibleAssetType,
		AccountIndex: txInfo.GasAccountIndex,
		L1Address:    gasAccount.L1Address,
		Balance:      gasAccount.AssetInfo[txInfo.GasFeeAssetId].String(),
		BalanceDelta: types.ConstructAccountAsset(
			txInfo.GasFeeAssetId,
			txInfo.GasFeeAssetAmount,
			types.ZeroBigInt,
		).String(),
		Order:           order,
		Nonce:           gasAccount.Nonce,
		AccountOrder:    accountOrder,
		CollectionNonce: gasAccount.CollectionNonce,
		IsGas:           true,
		PublicKey:       gasAccount.PublicKey,
	})
	return txDetails, nil
}

func (e *MintNftExecutor) Validate() error {
	if len(e.TxInfo.MetaData) > 2000 {
		return types.AppErrInvalidMetaData.RefineError(2000)
	}
	if len(e.TxInfo.MutableAttributes) > 2000 {
		return types.AppErrInvalidMutableAttributes.RefineError(2000)
	}
	return nil
}
