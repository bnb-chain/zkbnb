/*
 * Copyright © 2021 Zecrey Protocol
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
 *
 */

package util

import (
	"bytes"
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/commonTx"
	"github.com/zecrey-labs/zecrey-core/common/zecrey-legend/model/mempool"
	"github.com/zeromicro/go-zero/core/logx"
)

func ConvertTxToRegisterZNSPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		struct RegisterZNS {
			uint8 txType;
			bytes32 accountName;
			bytes32 pubKey;
		}
	*/
	if oTx.TxType != commonTx.TxTypeRegisterZns {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToRegisterZNSPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseRegisterZnsTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	nameBytes := AccountNameToBytes32(txInfo.AccountName)
	buf.Write(nameBytes[:])
	pkBytes, err := PubKeyStrToBytes32(txInfo.PubKey)
	if err != nil {
		logx.Errorf("[ConvertTxToRegisterZNSPubData] unable to convert pk to bytes32: %s", err.Error())
		return nil, err
	}
	buf.Write(pkBytes)

	return buf.Bytes(), nil
}

func ConvertTxToDepositPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		struct Deposit {
			uint8 txType;
			uint32 accountIndex;
			bytes32 accountNameHash;
			uint16 assetId;
			uint128 amount;
		}
	*/
	if oTx.TxType != commonTx.TxTypeDeposit {
		logx.Errorf("[ConvertTxToDepositPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToDepositPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseDepositTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(txInfo.AccountIndex))
	buf.Write(common.FromHex(txInfo.AccountNameHash))
	buf.Write(Uint16ToBytes(txInfo.AssetId))
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	return buf.Bytes(), nil
}

func ConvertTxToTransferPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		txType 1byte
		fromAccountIndex 4byte
		toAccountIndex 4byte
		assetId 2byte
		assetAmount 5byte
		gasFeeAccountIndex 4byte
		gasFeeAssetId 2byte
		gasFeeAssetAmount 2byte
		callDataHash 32byte
	*/
	if oTx.TxType != commonTx.TxTypeTransfer {
		logx.Errorf("[ConvertTxToTransferPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToTransferPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseTransferTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	packedAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	buf.Write(txInfo.CallDataHash)
	return buf.Bytes(), nil
}

func ConvertTxToSwapPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		txType 1byte
		fromAccountIndex 4byte
		toAccountIndex 4byte
		pairIndex 2byte
		assetAId 2byte
		assetBId 2byte
		assetAAmount 5byte
		assetBAmount 5byte
		treasuryAccountIndex 4byte
		treasuryFeeAmount 2byte
		gasFeeAccountIndex 4byte
		gasFeeAssetId 2byte
		gasFeeAssetAmount 2byte
	*/
	if oTx.TxType != commonTx.TxTypeSwap {
		logx.Errorf("[ConvertTxToSwapPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToSwapPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseSwapTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToSwapPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetAId)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetBId)))
	packedAssetAAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountDeltaBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountDeltaBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.TreasuryAccountIndex)))
	packedTreasuryFeeAmountDeltaBytes, err := FeeToPackedFeeBytes(txInfo.TreasuryFeeAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedTreasuryFeeAmountDeltaBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToAddLiquidityPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		txType 1byte
		fromAccountIndex 4byte
		toAccountIndex 4byte
		pairIndex 2byte
		assetAId 2byte
		assetBId 2byte
		assetAAmount 5byte
		assetBAmount 5byte
		lpAmount 5byte
		gasFeeAccountIndex 4byte
		gasFeeAssetId 2byte
		gasFeeAssetAmount 2byte
	*/
	if oTx.TxType != commonTx.TxTypeAddLiquidity {
		logx.Errorf("[ConvertTxToAddLiquidityPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToAddLiquidityPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseAddLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToAddLiquidityPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetAId)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetBId)))
	packedAssetAAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountBytes)
	packedAssetBAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountBytes)
	buf.Write(Uint40ToBytes(txInfo.LpAmount.Int64()))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToRemoveLiquidityPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		txType 1byte
		fromAccountIndex 4byte
		toAccountIndex 4byte
		pairIndex 2byte
		assetAId 2byte
		assetBId 2byte
		assetAAmount 5byte
		assetBAmount 5byte
		lpAmount 5byte
		gasFeeAccountIndex 4byte
		gasFeeAssetId 2byte
		gasFeeAssetAmount 2byte
	*/
	if oTx.TxType != commonTx.TxTypeRemoveLiquidity {
		logx.Errorf("[ConvertTxToRemoveLiquidityPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToRemoveLiquidityPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseRemoveLiquidityTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToRemoveLiquidityPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.PairIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetAId)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetBId)))
	packedAssetAAmountDeltaBytes, err := AmountToPackedAmountBytes(txInfo.AssetAAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAAmountDeltaBytes)
	packedAssetBAmountDeltaBytes, err := AmountToPackedAmountBytes(txInfo.AssetBAmountDelta)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetBAmountDeltaBytes)
	packedLpAmountBytes, err := AmountToPackedAmountBytes(txInfo.LpAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedLpAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToMintNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		txType 1byte
		fromAccountIndex 4byte
		toAccountIndex 4byte
		nftAssetId 4byte
		nftIndex 8byte
		nftContentHash 32byte
		creatorFeeRate 2byte
		gasFeeAccountIndex 4byte
		gasFeeAssetId 2byte
		gasFeeAssetAmount 2byte
	*/
	if oTx.TxType != commonTx.TxTypeMintNft {
		logx.Errorf("[ConvertTxToMintNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToMintNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseMintNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToMintNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.CreatorAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint64ToBytes(uint64(txInfo.NftIndex)))
	buf.Write(common.FromHex(txInfo.NftContentHash))
	buf.Write(Uint16ToBytes(uint16(txInfo.CreatorFeeRate)))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToTransferNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		txType 1byte
		fromAccountIndex 4byte
		toAccountIndex 4byte
		nftAssetId 4byte
		nftIndex 8byte
		nftContentHash 32byte
		gasFeeAccountIndex 4byte
		gasFeeAssetId 2byte
		gasFeeAssetAmount 2byte
	*/
	if oTx.TxType != commonTx.TxTypeTransferNft {
		logx.Errorf("[ConvertTxToMintNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToMintNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseTransferNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToMintNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.ToAccountIndex)))
	buf.Write(Uint64ToBytes(uint64(txInfo.NftIndex)))
	buf.Write(common.FromHex(txInfo.NftContentHash))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToSetNftPricePubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		txType 1byte
		accountIndex 4byte
		nftAssetId 4byte
		assetId 2byte
		assetAmount 5byte
		gasFeeAccountIndex 4byte
		gasFeeAssetId 2byte
		gasFeeAssetAmount 2byte
	*/
	if oTx.TxType != commonTx.TxTypeSetNftPrice {
		logx.Errorf("[ConvertTxToSetNftPricePubData] invalid tx type")
		return nil, errors.New("[ConvertTxToSetNftPricePubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseSetNftPriceTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToSetNftPricePubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	packedAssetAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToBuyNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		txType 1byte
		buyerAccountIndex 4byte
		ownerAccountIndex 4byte
		nftAssetId 4byte
		assetId 2byte
		assetAmount 5byte
		gasFeeAccountIndex 4byte
		gasFeeAssetId 2byte
		gasFeeAssetAmount 2byte
	*/
	if oTx.TxType != commonTx.TxTypeBuyNft {
		logx.Errorf("[ConvertTxToBuyNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToBuyNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseBuyNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToBuyNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(Uint32ToBytes(uint32(txInfo.OwnerAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	packedAssetAmountBytes, err := AmountToPackedAmountBytes(txInfo.AssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedAssetAmountBytes)
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToDepositNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		struct DepositNFT {
			uint8 txType;
			uint32 accountIndex;
			bytes32 accountNameHash;
			uint8 nftType;
			int64 nftIndex;
			address nftL1Address;
			uint256 nftTokenId;
			uint32 amount;
		}
	*/
	if oTx.TxType != commonTx.TxTypeDepositNft {
		logx.Errorf("[ConvertTxToDepositNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToDepositNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseDepositNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(txInfo.AccountIndex))
	buf.Write(common.FromHex(txInfo.AccountNameHash))
	buf.WriteByte(txInfo.NftType)
	buf.Write(Uint40ToBytes(int64(txInfo.NftIndex)))
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(Uint32ToBytes(txInfo.Amount))
	return buf.Bytes(), nil
}

func ConvertTxToWithdrawPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		struct Withdraw {
			uint8 txType;
			uint32 accountIndex;
			address toAddress;
			uint16 assetId;
			uint128 assetAmount;
			uint32 gasFeeAccountIndex;
			uint16 gasFeeAssetId;
			uint16 gasFeeAssetAmount;
		}
	*/
	if oTx.TxType != commonTx.TxTypeWithdraw {
		logx.Errorf("[ConvertTxToDepositNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToDepositNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseWithdrawTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.FromAccountIndex)))
	buf.Write(AddressStrToBytes(txInfo.ToAddress))
	buf.Write(Uint16ToBytes(uint16(txInfo.AssetId)))
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToWithdrawNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		// Withdraw Nft pubdata
		struct WithdrawNFT {
			uint8 txType;
			uint32 accountIndex;
			bytes32 accountNameHash;
			uint8 nftType;
			uint40 nftIndex;
			bytes32 nftContentHash;
			address nftL1Address;
			uint256 nftL1TokenId;
			uint32 amount;
			address toAddress;
			address proxyAddress;
			uint32 gasFeeAccountIndex;
			uint16 gasFeeAssetId;
			uint16 gasFeeAssetAmount;
		}
	*/
	if oTx.TxType != commonTx.TxTypeWithdrawNft {
		logx.Errorf("[ConvertTxToWithdrawNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToWithdrawNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseWithdrawNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToWithdrawNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common.FromHex(txInfo.AccountNameHash))
	buf.WriteByte(txInfo.NftType)
	buf.Write(Uint40ToBytes(txInfo.NftIndex))
	buf.Write(common.FromHex(txInfo.NftContentHash))
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(AddressStrToBytes(txInfo.ToAddress))
	buf.Write(AddressStrToBytes(txInfo.ProxyAddress))
	buf.Write(Uint32ToBytes(uint32(txInfo.GasAccountIndex)))
	buf.Write(Uint16ToBytes(uint16(txInfo.GasFeeAssetId)))
	packedFeeBytes, err := FeeToPackedFeeBytes(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ConvertTxToDepositPubData] unable to convert amount to packed fee amount: %s", err.Error())
		return nil, err
	}
	buf.Write(packedFeeBytes)
	return buf.Bytes(), nil
}

func ConvertTxToFullExitPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		// full exit pubdata
		struct FullExit {
			uint8 txType;
			uint32 accountIndex;
			bytes32 accountNameHash;
			uint16 assetId;
			uint128 assetAmount;
		}
	*/
	if oTx.TxType != commonTx.TxTypeFullExit {
		logx.Errorf("[ConvertTxToFullExitPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToFullExitPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseFullExitTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToFullExitPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(uint32(txInfo.AccountIndex)))
	buf.Write(common.FromHex(txInfo.AccountNameHash))
	buf.Write(Uint16ToBytes(txInfo.AssetId))
	buf.Write(Uint128ToBytes(txInfo.AssetAmount))
	return buf.Bytes(), nil
}

func ConvertTxToFullExitNftPubData(oTx *mempool.MempoolTx) (pubData []byte, err error) {
	/*
		// full exit nft pubdata
		struct FullExitNFT {
			uint8 txType;
			uint32 accountIndex;
			bytes32 accountNameHash;
			uint8 nftType;
			uint40 nftIndex;
			bytes32 nftContentHash;
			address nftL1Address;
			uint256 nftL1TokenId;
			uint32 amount;
			address toAddress;
			address proxyAddress;
		}
	*/
	if oTx.TxType != commonTx.TxTypeFullExitNft {
		logx.Errorf("[ConvertTxToFullExitNftPubData] invalid tx type")
		return nil, errors.New("[ConvertTxToFullExitNftPubData] invalid tx type")
	}
	// parse tx
	txInfo, err := commonTx.ParseFullExitNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConvertTxToFullExitNftPubData] unable to parse tx info: %s", err.Error())
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(uint8(oTx.TxType))
	buf.Write(Uint32ToBytes(txInfo.AccountIndex))
	buf.Write(common.FromHex(txInfo.AccountNameHash))
	buf.WriteByte(txInfo.NftType)
	buf.Write(Uint40ToBytes(int64(txInfo.NftIndex)))
	buf.Write(txInfo.NftContentHash)
	buf.Write(AddressStrToBytes(txInfo.NftL1Address))
	buf.Write(Uint256ToBytes(txInfo.NftL1TokenId))
	buf.Write(Uint32ToBytes(txInfo.Amount))
	buf.Write(AddressStrToBytes(txInfo.ToAddress))
	buf.Write(AddressStrToBytes(txInfo.ProxyAddress))
	return buf.Bytes(), nil
}

// TODO create block commitment
func CreateBlockCommitment(lastBlockHeight, currentBlockHeight int64, pubdata []byte) string {
	var buf bytes.Buffer
	buf.Write(Int64ToBytes(lastBlockHeight))
	buf.Write(Int64ToBytes(currentBlockHeight))
	buf.Write(pubdata)
	hFunc := mimc.NewMiMC()
	hFunc.Write(buf.Bytes())
	commitment := hFunc.Sum(nil)
	return common.Bytes2Hex(commitment)
}
