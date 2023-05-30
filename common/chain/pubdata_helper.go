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
 *
 */

package chain

import (
	types2 "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"

	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/types"
)

func ParseDepositPubData(pubData []byte) (tx *txtypes.DepositTxInfo, err error) {
	/*
		struct Deposit {
			uint8 txType;
			uint32 accountIndex;
			bytes20 L1Address;
			uint16 assetId;
			uint128 amount;
		}
	*/
	if len(pubData) != types.DepositPubDataSize {
		logx.Error("[ParseDepositPubData] invalid size")
		return nil, types.AppErrDepositPubDataInvalidSize
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	_, amount := common2.ReadUint128(pubData, offset)
	tx = &txtypes.DepositTxInfo{
		TxType:       txType,
		AccountIndex: int64(accountIndex),
		L1Address:    l1Address,
		AssetId:      int64(assetId),
		AssetAmount:  amount,
	}
	return tx, nil
}

func ParseDepositNftPubData(pubData []byte) (tx *txtypes.DepositNftTxInfo, err error) {
	if len(pubData) != types.DepositNftPubDataSize {
		logx.Error("[ParseDepositNftPubData] invalid size")
		return nil, types.AppErrDepositNFTPubDataInvalidSize
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, royaltyRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, nftContentHash := common2.ReadBytes32(pubData, offset)
	_, nftContentType := common2.ReadUint8(pubData, offset)

	tx = &txtypes.DepositNftTxInfo{
		TxType:              txType,
		AccountIndex:        int64(accountIndex),
		NftIndex:            nftIndex,
		CreatorAccountIndex: int64(creatorAccountIndex),
		RoyaltyRate:         int64(royaltyRate),
		NftContentHash:      nftContentHash,
		L1Address:           l1Address,
		CollectionId:        int64(collectionId),
		NftContentType:      int64(nftContentType),
	}
	return tx, nil
}

func ParseFullExitPubData(pubData []byte) (tx *txtypes.FullExitTxInfo, err error) {
	if len(pubData) != types.FullExitPubDataSize {
		logx.Error("[ParseFullExitPubData] invalid size")
		return nil, types.AppErrFullExitPubDataInvalidSize
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetAmount := common2.ReadUint128(pubData, offset)
	_, l1Address := common2.ReadAddress(pubData, offset)
	tx = &txtypes.FullExitTxInfo{
		TxType:       txType,
		AccountIndex: int64(accountIndex),
		L1Address:    l1Address,
		AssetId:      int64(assetId),
		AssetAmount:  assetAmount,
	}
	return tx, nil
}

func ParseFullExitNftPubData(pubData []byte) (tx *txtypes.FullExitNftTxInfo, err error) {
	if len(pubData) != types.FullExitNftPubDataSize {
		logx.Error("[ParseFullExitNftPubData] invalid size")
		return nil, types.AppErrFullExitNftPubDataInvalidSize
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, royaltyRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, creatorL1Address := common2.ReadAddress(pubData, offset)
	offset, nftContentHash := common2.ReadBytes32(pubData, offset)
	_, nftContentType := common2.ReadUint8(pubData, offset)

	tx = &txtypes.FullExitNftTxInfo{
		TxType:              txType,
		AccountIndex:        int64(accountIndex),
		CreatorAccountIndex: int64(creatorAccountIndex),
		RoyaltyRate:         int64(royaltyRate),
		NftIndex:            nftIndex,
		CollectionId:        int64(collectionId),
		L1Address:           l1Address,
		CreatorL1Address:    creatorL1Address,
		NftContentHash:      nftContentHash,
		NftContentType:      int64(nftContentType),
	}
	return tx, nil
}

func ParsePubDataForDesert(pubDataStr string) ([]txtypes.TxInfo, error) {
	pubData := common.FromHex(pubDataStr)
	txInfos := make([]txtypes.TxInfo, 0)
	sizePerTx := types2.PubDataBitsSizePerTx / 8

	for i := 0; i < len(pubData)/sizePerTx; i++ {
		subPubData := pubData[i*sizePerTx : (i+1)*sizePerTx]
		offset := 0
		offset, txType := common2.ReadUint8(subPubData, offset)
		switch txType {
		case types.TxTypeAtomicMatch:
			txInfo, err := ParseAtomicMatchPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeCancelOffer:
			txInfo, err := ParseCancelOfferPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeCreateCollection:
			txInfo, err := ParseCollectionPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeDeposit:
			txInfo, err := ParseDepositPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeDepositNft:
			txInfo, err := ParseDepositNftPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeFullExit:
			txInfo, err := ParseFullExitPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeFullExitNft:
			txInfo, err := ParseFullExitNftPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeMintNft:
			txInfo, err := ParseMintNftPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeChangePubKey:
			txInfo, err := ParseChangePubKeyPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeTransfer:
			txInfo, err := ParseTransferPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeTransferNft:
			txInfo, err := ParseTransferNftPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeWithdraw:
			txInfo, err := ParseWithdrawPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeWithdrawNft:
			txInfo, err := ParseWithdrawNftPubDataForDesert(subPubData)
			if err != nil {
				return nil, err
			}
			txInfos = append(txInfos, txInfo)
			break
		case types.TxTypeEmpty:
			break
		}
	}

	return txInfos, nil
}

func ParseAtomicMatchPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, buyOfferAccountIndex := common2.ReadUint32(pubData, offset)
	offset, buyOfferOfferId := common2.ReadUint24(pubData, offset)
	offset, sellOfferAccountIndex := common2.ReadUint32(pubData, offset)
	offset, sellOfferOfferId := common2.ReadUint24(pubData, offset)
	offset, buyOfferNftIndex := common2.ReadUint40(pubData, offset)
	offset, sellOfferAssetId := common2.ReadUint16(pubData, offset)
	offset, buyOfferAssetPackedAmount := common2.ReadUint40(pubData, offset)
	buyOfferAssetAmount, err := util.UnpackAmount(big.NewInt(buyOfferAssetPackedAmount))
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}

	offset, royaltyPackedAmount := common2.ReadUint40(pubData, offset)
	royaltyAmount, err := util.UnpackAmount(big.NewInt(royaltyPackedAmount))
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}

	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)

	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.UnpackFee(big.NewInt(int64(gasFeeAssetPackedAmount)))
	if err != nil {
		return nil, err
	}

	offset, buyProtocolPackedAmount := common2.ReadUint40(pubData, offset)
	buyProtocolAmount, err := util.UnpackAmount(big.NewInt(buyProtocolPackedAmount))
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	offset, buyChanelAccountIndex := common2.ReadUint32(pubData, offset)
	offset, buyChanelPackedAmount := common2.ReadUint40(pubData, offset)
	buyChanelAmount, err := util.UnpackAmount(big.NewInt(buyChanelPackedAmount))
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}
	offset, sellChannelAccountIndex := common2.ReadUint32(pubData, offset)
	offset, sellChannelPackedAmount := common2.ReadUint40(pubData, offset)
	sellChanelAmount, err := util.UnpackAmount(big.NewInt(sellChannelPackedAmount))
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return nil, err
	}

	txInfo = &txtypes.AtomicMatchTxInfo{
		AccountIndex: int64(accountIndex),
		BuyOffer: &txtypes.OfferTxInfo{
			AccountIndex:        int64(buyOfferAccountIndex),
			OfferId:             int64(buyOfferOfferId),
			NftIndex:            buyOfferNftIndex,
			AssetAmount:         buyOfferAssetAmount,
			ChannelAccountIndex: int64(buyChanelAccountIndex),
			ProtocolAmount:      buyProtocolAmount,
			AssetId:             int64(sellOfferAssetId),
		},
		SellOffer: &txtypes.OfferTxInfo{
			AccountIndex:        int64(sellOfferAccountIndex),
			OfferId:             int64(sellOfferOfferId),
			AssetId:             int64(sellOfferAssetId),
			ChannelAccountIndex: int64(sellChannelAccountIndex),
			NftIndex:            buyOfferNftIndex,
			AssetAmount:         buyOfferAssetAmount,
		},
		SellChannelAmount: sellChanelAmount,
		BuyChannelAmount:  buyChanelAmount,
		RoyaltyAmount:     royaltyAmount,
		GasAccountIndex:   types.GasAccount,
		GasFeeAssetAmount: gasFeeAssetAmount,
		GasFeeAssetId:     int64(gasFeeAssetId),
	}
	return txInfo, nil
}

func ParseCancelOfferPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, offerId := common2.ReadUint24(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, _ := util.UnpackFee(big.NewInt(int64(gasFeeAssetPackedAmount)))

	txInfo = &txtypes.CancelOfferTxInfo{
		AccountIndex:      int64(accountIndex),
		GasAccountIndex:   types.GasAccount,
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
		OfferId:           int64(offerId),
	}

	return txInfo, nil
}

func ParseCollectionPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, _ := util.UnpackFee(big.NewInt(int64(gasFeeAssetPackedAmount)))

	txInfo = &txtypes.CreateCollectionTxInfo{
		AccountIndex:      int64(accountIndex),
		GasAccountIndex:   types.GasAccount,
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
		CollectionId:      int64(collectionId),
	}
	return txInfo, nil
}

func ParseDepositPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetAmount := common2.ReadUint128(pubData, offset)

	txInfo = &txtypes.DepositTxInfo{
		AccountIndex: int64(accountIndex),
		AssetId:      int64(assetId),
		AssetAmount:  assetAmount,
		L1Address:    l1Address,
	}

	return txInfo, nil
}

func ParseDepositNftPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, royaltyRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, nftContentHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	offset, nftContentType := common2.ReadUint8(pubData, offset)

	txInfo = &txtypes.DepositNftTxInfo{
		AccountIndex:        int64(accountIndex),
		NftIndex:            nftIndex,
		CreatorAccountIndex: int64(creatorAccountIndex),
		CollectionId:        int64(collectionId),
		RoyaltyRate:         int64(royaltyRate),
		NftContentHash:      nftContentHash,
		L1Address:           l1Address,
		NftContentType:      int64(nftContentType),
	}

	return txInfo, nil
}

func ParseFullExitPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetAmount := common2.ReadUint128(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)

	txInfo = &txtypes.FullExitTxInfo{
		AccountIndex: int64(accountIndex),
		AssetId:      int64(assetId),
		AssetAmount:  assetAmount,
		L1Address:    l1Address,
	}

	return txInfo, nil
}

func ParseFullExitNftPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, royaltyRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, creatorL1Address := common2.ReadAddress(pubData, offset)
	offset, nftContentHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	offset, nftContentType := common2.ReadUint8(pubData, offset)

	txInfo = &txtypes.FullExitNftTxInfo{
		AccountIndex:        int64(accountIndex),
		CreatorAccountIndex: int64(creatorAccountIndex),
		RoyaltyRate:         int64(royaltyRate),
		NftIndex:            nftIndex,
		CollectionId:        int64(collectionId),
		L1Address:           l1Address,
		CreatorL1Address:    creatorL1Address,
		NftContentHash:      nftContentHash,
		NftContentType:      int64(nftContentType),
	}

	return txInfo, nil
}

func ParseMintNftPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAddress := common2.ReadAddress(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)

	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.UnpackFee(big.NewInt(int64(gasFeeAssetPackedAmount)))
	if err != nil {
		return nil, err
	}
	offset, royaltyRate := common2.ReadUint16(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, nftContentHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)

	txInfo = &txtypes.MintNftTxInfo{
		CreatorAccountIndex: int64(creatorAccountIndex),
		ToAccountIndex:      int64(toAccountIndex),
		ToL1Address:         toAddress,
		NftIndex:            nftIndex,
		GasAccountIndex:     types.GasAccount,
		GasFeeAssetId:       int64(gasFeeAssetId),
		GasFeeAssetAmount:   gasFeeAssetAmount,
		NftCollectionId:     int64(collectionId),
		NftContentHash:      common.Bytes2Hex(nftContentHash),
		RoyaltyRate:         int64(royaltyRate),
	}

	return txInfo, nil
}

func ParseChangePubKeyPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, pubKeyX := common2.ReadBytes32(pubData, offset)
	offset, pubKeyY := common2.ReadBytes32(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, nonce := common2.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, packedFee := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.UnpackFee(new(big.Int).SetInt64(int64(packedFee)))
	if err != nil {
		return nil, err
	}

	txInfo = &txtypes.ChangePubKeyInfo{
		AccountIndex:      int64(accountIndex),
		PubKeyX:           pubKeyX,
		PubKeyY:           pubKeyY,
		L1Address:         l1Address,
		Nonce:             int64(nonce),
		GasAccountIndex:   types.GasAccount,
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
	}

	return txInfo, nil
}

func ParseTransferPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, fromAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAddress := common2.ReadAddress(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, packedAmount := common2.ReadUint40(pubData, offset)
	assetAmount, err := util.UnpackAmount(big.NewInt(packedAmount))
	if err != nil {
		return nil, err
	}
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, packedFee := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.UnpackFee(big.NewInt(int64(packedFee)))
	if err != nil {
		return nil, err
	}

	txInfo = &txtypes.TransferTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToL1Address:       toAddress,
		ToAccountIndex:    int64(toAccountIndex),
		GasAccountIndex:   types.GasAccount,
		AssetId:           int64(assetId),
		AssetAmount:       assetAmount,
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
	}

	return txInfo, nil
}

func ParseTransferNftPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, fromAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAddress := common2.ReadAddress(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, packedFee := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.UnpackFee(big.NewInt(int64(packedFee)))
	if err != nil {
		return nil, err
	}
	offset, callDataHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	txInfo = &txtypes.TransferNftTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToAccountIndex:    int64(toAccountIndex),
		ToL1Address:       toAddress,
		NftIndex:          nftIndex,
		GasAccountIndex:   types.GasAccount,
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
		CallDataHash:      callDataHash,
	}

	return txInfo, nil
}

func ParseWithdrawPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, fromAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAddress := common2.ReadAddress(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetAmount := common2.ReadUint128(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.UnpackFee(big.NewInt(int64(gasFeeAssetPackedAmount)))
	if err != nil {
		return nil, err
	}
	txInfo = &txtypes.WithdrawTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToAddress:         toAddress,
		GasAccountIndex:   types.GasAccount,
		AssetId:           int64(assetId),
		AssetAmount:       assetAmount,
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
	}

	return txInfo, nil
}

func ParseWithdrawNftPubDataForDesert(pubData []byte) (txInfo txtypes.TxInfo, err error) {
	offset := 1
	offset, fromAccountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, royaltyRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.UnpackFee(big.NewInt(int64(gasFeeAssetPackedAmount)))
	if err != nil {
		return nil, err
	}
	offset, toAddress := common2.ReadAddress(pubData, offset)
	offset, creatorL1Address := common2.ReadAddress(pubData, offset)
	offset, nftContentHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	offset, nftContentType := common2.ReadUint8(pubData, offset)
	txInfo = &txtypes.WithdrawNftTxInfo{
		AccountIndex:        int64(fromAccountIndex),
		CreatorAccountIndex: int64(creatorAccountIndex),
		RoyaltyRate:         int64(royaltyRate),
		NftIndex:            nftIndex,
		ToAddress:           toAddress,
		CollectionId:        int64(collectionId),
		NftContentHash:      nftContentHash,
		CreatorL1Address:    creatorL1Address,
		GasAccountIndex:     types.GasAccount,
		GasFeeAssetId:       int64(gasFeeAssetId),
		GasFeeAssetAmount:   gasFeeAssetAmount,
		NftContentType:      int64(nftContentType),
	}

	return txInfo, nil
}
